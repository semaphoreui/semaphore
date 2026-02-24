package sockets

import log "github.com/sirupsen/logrus"

// Broadcaster provides cross-node WebSocket message delivery for HA setups.
// When configured, Message() delegates to the broadcaster which publishes
// messages to all nodes in the cluster via Redis Pub/Sub.
type Broadcaster interface {
	// Start begins listening for messages from other nodes.
	Start()
	// Publish delivers a message to all nodes in the cluster.
	// The implementation must also deliver the message to local clients
	// by calling LocalBroadcast.
	Publish(userID int, msg []byte)
	// Stop shuts down the broadcaster.
	Stop()
}

var broadcaster Broadcaster

// SetBroadcaster configures a cross-node broadcaster for HA mode.
// When set, Message() delegates to the broadcaster instead of the local hub.
func SetBroadcaster(b Broadcaster) {
	broadcaster = b
}

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

// LocalBroadcast delivers a message to locally-connected WebSocket clients
// only. Used by Broadcaster implementations to relay messages received from
// other nodes without re-publishing them.
func LocalBroadcast(userID int, message []byte) {
	h.broadcast <- &sendRequest{
		userID: userID,
		msg:    message,
	}
}
