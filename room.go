package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	Message   string              `json:"message,omitempty"`
}

type TurnCredentails struct {
	Urls       []string `json:"url"`
	Username   string   `json:"username"`
	Credential string   `json:"credential"`
}

func (s *signalingServer) getTurnCredentails(w http.ResponseWriter, _ *http.Request) {

	urlEnv := os.Getenv("URL")
	username := os.Getenv("USERNAME")
	credential := os.Getenv("CREDENTIAL")

	if urlEnv == "" || username == "" || credential == "" {
		fmt.Println("Env not set")
		http.Error(w, "env is missing required variables", http.StatusInternalServerError)
		return
	}

	hosts := strings.Split(urlEnv, ",")

	for i := range hosts {
		hosts[i] = strings.TrimSpace(hosts[i])
	}

	w.Header().Set("Access-Control-Allow-Origin", "chetactoee.vercel.app")
	w.Header().Set("Content-Type", "application/json")

	creds := TurnCredentails{
		Urls:       hosts,
		Username:   username,
		Credential: credential,
	}

	json.NewEncoder(w).Encode(creds)
}

func (s *signalingServer) handleSignalMessage(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:8080", "localhost:8090", "chetactoee.vercel.app"},
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
		case "join-room":
			s.joinRoom(ctx, conn, msg)
		case "ice-candidate", "offer", "answer":
			s.relay(ctx, conn, msg)

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

	s.rooms[roomCode] = &Room{
		hostConn: conn,
	}
}

func (s *signalingServer) joinRoom(ctx context.Context, conn *websocket.Conn, msg SignalMessage) {
	roomCode := msg.RoomCode

	room, err := s.checkRoom(roomCode)

	if err != nil {
		err := wsjson.Write(ctx, conn, SignalMessage{
			Type:    "error",
			Message: err.Error(),
		})
		if err != nil {
			log.Println("Error writing missing room code error: ", err)
		}
		return
	}

	s.rooms[*roomCode].peerConn = conn

	wsjson.Write(ctx, room.peerConn, SignalMessage{
		Type: "wait",
	})

	wsjson.Write(ctx, room.hostConn, SignalMessage{
		Type: "peer-joined",
	})
}

func (s *signalingServer) relay(ctx context.Context, conn *websocket.Conn, msg SignalMessage) {
	room, err := s.checkRoom(msg.RoomCode)

	if err != nil {
		err := wsjson.Write(ctx, conn, SignalMessage{
			Type:    "error",
			Message: err.Error(),
		})
		if err != nil {
			log.Println("Error writing missing room code error: ", err)
		}
		return
	}

	var target *websocket.Conn

	switch conn {
	case room.hostConn:
		target = room.peerConn
	case room.peerConn:
		target = room.hostConn
	default:
		err := wsjson.Write(ctx, conn, SignalMessage{
			Type:    "error",
			Message: "Connection not recognized",
		})
		log.Println(err)
		log.Println("relay: connection not recognized as host or peer for this room")
		return
	}

	if target == nil {
		log.Println("relay target not connected yet, dropping message:", msg.Type)
		return
	}

	if err := wsjson.Write(ctx, target, msg); err != nil {
		log.Println("relay write error:", err)
	}
}

func (s *signalingServer) checkRoom(roomCode *string) (*Room, error) {
	log.Println(*roomCode)

	if roomCode == nil {
		return nil, errors.New("Room code is required")
	}

	room, ok := s.rooms[*roomCode]

	if !ok {
		return nil, errors.New("Room not found")
	}

	return room, nil
}
