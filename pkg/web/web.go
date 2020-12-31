package web

import (
	"fmt"
	"github.com/go-stomp/stomp"
	stompServer "github.com/go-stomp/stomp/server"
	"github.com/pkg/errors"
	"net"
	"sync"
	"time"
)

type WebPlugin struct {
	coreBridge     ICoreBridge
	stompServer    net.Listener
	stompPublisher *stomp.Conn
	coreListener   *appStateCoreListener
	// TODO: add HTTP
}

func StartWebPlugin(coreBridge ICoreBridge, config *StompConfig) (*WebPlugin, error) {
	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create web stomp listener")
	}

	sts := stompServer.Server{
		Addr:          tcpListener.Addr().String(),
		Authenticator: config,
		HeartBeat:     30 * time.Second,
	}

	err = sts.Serve(tcpListener)
	if err != nil {
		_ = tcpListener.Close()
		return nil, errors.Wrap(err, "failed to start web stomp server")
	}

	conn, err := stomp.Dial(
		"tcp",
		fmt.Sprintf("127.0.0.1:%d", config.Port),
		stomp.ConnOpt.Login("guest", "guest"),
		stomp.ConnOpt.Host("/"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start stomp publisher")
	}

	w := &WebPlugin{
		coreBridge:     coreBridge,
		stompServer:    tcpListener,
		stompPublisher: conn,
		coreListener: &appStateCoreListener{
			state:          &State{},
			lock:           &sync.Mutex{},
			stompPublisher: conn,
		},
	}

	return w, nil
}

func (w *WebPlugin) Shutdown() {
	_ = w.stompPublisher.Disconnect()
	_ = w.stompServer.Close()
}
