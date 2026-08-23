package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type SignalMessage struct {
	Type      string      `json:"type"`
	Candidate interface{} `json:"candidate,omitempty"`
}

func (s *signalingServer) createRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	defer conn.CloseNow()

	ctx := r.Context()
	for {
		var msg SignalMessage
		err := wsjson.Read(ctx, conn, &msg)

		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				break
			}
			log.Println("read error:", err)
			break
		}

		if msg.Type == "ice-candidate" {
			log.Printf("Received ICE candidate: %+v", msg.Candidate)
		}
	}
}

func (s *signalingServer) joinRoom(w http.ResponseWriter, r *http.Request) {
}
