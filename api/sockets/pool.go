package sockets

import log "github.com/sirupsen/logrus"

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered websocket connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan *sendRequest

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

type sendRequest struct {
	userID int
	msg    []byte
}

var h = hub{
	broadcast:   make(chan *sendRequest),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for conn := range h.connections {
				if m.userID > 0 && m.userID != conn.userID {
					continue
				}

				select {
				case conn.send <- m.msg:
				default:

					log.WithFields(log.Fields{
						"context": "websocket",
						"user_id": conn.userID,
					}).Error("Connection send channel is full, connection closing")

					close(conn.send)
					delete(h.connections, conn)
					_ = conn.ws.Close() // Close the WebSocket connection first
				}
			}
		}
	}
}

// StartWS starts the web sockets in a goroutine
func StartWS() {
	h.run()
}
