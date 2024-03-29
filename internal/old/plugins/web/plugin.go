package web

import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/old/core/broadcast"
	"github.com/anthonyraymond/joal-cli/internal/old/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/old/plugins/types"
	"github.com/go-stomp/stomp/v3"
	stompServer "github.com/go-stomp/stomp/v3/server"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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

type plugin struct {
	configLoader       *webConfigLoader
	staticFilesDir     string
	coreBridge         types.ICoreBridge
	stompServer        net.Listener
	stompPublisher     *stomp.Conn
	coreListener       *appStateCoreListener
	httpServer         *http.Server
	wsListener         net.Listener
	unregisterListener func()
}

func (w *plugin) Name() string {
	return "Web UI"
}

func ShouldEnablePlugin() bool {
	for _, arg := range os.Args {
		if arg == "--no-webui" {
			return false
		}
	}
	return true
}

func BootStrap(pluginsRootDir string, coreBridge types.ICoreBridge, client *http.Client) (types.IJoalPlugin, error) {
	configRoot := filepath.Join(pluginsRootDir, "web")

	p := &plugin{
		configLoader:       newWebConfigLoader(configRoot),
		staticFilesDir:     staticFilesDirFromRoot(configRoot),
		coreBridge:         coreBridge,
		stompServer:        nil,
		stompPublisher:     nil,
		coreListener:       nil,
		httpServer:         nil,
		wsListener:         nil,
		unregisterListener: nil,
	}

	log := logs.GetLogger().With(zap.String("plugin", p.Name()))

	err := bootstrap(configRoot, client, log)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap web plugin: %w", err)
	}
	return p, nil
}

func (w *plugin) Start() error {
	log := logs.GetLogger().With(zap.String("plugin", w.Name()))

	conf, err := w.configLoader.ReadConfig()
	if err != nil {
		return err
	}
	w.coreListener = &appStateCoreListener{
		state: state{}.initialState(),
		lock:  &sync.Mutex{},
	}

	// Create a listener for the websocket server
	//  The listener is a faked one, if a client comes to the http websocket negotiation endpoint
	//  he will be upgraded then be available to the wsListener#Accept() (just like a real net.Listener)
	wsListener, err := newWebSocketListener()
	if err != nil {
		shutdown(w, nil)
		return fmt.Errorf("failed to create web stomp listener: %w", err)
	}
	w.wsListener = wsListener

	// Start the stomp server, so it's ready to accept connection as soon as the HTTP server is up
	startStompServer(conf.Stomp, w.wsListener, log)

	router := mux.NewRouter()
	// Register web ui static files endpoint
	router.Handle(conf.Http.withSecretPathPrefix(conf.Http.WebUiUrl), webUiStaticFilesHandler(conf.Http.WebUiUrl, w.staticFilesDir)) // TODO: replace with a SPA handler from gorilla/mux documentation
	// Register HTTP API
	registerApiRoutes(router.PathPrefix(conf.Http.withSecretPathPrefix(conf.Http.HttpApiUrl)).Subrouter(), func() types.ICoreBridge { return w.coreBridge }, func() *state { return w.coreListener.state })
	// Register the websocket negotiation endpoint
	router.HandleFunc(conf.Http.withSecretPathPrefix(conf.Http.WsNegotiationEndpointUrl), wsListener.HttpNegotiationHandleFunc(conf.WebSocket))

	// TODO: move cors somewhere else (maybe in startHttpServer), and move params in config
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodConnect, http.MethodHead, http.MethodTrace},
		Debug:          false,
	}).Handler(router)
	// Start Http server
	w.httpServer, err = startHttpServer(handler, conf.Http, log)
	if err != nil {
		shutdown(w, nil)
		return err
	}

	// Create a client connected to our stomp server to be able to dispatch messages
	stompPublisher, err := createStompPublisher(conf.Http, conf.WebSocket, conf.Stomp)
	if err != nil {
		shutdown(w, nil)
		return fmt.Errorf("failed to create the stomp publisher: %w", err)
	}

	w.stompPublisher = stompPublisher
	w.coreListener.stompPublisher = stompPublisher
	w.unregisterListener = broadcast.RegisterListener(w.coreListener)

	return nil
}

func (w *plugin) AfterCoreLoaded(coreBridge types.ICoreBridge) {
	w.coreBridge = coreBridge
}

func (w *plugin) Shutdown(ctx context.Context) {
	log := logs.GetLogger().With(zap.String("plugin", w.Name()))

	log.Info("Shutting down plugin")
	shutdown(w, ctx)
}

// a lock free and nil safe version of Shutdown()
func shutdown(w *plugin, ctx context.Context) {
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

func startHttpServer(httpHandler http.Handler, config *httpConfig, log *zap.Logger) (*http.Server, error) {
	// Create a listener for the HTTP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to start listenet on port %d: %w", config.Port, err)
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

func startStompServer(config *stompConfig, wsListener net.Listener, log *zap.Logger) {
	go func() {
		err := (&stompServer.Server{
			Authenticator: config,
			HeartBeat:     config.HeartBeat,
			Log:           wrapZapLogger(log),
		}).Serve(wsListener)
		if err != nil {
			log.Error("stomp server has been closed", zap.Error(err))
		}
	}()
}

func createStompPublisher(httpConf *httpConfig, wsConfig *webSocketConfig, stompConfig *stompConfig) (*stomp.Conn, error) {
	negotiationEndpoint, err := url.Parse(fmt.Sprintf("ws://localhost:%d%s", httpConf.Port, httpConf.withSecretPathPrefix(httpConf.WsNegotiationEndpointUrl)))
	if err != nil {
		return nil, fmt.Errorf("failed to create stomp negotiation endpoint URL: %w", err)
	}

	c, _, err := websocket.Dial(context.Background(), negotiationEndpoint.String(), &websocket.DialOptions{
		Subprotocols:         wsConfig.AcceptedSubProtocols,
		CompressionMode:      0,
		CompressionThreshold: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to local websocket endpoint: %w", err)
	}
	c.SetReadLimit(int64(wsConfig.MaxReadLimit))
	conn, err := stomp.Connect(
		websocket.NetConn(context.Background(), c, websocket.MessageText),
		stomp.ConnOpt.Login(stompConfig.Login, stompConfig.Password),
		stomp.ConnOpt.Host(negotiationEndpoint.Host), // FIXME: this may need an adaptation
		stomp.ConnOpt.HeartBeat(stompConfig.HeartBeat, stompConfig.HeartBeat),
		stomp.ConnOpt.UseStomp,
	)
	if err != nil {
		_ = c.Close(websocket.StatusGoingAway, "closing websocket because STOMP connect have failed")
		return nil, fmt.Errorf("failed to start stomp publisher: %w", err)
	}
	return conn, nil
}

func webUiStaticFilesHandler(webUiUrl string, staticFilesPath string) http.Handler {
	return http.StripPrefix(webUiUrl, http.FileServer(http.Dir(staticFilesPath)))
}
