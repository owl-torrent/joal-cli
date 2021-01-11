package web

import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/anthonyraymond/joal-cli/pkg/plugins"
	"github.com/go-stomp/stomp"
	stompServer "github.com/go-stomp/stomp/server"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Plugin struct {
	enabled        bool
	coreBridge     plugins.ICoreBridge
	stompServer    net.Listener
	stompPublisher *stomp.Conn
	coreListener   *appStateCoreListener
	httpServer     *http.Server
	wsListener     net.Listener
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

	httpHandler := http.NewServeMux()
	// Register web ui static files endpoint
	httpHandler.Handle(conf.Http.WebUiPath, http.StripPrefix(conf.Http.WebUiPath, http.FileServer(http.Dir(staticFilesDir(configFolder)))))
	// Register the websocket negotiation endpoint
	httpHandler.HandleFunc(normalizeStompUrlPath(conf.Stomp.UrlPath), wsListener.HttpNegotiationHandler(conf.WebSocket))

	w.httpServer = &http.Server{
		Handler:           httpHandler,
		ReadTimeout:       conf.Http.ReadTimeout,
		ReadHeaderTimeout: conf.Http.ReadHeaderTimeout,
		WriteTimeout:      conf.Http.WriteTimeout,
		IdleTimeout:       conf.Http.IdleTimeout,
		MaxHeaderBytes:    conf.Http.MaxHeaderBytes,
	}

	// Start Http server
	err = startHttpServer(w.httpServer, conf.Http, log)
	if err != nil {
		w.enabled = false
		shutdown(w, nil)
		return err
	}

	// Create a client connected to our stomp server to be able to dispatch messages
	negotiationEndpoint, err := url.Parse(fmt.Sprintf("ws://localhost:%d%s", conf.Http.Port, normalizeStompUrlPath(conf.Stomp.UrlPath)))
	if err != nil {
		w.enabled = false
		shutdown(w, nil)
		return errors.Wrap(err, "failed to create stomp negotiation endpoint URL")
	}
	stompPublisher, err := createStompPublisher(negotiationEndpoint, conf.WebSocket, conf.Stomp)
	if err != nil {
		w.enabled = false
		shutdown(w, nil)
		return errors.Wrap(err, "failed to create the stomp publisher")
	}

	w.enabled = true
	w.stompServer = wsListener
	w.stompPublisher = stompPublisher
	w.coreListener = &appStateCoreListener{
		state:          State{}.InitialState(),
		lock:           &sync.Mutex{},
		stompPublisher: stompPublisher,
	}

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
}

func normalizeStompUrlPath(path string) string {
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return "/ws"
	}
	if path[0] != '/' {
		return fmt.Sprintf("/%s", path)
	}
	return path
}

func startHttpServer(server *http.Server, config *HttpConfig, log *zap.Logger) error {
	// Create a listener for the HTTP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return errors.Wrapf(err, "failed to start listenet on port %d", config.Port)
	}
	// Starts the Http server
	go func() {
		if err := server.Serve(listener); err != nil {
			if err != http.ErrServerClosed {
				log.Error("http server has been closed", zap.Error(err))
			}
		}
	}()
	return nil
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

func createStompPublisher(negotiationEndpointUrl *url.URL, wsConfig *WebSocketConfig, stompConfig *StompConfig) (*stomp.Conn, error) {
	c, _, err := websocket.Dial(context.Background(), negotiationEndpointUrl.String(), &websocket.DialOptions{
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
		stomp.ConnOpt.Host(negotiationEndpointUrl.Host),
		stomp.ConnOpt.HeartBeat(stompConfig.HeartBeat, stompConfig.HeartBeat),
		stomp.ConnOpt.UseStomp,
	)
	if err != nil {
		_ = c.Close(websocket.StatusGoingAway, "closing websocket because STOMP connect have failed")
		return nil, errors.Wrap(err, "failed to start stomp publisher")
	}
	return conn, nil
}
