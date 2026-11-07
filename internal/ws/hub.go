package ws

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	ID       int
	UUID     string
	UserName string
}

type Hub struct {
	// OnMessage handles message types the hub does not know. nil rejects them.
	OnMessage func(c *Client, m Message)

	mu      sync.RWMutex
	clients map[*Client]struct{}
	users   map[string]map[*Client]struct{}
	closed  bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		users:   make(map[string]map[*Client]struct{}),
	}
}

// Serve registers the connection and blocks until it ends.
func (h *Hub) Serve(conn *websocket.Conn, u User) {
	c := newClient(h, conn, u)
	if !h.add(c) { // hub is shutting down
		_ = conn.Close()
		return
	}

	go c.writePump()
	c.readPump() // blocks
	c.Close()
}

// EmitUser sends an event to every connection of one user.
func (h *Hub) EmitUser(userUUID, typ string, data any) {
	h.mu.RLock()
	set := h.users[userUUID]
	targets := make([]*Client, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	payload := encode(typ, data)
	for _, c := range targets {
		c.Send(payload)
	}
}

// EmitAll sends an event to every connected client.
func (h *Hub) EmitAll(typ string, data any) {
	h.mu.RLock()
	targets := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	payload := encode(typ, data)
	for _, c := range targets {
		c.Send(payload)
	}
}

// DisconnectUser closes every connection of a user, for example after a new login.
func (h *Hub) DisconnectUser(userUUID string) {
	h.mu.RLock()
	set := h.users[userUUID]
	targets := make([]*Client, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		c.Close()
	}
}

// Shutdown closes every connection. Call it before stopping the server.
func (h *Hub) Shutdown() {
	h.mu.Lock()
	h.closed = true
	targets := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		targets = append(targets, c)
	}
	h.mu.Unlock()

	for _, c := range targets {
		c.Close()
	}
}

func (h *Hub) Stats() (clients, users int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients), len(h.users)
}

func (h *Hub) add(c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return false
	}
	h.clients[c] = struct{}{}

	set := h.users[c.User.UUID]
	if set == nil {
		set = make(map[*Client]struct{})
		h.users[c.User.UUID] = set
	}
	set[c] = struct{}{}
	return true
}

func (h *Hub) remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[c]; !ok {
		return
	}
	delete(h.clients, c)

	if set := h.users[c.User.UUID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.users, c.User.UUID)
		}
	}
}

// handle routes one inbound message.
func (h *Hub) handle(c *Client, m Message) {
	switch m.Type {
	case TypePing:
		c.Send(encode(TypePong, nil))
	default:
		if h.OnMessage != nil {
			h.OnMessage(c, m)
			return
		}
		c.Send(errMsg("unknown_type"))
	}
}
