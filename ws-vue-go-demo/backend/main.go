package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	ID      int64  `json:"id"`
	Room    string `json:"room"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Ts      int64  `json:"ts"` // unix ms
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn     *websocket.Conn
	username string
	roomID   string
}

var rooms = make(map[string]map[*Client]bool) // roomID -> set of clients
var roomsMu sync.Mutex
var db *sql.DB

func ensureDB(path string) (*sql.DB, error) {
	needInit := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		needInit = true
	}
	d, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// Set WAL mode for concurrency
	_, _ = d.Exec("PRAGMA journal_mode = WAL;")
	if needInit {
		schema := `
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room TEXT NOT NULL,
    sender TEXT NOT NULL,
    content TEXT NOT NULL,
    ts INTEGER NOT NULL
);
CREATE INDEX idx_room_ts ON messages(room, ts DESC);
`
		_, err = d.Exec(schema)
		if err != nil {
			return nil, err
		}
	}
	return d, nil
}

func saveMessage(m *Message) error {
	res, err := db.Exec("INSERT INTO messages(room,sender,content,ts) VALUES(?,?,?,?)",
		m.Room, m.Sender, m.Content, m.Ts)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	m.ID = id
	return nil
}

func loadHistory(room string, page, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit
	rows, err := db.Query("SELECT id,room,sender,content,ts FROM messages WHERE room = ? ORDER BY ts ASC LIMIT ? OFFSET ?", room, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Room, &m.Sender, &m.Content, &m.Ts); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func broadcastToRoom(room string, payload interface{}) {
	roomsMu.Lock()
	clients := rooms[room]
	roomsMu.Unlock()
	if clients == nil {
		return
	}
	b, _ := json.Marshal(payload)
	for c := range clients {
		if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Println("write to client err:", err)
			// close and remove
			c.conn.Close()
			roomsMu.Lock()
			delete(rooms[room], c)
			roomsMu.Unlock()
		}
	}
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	// Query params: ?username=...&room=...
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	roomID := strings.TrimSpace(r.URL.Query().Get("room"))
	if username == "" || roomID == "" {
		http.Error(w, "missing username or room", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	client := &Client{conn: conn, username: username, roomID: roomID}

	// register
	roomsMu.Lock()
	if rooms[roomID] == nil {
		rooms[roomID] = make(map[*Client]bool)
	}
	rooms[roomID][client] = true
	roomsMu.Unlock()

	// send welcome or system info (option)
	sys := map[string]interface{}{
		"type": "system",
		"text": fmt.Sprintf("Bạn đã vào phòng %s với tên %s", roomID, username),
		"ts":   time.Now().UnixMilli(),
	}
	_ = conn.WriteJSON(sys)

	// read loop
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("read err:", err)
			break
		}
		// treat data as text content
		content := string(data)
		m := &Message{
			Room:    roomID,
			Sender:  username,
			Content: content,
			Ts:      time.Now().UnixMilli(),
		}
		// save
		if err := saveMessage(m); err != nil {
			log.Println("save msg err:", err)
		}

		// Broadcast message object to all in room
		payload := map[string]interface{}{
			"type":    "message",
			"message": m,
		}
		broadcastToRoom(roomID, payload)
	}

	// cleanup on disconnect
	roomsMu.Lock()
	delete(rooms[roomID], client)
	roomsMu.Unlock()
	conn.Close()
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	// GET /history?room=ROOM&page=1&limit=30
	room := r.URL.Query().Get("room")
	if room == "" {
		http.Error(w, "missing room", http.StatusBadRequest)
		return
	}
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page := 1
	limit := 30
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	msgs, err := loadHistory(room, page, limit)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

func main() {
	var err error
	// DB file inside backend folder
	dbPath := filepath.Join(".", "chat.db")
	db, err = ensureDB(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// static dist detection (frontend/dist relative)
	distDir := filepath.Join("..", "frontend", "dist")
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		// fallback: ./dist
		distDir = "dist"
	}
	// Serve SPA static
	fs := http.FileServer(http.Dir(distDir))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleWS)
	http.HandleFunc("/history", historyHandler)

	addr := "0.0.0.0:8080"
	log.Printf("listening %s static=%s\n", addr, distDir)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
