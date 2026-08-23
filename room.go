package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type WebRTCIceCandidate struct {
	Candidate        string  `json:"candidate"`
	SdpMid           *string `json:"sdpMid"`
	SdpMLineIndex    *uint16 `json:"sdpMLineIndex"`
	UsernameFragment *string `json:"usernameFragment"`
}

type SignalMessage struct {
	Type      string              `json:"type"`
	Candidate *WebRTCIceCandidate `json:"candidate,omitempty"`
	RoomCode    *string             `json:"roomCode,omitempty"`
	SDP       string              `json:"sdp,omitempty"`
}

func (s *signalingServer) createRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:8080"},
	})

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

		log.Printf("MSG: %+v", msg)
		if msg.Type == "create-room" {
			roomCode := generateId(5)
			err := wsjson.Write(ctx, conn, SignalMessage{
				Type:   "room-code",
				RoomCode: &roomCode,
			})

			if err != nil {
				log.Println("write error: ", err)
				break
			}

		} else if msg.Type == "ice-candidate" {
			log.Printf("Received ICE candidate: %+v", msg.Candidate)
		}
	}
}

func (s *signalingServer) joinRoom(w http.ResponseWriter, r *http.Request) {
}
