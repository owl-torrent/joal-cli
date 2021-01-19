package web

import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"go.uber.org/zap"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

// This is a wrapper around the github.com/gorilla/websocket library
//  the library but it does not implement the the go net.Listener, which is required by go-stomp.
//  This wrapper is simply here to encapsulate the lib in a struct implementing net.Listener.
type websocketListener struct {
	connChan chan net.Conn
	closed   bool
	lock     *sync.RWMutex
}

func newWebSocketListener() (*websocketListener, error) {
	l := &websocketListener{
		connChan: make(chan net.Conn),
		closed:   false,
		lock:     &sync.RWMutex{},
	}

	return l, nil
}

func (w *websocketListener) HttpNegotiationHandleFunc(conf *webSocketConfig) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		w.lock.RLock()
		if w.closed {
			w.lock.RUnlock()
			http.Error(writer, "websocket listener is closed, negotiation endpoint not longer accept connections", 418)
			return
		}
		w.lock.RUnlock()

		log := logs.GetLogger()
		client, err := websocket.Accept(writer, request, &websocket.AcceptOptions{
			Subprotocols:         conf.AcceptedSubProtocols,
			InsecureSkipVerify:   conf.InsecureSkipVerify,
			OriginPatterns:       conf.OriginPatterns,
			CompressionMode:      0,
			CompressionThreshold: 0,
		})
		if err != nil {
			log.Error("failed to upgrade to websocket:", zap.Error(err))
			return
		}
		client.SetReadLimit(int64(conf.MaxReadLimit))

		select {
		case w.connChan <- websocket.NetConn(context.Background(), client, websocket.MessageText):
			return
		case <-time.After(5 * time.Second):
			log.Warn("Websocket connection upgraded successfully but the stomp server has not claimed it before timeout")
		}
	}
}

func (w *websocketListener) Accept() (net.Conn, error) {
	w.lock.RLock()
	if w.closed {
		w.lock.RUnlock()
		return nil, fmt.Errorf("listener is closed")
	}
	w.lock.RUnlock()
	conn, ok := <-w.connChan
	if !ok {
		// chan was closed
		return nil, fmt.Errorf("listener is closed")
	}

	return conn, nil
}

func (w *websocketListener) Close() error {
	w.lock.Lock()
	if w.closed {
		w.lock.Unlock()
		return fmt.Errorf("already closed")
	}
	w.closed = true
	close(w.connChan)
	w.lock.Unlock()
	return nil
}

func (w *websocketListener) Addr() net.Addr {
	return websocketAddr{}
}

type websocketAddr struct {
}

func (a websocketAddr) Network() string {
	return "websocket"
}

func (a websocketAddr) String() string {
	return "websocket/unknown-addr"
}
