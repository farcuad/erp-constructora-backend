package notifications

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSHub struct {
	// Mapa: clients[companyID][userID] -> *websocket.Conn
	clients map[string]map[string]*websocket.Conn
	mu      sync.RWMutex
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[string]map[string]*websocket.Conn),
	}
}

func (h *WSHub) Register(companyID, userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[companyID]; !ok {
		h.clients[companyID] = make(map[string]*websocket.Conn)
	}
	h.clients[companyID][userID] = conn
}

func (h *WSHub) Unregister(companyID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if userMap, ok := h.clients[companyID]; ok {
		if conn, ok := userMap[userID]; ok {
			conn.Close()
			delete(userMap, userID)
		}
	}
}

// SendToUsers envía la notificación por WebSockets solo a los usuarios indicados que tengan la app abierta
func (h *WSHub) SendToUsers(companyID string, targetUserIDs []string, notification Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userMap, ok := h.clients[companyID]
	if !ok {
		return
	}

	for _, userID := range targetUserIDs {
		if conn, ok := userMap[userID]; ok {
			_ = conn.WriteJSON(notification)
		}
	}
}
