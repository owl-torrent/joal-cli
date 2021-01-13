package web

import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/core/broadcast"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/anthonyraymond/joal-cli/pkg/plugins"
	"github.com/go-stomp/stomp"
	stompServer "github.com/go-stomp/stomp/server"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Plugin struct {
	enabled            bool
	coreBridge         plugins.ICoreBridge
	stompServer        net.Listener
	stompPublisher     *stomp.Conn
	coreListener       *appStateCoreListener
	httpServer         *http.Server
	wsListener         net.Listener
	unregisterListener func()
}

func (w *Plugin) Name() string {
	return "Web UI"
}

func (w *Plugin) ShouldEnable() bool {
	for _, arg := range os.Args {
		if arg == "--no-webui" {
			return false
		}
	}
	return true
}

func (w *Plugin) Initialize(configFolder string) error {
	log := logs.GetLogger().With(zap.String("plugin", w.Name()))

	configFolder = filepath.Join(configFolder, "web")
	configLoader, err := NewWebConfigLoader(configFolder, &http.Client{}, log)
	if err != nil {
		w.enabled = false
		return err
	}
	conf, err := configLoader.LoadConfigAndInitIfNeeded()
	if err != nil {
		w.enabled = false
		return err
	}
	w.coreListener = &appStateCoreListener{
		state: State{}.InitialState(),
		lock:  &sync.Mutex{},
	}

	// Create a listener for the websocket server
	//  The listener is a faked one, if a client comes to the http websocket negotiation endpoint
	//  he will be upgraded then be available to the wsListener#Accept() (just like a real net.Listener)
	wsListener, err := NewWebSocketListener()
	if err != nil {
		w.enabled = false
		shutdown(w, nil)
		return errors.Wrap(err, "failed to create web stomp listener")
	}
	w.wsListener = wsListener

	// Start the stomp server, so it's ready to accept connection as soon as the HTTP server is up
	startStompServer(conf.Stomp, w.wsListener, log)

	router := mux.NewRouter()
	// Register web ui static files endpoint
	router.Handle(conf.Http.WebUiUrl, webUiStaticFilesHandler(conf.Http.WebUiUrl, staticFilesDir(configFolder))) // TODO: replace with a SPA handler from mux documentation
	// Register HTTP API
	registerApiRoutes(router.PathPrefix(conf.Http.HttpApiUrl).Subrouter(), func() plugins.ICoreBridge { return w.coreBridge }, func() *State { return w.coreListener.state })
	// Register the websocket negotiation endpoint
	router.HandleFunc(conf.Http.WsNegotiationEndpointUrl, wsListener.HttpNegotiationHandleFunc(conf.WebSocket))

	// Start Http server
	w.httpServer, err = startHttpServer(router, conf.Http, log)
	if err != nil {
		w.enabled = false
		shutdown(w, nil)
		return err
	}

	// Create a client connected to our stomp server to be able to dispatch messages
	stompPublisher, err := createStompPublisher(conf.Http, conf.WebSocket, conf.Stomp)
	if err != nil {
		w.enabled = false
		shutdown(w, nil)
		return errors.Wrap(err, "failed to create the stomp publisher")
	}

	w.enabled = true
	w.stompPublisher = stompPublisher
	w.coreListener.stompPublisher = stompPublisher
	w.unregisterListener = broadcast.RegisterListener(w.coreListener)

	return nil
}

func (w *Plugin) AfterCoreLoaded(coreBridge plugins.ICoreBridge) {
	w.coreBridge = coreBridge
}

func (w *Plugin) Shutdown(ctx context.Context) {
	if !w.enabled {
		return
	}
	log := logs.GetLogger().With(zap.String("plugin", w.Name()))

	log.Info("Shutting down plugin")
	shutdown(w, ctx)
}

// a lock free and nil safe version of Shutdown()
func shutdown(w *Plugin, ctx context.Context) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	if w.wsListener != nil {
		_ = w.wsListener.Close()
	}
	if w.httpServer != nil {
		if err := w.httpServer.Shutdown(ctx); err != nil {
			_ = w.httpServer.Close()
		}
	}
	if w.stompPublisher != nil {
		_ = w.stompPublisher.Disconnect()
	}
	if w.stompServer != nil {
		_ = w.stompServer.Close()
	}
	if w.unregisterListener != nil {
		w.unregisterListener()
	}
}

func startHttpServer(httpHandler http.Handler, config *HttpConfig, log *zap.Logger) (*http.Server, error) {
	// Create a listener for the HTTP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to start listenet on port %d", config.Port)
	}
	server := &http.Server{
		Handler:           httpHandler,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}

	// Starts the Http server
	go func() {
		if err := server.Serve(listener); err != nil {
			if err != http.ErrServerClosed {
				log.Error("http server has been closed", zap.Error(err))
			}
		}
	}()
	return server, nil
}

func startStompServer(config *StompConfig, wsListener net.Listener, log *zap.Logger) {
	go func() {
		err := (&stompServer.Server{
			Authenticator: config,
			HeartBeat:     config.HeartBeat,
		}).Serve(wsListener)
		if err != nil {
			log.Error("stomp server has been closed", zap.Error(err))
		}
	}()
}

func createStompPublisher(httpConf *HttpConfig, wsConfig *WebSocketConfig, stompConfig *StompConfig) (*stomp.Conn, error) {
	negotiationEndpoint, err := url.Parse(fmt.Sprintf("ws://localhost:%d%s", httpConf.Port, httpConf.WsNegotiationEndpointUrl))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stomp negotiation endpoint URL")
	}

	c, _, err := websocket.Dial(context.Background(), negotiationEndpoint.String(), &websocket.DialOptions{
		Subprotocols:         wsConfig.AcceptedSubProtocols,
		CompressionMode:      0,
		CompressionThreshold: 0,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to local websocket endpoint")
	}
	c.SetReadLimit(int64(wsConfig.MaxReadLimit))
	conn, err := stomp.Connect(
		websocket.NetConn(context.Background(), c, websocket.MessageText),
		stomp.ConnOpt.Login(stompConfig.Login, stompConfig.Password),
		stomp.ConnOpt.Host(negotiationEndpoint.Host), // TODO: this may need an adaptation
		stomp.ConnOpt.HeartBeat(stompConfig.HeartBeat, stompConfig.HeartBeat),
		stomp.ConnOpt.UseStomp,
	)
	if err != nil {
		_ = c.Close(websocket.StatusGoingAway, "closing websocket because STOMP connect have failed")
		return nil, errors.Wrap(err, "failed to start stomp publisher")
	}
	return conn, nil
}

func webUiStaticFilesHandler(webUiUrl string, staticFilesPath string) http.Handler {
	return http.StripPrefix(webUiUrl, http.FileServer(http.Dir(staticFilesPath)))
}
