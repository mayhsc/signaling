package main

import "net/http"

type signalingServer struct {
	serveMux http.ServeMux
}

func newSignalingServer() *signalingServer {
	s := &signalingServer{}

	s.serveMux.HandleFunc("/", s.createRoom)
	s.serveMux.HandleFunc("/join", s.joinRoom)

	return s
}

func (c *signalingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.serveMux.ServeHTTP(w, r)
}
