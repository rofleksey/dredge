package ws

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
	// sendQueue caps outbound broadcast backlog per client; full queue drops newer messages (see BroadcastJSON).
	sendQueue = 16
)

// NewHub builds a hub that only accepts browser WebSocket connections whose Origin matches allowedOrigin
// (e.g. http://localhost:5173). Requests with no Origin header are allowed (non-browser clients).
func NewHub(allowedOrigin string) *Hub {
	allowedOrigin = strings.TrimSpace(allowedOrigin)

	return &Hub{
		allowedOrigin: allowedOrigin,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				o := r.Header.Get("Origin")
				if o == "" {
					return true
				}

				if allowedOrigin == "" {
					return false
				}

				return o == allowedOrigin
			},
		},
		clients: make(map[*client]struct{}),
	}
}

// Upgrade authenticates the client and registers the connection for broadcast.
// Optional initial messages are queued to this client before pumpConn starts (same drop policy as BroadcastJSON).
func (h *Hub) Upgrade(w http.ResponseWriter, r *http.Request, userID int64, initial ...any) error {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	cl := &client{conn: conn, userID: userID, send: make(chan any, sendQueue)}

	h.mu.Lock()
	h.clients[cl] = struct{}{}
	h.mu.Unlock()

	for _, msg := range initial {
		select {
		case cl.send <- msg:
		default:
			// Same policy as BroadcastJSON: drop if the client is already backlogged.
		}
	}

	go h.pumpConn(cl)

	return nil
}

func (h *Hub) pumpConn(cl *client) {
	conn := cl.conn
	defer h.removeClient(cl)

	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	errCh := make(chan error, 1)

	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				return
			}
		case msg := <-cl.send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}
		}
	}
}

func (h *Hub) removeClient(cl *client) {
	h.mu.Lock()
	delete(h.clients, cl)
	h.mu.Unlock()

	_ = cl.conn.Close()
}

func (h *Hub) BroadcastJSON(v any) {
	h.mu.Lock()

	clients := make([]*client, 0, len(h.clients))
	for cl := range h.clients {
		clients = append(clients, cl)
	}
	h.mu.Unlock()

	for _, cl := range clients {
		select {
		case cl.send <- v:
		default:
			// Drop when the client is slow; pumpConn will keep draining pings and prior messages.
		}
	}
}
