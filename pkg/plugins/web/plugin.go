package web

import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/plugins"
	"github.com/go-stomp/stomp"
	stompServer "github.com/go-stomp/stomp/server"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"sync"
	"time"
)

type Plugin struct {
	enabled        bool
	coreBridge     plugins.ICoreBridge
	stompServer    net.Listener
	stompPublisher *stomp.Conn
	coreListener   *appStateCoreListener
	// TODO: add HTTP
}

func (w *Plugin) SubFolder() string {
	return "web"
}

func (w *Plugin) Name() string {
	return "Web UI"
}

func (w *Plugin) Initialize(configFolder string) error {
	configLoader, err := NewWebConfigLoader(configFolder, &http.Client{})
	if err != nil {
		w.enabled = false
		return err
	}
	conf, err := configLoader.LoadConfigAndInitIfNeeded()
	if err != nil {
		w.enabled = false
		return err
	}

	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Stomp.Port))
	if err != nil {
		w.enabled = false
		return errors.Wrap(err, "failed to create web stomp listener")
	}

	sts := stompServer.Server{
		Addr:          tcpListener.Addr().String(),
		Authenticator: conf.Stomp,
		HeartBeat:     30 * time.Second,
	}

	err = sts.Serve(tcpListener)
	if err != nil {
		_ = tcpListener.Close()
		w.enabled = false
		return errors.Wrap(err, "failed to start web stomp server")
	}

	conn, err := stomp.Dial(
		"tcp",
		fmt.Sprintf("127.0.0.1:%d", conf.Stomp.Port),
		stomp.ConnOpt.Login("guest", "guest"),
		stomp.ConnOpt.Host("/"),
	)
	if err != nil {
		_ = tcpListener.Close()
		w.enabled = false
		return errors.Wrap(err, "failed to start stomp publisher")
	}

	w.enabled = true
	w.stompServer = tcpListener
	w.stompPublisher = conn
	w.coreListener = &appStateCoreListener{
		state:          &State{},
		lock:           &sync.Mutex{},
		stompPublisher: conn,
	}

	return nil
}

func (w *Plugin) AfterCoreLoaded(coreBridge plugins.ICoreBridge) {
	w.coreBridge = coreBridge
}

func (w *Plugin) Enabled() bool {
	return w.enabled
}

func (w *Plugin) Shutdown(_ context.Context) {
	_ = w.stompPublisher.Disconnect()
	_ = w.stompServer.Close()
}
