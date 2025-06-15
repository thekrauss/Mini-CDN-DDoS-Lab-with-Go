package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients map[*websocket.Conn]bool
	lock    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.lock.Lock()
	h.clients[conn] = true
	h.lock.Unlock()
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.lock.Lock()
	delete(h.clients, conn)
	h.lock.Unlock()
	conn.Close()
}

func (h *Hub) Broadcast(message []byte) {
	h.lock.RLock()
	for conn := range h.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
	h.lock.RUnlock()
}
