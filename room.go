package main

import (
	"context"
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
	RoomCode  *string             `json:"roomCode,omitempty"`
	SDP       string              `json:"sdp,omitempty"`
}

func (s *signalingServer) handleSignalMessage(w http.ResponseWriter, r *http.Request) {
	s.rooms = make(map[string]*Room)
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
		switch msg.Type {
		case "create-room":
			s.createRoom(ctx, conn)
		case "host-candidate":
			roomCode := msg.RoomCode
			s.rooms[*roomCode].host = msg.Candidate
			log.Printf("Room: %+v", s.rooms[*roomCode])
		case "offer":
			roomCode := msg.RoomCode
			s.rooms[*roomCode].offer = msg.SDP
			log.Printf("Room: %+v", s.rooms[*roomCode])
		}
	}
}

func (s *signalingServer) createRoom(ctx context.Context, conn *websocket.Conn) {
	roomCode := generateId(5)
	err := wsjson.Write(ctx, conn, SignalMessage{
		Type:     "room-code",
		RoomCode: &roomCode,
	})

	if err != nil {
		log.Println("write error: ", err)
		return
	}

	s.rooms[roomCode] = &Room{}
}

func (s *signalingServer) joinRoom(w http.ResponseWriter, r *http.Request) {
}
