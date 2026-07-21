package realtime

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type SeatEvent struct {
	Type   string `json:"type"`
	SeatID string `json:"seat_id"`
	Status string `json:"status"`
}

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) HandleConnection(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connection, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.clients[connection] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, connection)
		h.mu.Unlock()

		connection.Close()
	}()

	for {
		if _, _, err := connection.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Hub) Broadcast(event SeatEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for connection := range h.clients {
		if err := connection.WriteMessage(websocket.TextMessage, data); err != nil {
			connection.Close()
			delete(h.clients, connection)
		}
	}
}
