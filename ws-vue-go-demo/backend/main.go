package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// In production you should check origin and restrict allowed origins.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Add(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *Hub) Remove(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("write error, removing client:", err)
			c.Close()
			delete(h.clients, c)
		}
	}
}

func wsHandler(h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}
		defer conn.Close()

		h.Add(conn)
		defer h.Remove(conn)

		// Send welcome
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Chào từ server WebSocket!")); err != nil {
			log.Println("welcome write:", err)
			return
		}

		// Read loop
		for {
			typ, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if typ != websocket.TextMessage {
				continue
			}
			log.Printf("Received: %s\n", string(msg))
			// echo/broadcast to everyone
			h.Broadcast([]byte(fmt.Sprintf("Echo: %s", string(msg))))
		}
	}
}

func main() {
	// dist folder expected at ../frontend/dist (relative to backend dir)
	distDir := filepath.Join(".", "..", "frontend", "dist")
	// If dist not found, can try ./dist
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		distDir = "./dist" // fallback (if you moved dist to backend folder)
	}

	// File server for SPA
	fs := http.FileServer(http.Dir(distDir))
	http.Handle("/", fs)

	hub := NewHub()
	http.HandleFunc("/ws", wsHandler(hub))

	addr := "0.0.0.0:8080" // listen on all interfaces for Tailscale + local
	log.Printf("Server listening on %s, serving static from %s\n", addr, distDir)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
