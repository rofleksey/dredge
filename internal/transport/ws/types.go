package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	conn   *websocket.Conn
	userID int64
	send   chan any
}

type Hub struct {
	allowedOrigin string
	upgrader      websocket.Upgrader
	mu            sync.Mutex
	clients       map[*client]struct{}
}
