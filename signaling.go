package main

import "net/http"

type signalingServer struct {
	serveMux http.ServeMux
	rooms    map[string]*Room
}

type Room struct {
	host *WebRTCIceCandidate
	peer *WebRTCIceCandidate
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
