package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket configuration.
const (
	MaxMessageSize = 512 * 1024 // 512KB
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	WriteWait      = 10 * time.Second
)

// Upgrader configures the WebSocket upgrader.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Client represents a single WebSocket connection.
type Client struct {
	ID     string
	UserID string
	Rooms  map[string]bool
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
	mu     sync.Mutex
}

// Message represents a WebSocket message.
type Message struct {
	Type    string      `json:"type"`
	Room    string      `json:"room,omitempty"`
	Payload interface{} `json:"payload"`
}

// Hub manages WebSocket clients and room-based messaging.
type Hub struct {
	mu sync.RWMutex
	// clients maps client ID -> *Client
	clients map[string]*Client
	// rooms maps room name -> set of client IDs
	rooms map[string]map[string]bool
	// userID -> set of client IDs
	userClients map[string]map[string]bool
	// register/unregister channels
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	// done signals the Run loop to stop
	done chan struct{}
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		rooms:       make(map[string]map[string]bool),
		userClients: make(map[string]map[string]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		done:       make(chan struct{}),
	}
}

// Stop gracefully stops the hub's main loop.
func (h *Hub) Stop() {
	close(h.done)
}

// Run starts the hub's main loop for managing clients.
func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			log.Printf("WebSocket hub shutting down...")
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			if _, ok := h.userClients[client.UserID]; !ok {
				h.userClients[client.UserID] = make(map[string]bool)
			}
			h.userClients[client.UserID][client.ID] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered: %s (user: %s)", client.ID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				// Remove from all rooms
				for roomName := range client.Rooms {
					if room, exists := h.rooms[roomName]; exists {
						delete(room, client.ID)
						if len(room) == 0 {
							delete(h.rooms, roomName)
						}
					}
				}
				// Remove from user clients
				if userRooms, ok := h.userClients[client.UserID]; ok {
					delete(userRooms, client.ID)
					if len(userRooms) == 0 {
						delete(h.userClients, client.UserID)
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()
			client.Conn.Close()
			log.Printf("WebSocket client unregistered: %s", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			if message.Room != "" {
				// Broadcast to room
				if room, ok := h.rooms[message.Room]; ok {
					for clientID := range room {
						if client, exists := h.clients[clientID]; exists {
							select {
							case client.Send <- mustMarshal(message):
							default:
								// Client send buffer is full, unregister
								delete(h.clients, clientID)
								close(client.Send)
							}
						}
					}
				}
			} else {
				// Broadcast to all clients
				for _, client := range h.clients {
					select {
					case client.Send <- mustMarshal(message):
					default:
						delete(h.clients, client.ID)
						close(client.Send)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// JoinRoom adds a client to a room.
func (h *Hub) JoinRoom(clientID, roomName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.clients[clientID]
	if !ok {
		return
	}

	if _, ok := h.rooms[roomName]; !ok {
		h.rooms[roomName] = make(map[string]bool)
	}
	h.rooms[roomName][clientID] = true
	client.Rooms[roomName] = true

	log.Printf("Client %s joined room: %s", clientID, roomName)
}

// LeaveRoom removes a client from a room.
func (h *Hub) LeaveRoom(clientID, roomName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.clients[clientID]; ok {
		delete(client.Rooms, roomName)
	}
	if room, ok := h.rooms[roomName]; ok {
		delete(room, clientID)
		if len(room) == 0 {
			delete(h.rooms, roomName)
		}
	}
}

// SendToUser sends a message to all connections of a specific user.
func (h *Hub) SendToUser(userID string, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clientIDs, ok := h.userClients[userID]
	if !ok {
		return
	}

	data := mustMarshal(message)
	for clientID := range clientIDs {
		if client, exists := h.clients[clientID]; exists {
			select {
			case client.Send <- data:
			default:
				log.Printf("Client %s send buffer full, dropping message", clientID)
			}
		}
	}
}

// SendToRoom sends a message to all clients in a room.
func (h *Hub) SendToRoom(roomName string, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[roomName]
	if !ok {
		return
	}

	data := mustMarshal(message)
	for clientID := range room {
		if client, exists := h.clients[clientID]; exists {
			select {
			case client.Send <- data:
			default:
				log.Printf("Client %s send buffer full, dropping message", clientID)
			}
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- message
}

// HandleWebSocket upgrades HTTP to WebSocket and registers the client.
func (h *Hub) HandleWebSocket(ginContext *gin.Context, userID string) {
	conn, err := Upgrader.Upgrade(ginContext.Writer, ginContext.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	clientID := generateClientID()
	client := &Client{
		ID:     clientID,
		UserID: userID,
		Rooms:  make(map[string]bool),
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h,
	}

	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
	}()

	c.Conn.SetReadLimit(MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (room subscription, etc.)
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid WebSocket message: %v", err)
			continue
		}

		switch msg.Type {
		case "subscribe":
			if room, ok := msg.Payload.(string); ok {
				c.Hub.JoinRoom(c.ID, room)
			}
		case "unsubscribe":
			if room, ok := msg.Payload.(string); ok {
				c.Hub.LeaveRoom(c.ID, room)
			}
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Stats returns hub statistics.
func (h *Hub) Stats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return map[string]interface{}{
		"total_clients": len(h.clients),
		"total_rooms":   len(h.rooms),
		"total_users":   len(h.userClients),
	}
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return []byte{}
	}
	return data
}

func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
