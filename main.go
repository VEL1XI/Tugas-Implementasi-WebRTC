package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

// Room menyimpan koneksi peer dalam satu room
type Room struct {
	clients map[*websocket.Conn]string
	mu      sync.RWMutex
}

// RoomManager mengelola semua room
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

var roomManager = &RoomManager{
	rooms: make(map[string]*Room),
}

// Message struktur untuk signaling
type Message struct {
	Type     string                     `json:"type"`
	RoomID   string                     `json:"roomId,omitempty"`
	UserID   string                     `json:"userId,omitempty"`
	TargetID string                     `json:"targetId,omitempty"`
	SDP      *webrtc.SessionDescription `json:"sdp,omitempty"`
	ICE      *webrtc.ICECandidateInit   `json:"ice,omitempty"`
	Users    []string                   `json:"users,omitempty"`
}

func (rm *RoomManager) getOrCreateRoom(roomID string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		return room
	}

	room := &Room{
		clients: make(map[*websocket.Conn]string),
	}
	rm.rooms[roomID] = room
	return room
}

func (room *Room) addClient(conn *websocket.Conn, userID string) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.clients[conn] = userID
}

func (room *Room) removeClient(conn *websocket.Conn) {
	room.mu.Lock()
	defer room.mu.Unlock()
	delete(room.clients, conn)
}

func (room *Room) getUsers() []string {
	room.mu.RLock()
	defer room.mu.RUnlock()

	users := make([]string, 0, len(room.clients))
	for _, userID := range room.clients {
		users = append(users, userID)
	}
	return users
}

func (room *Room) broadcast(msg Message, excludeConn *websocket.Conn) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	for conn := range room.clients {
		if conn != excludeConn {
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Error broadcasting: %v", err)
			}
		}
	}
}

func (room *Room) sendToUser(targetID string, msg Message) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	for conn, userID := range room.clients {
		if userID == targetID {
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Error sending to user: %v", err)
			}
			return
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	var currentRoom *Room
	var currentUserID string
	var currentRoomID string

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		switch msg.Type {
		case "join":
			// User bergabung ke room
			currentRoomID = msg.RoomID
			currentUserID = msg.UserID
			currentRoom = roomManager.getOrCreateRoom(currentRoomID)

			// Kirim daftar user yang sudah ada ke user baru
			existingUsers := currentRoom.getUsers()
			conn.WriteJSON(Message{
				Type:  "users",
				Users: existingUsers,
			})

			// Tambah user baru ke room
			currentRoom.addClient(conn, currentUserID)

			// Broadcast ke semua user bahwa ada user baru
			currentRoom.broadcast(Message{
				Type:   "user-joined",
				UserID: currentUserID,
			}, conn)

			log.Printf("User %s joined room %s", currentUserID, currentRoomID)

		case "offer":
			// Forward SDP offer ke target user
			if currentRoom != nil {
				currentRoom.sendToUser(msg.TargetID, Message{
					Type:   "offer",
					UserID: currentUserID,
					SDP:    msg.SDP,
				})
			}

		case "answer":
			// Forward SDP answer ke target user
			if currentRoom != nil {
				currentRoom.sendToUser(msg.TargetID, Message{
					Type:   "answer",
					UserID: currentUserID,
					SDP:    msg.SDP,
				})
			}

		case "ice-candidate":
			// Forward ICE candidate ke target user
			if currentRoom != nil {
				currentRoom.sendToUser(msg.TargetID, Message{
					Type:   "ice-candidate",
					UserID: currentUserID,
					ICE:    msg.ICE,
				})
			}

		case "leave":
			if currentRoom != nil {
				currentRoom.removeClient(conn)
				currentRoom.broadcast(Message{
					Type:   "user-left",
					UserID: currentUserID,
				}, nil)
				log.Printf("User %s left room %s", currentUserID, currentRoomID)
			}
		}
	}

	// Cleanup saat koneksi terputus
	if currentRoom != nil {
		currentRoom.removeClient(conn)
		currentRoom.broadcast(Message{
			Type:   "user-left",
			UserID: currentUserID,
		}, nil)
		log.Printf("User %s disconnected from room %s", currentUserID, currentRoomID)
	}
}

// Handler untuk demonstrasi Pion WebRTC data channel
func handlePionDemo(w http.ResponseWriter, r *http.Request) {
	// Buat konfigurasi WebRTC
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Buat peer connection baru
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Event handler untuk ICE connection state
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State changed: %s", state.String())
	})

	// Event handler untuk data channel
	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Printf("Data Channel '%s' opened", dc.Label())

		dc.OnOpen(func() {
			log.Printf("Data channel '%s' open", dc.Label())
			dc.SendText("Hello from Pion WebRTC server!")
		})

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from DataChannel '%s': '%s'", dc.Label(), string(msg.Data))
			// Echo back
			dc.SendText(fmt.Sprintf("Server received: %s", string(msg.Data)))
		})
	})

	// Parse SDP offer dari client
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set remote description
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set local description
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Tunggu ICE gathering selesai
	<-webrtc.GatheringCompletePromise(peerConnection)

	// Kirim answer ke client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peerConnection.LocalDescription())
}

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// WebSocket endpoint untuk signaling
	http.HandleFunc("/ws", handleWebSocket)

	// Endpoint untuk demo Pion data channel
	http.HandleFunc("/pion-demo", handlePionDemo)

	port := ":8080"
	log.Printf("Server berjalan di http://localhost%s", port)
	log.Printf("Buka browser dan akses http://localhost%s untuk video call", port)
	log.Printf("Akses http://localhost%s/datachannel.html untuk demo data channel", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
