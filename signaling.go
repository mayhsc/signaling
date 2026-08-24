package main

import (
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

type signalingServer struct {
	serveMux http.ServeMux
	rooms    map[string]*Room
}

type Room struct {
	mu sync.Mutex
	hostConn *websocket.Conn
	peerConn *websocket.Conn
	offer string
	answer string
}

func newSignalingServer() *signalingServer {
	s := &signalingServer{}

	s.serveMux.HandleFunc("/", s.handleSignalMessage)

	return s
}

func (c *signalingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.serveMux.ServeHTTP(w, r)
}
