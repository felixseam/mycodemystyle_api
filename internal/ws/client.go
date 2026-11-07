package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"

	"mycodemystyle_api/internal/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second // must be shorter than pongWait
	maxMessageSize = 4096
	sendBuffer     = 256
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	User User

	send chan []byte
	done chan struct{}
	once sync.Once
}

func newClient(h *Hub, conn *websocket.Conn, u User) *Client {
	return &Client{
		hub:  h,
		conn: conn,
		User: u,
		send: make(chan []byte, sendBuffer),
		done: make(chan struct{}),
	}
}

// Send queues a message for this client. A client that cannot keep up
// (full buffer) is disconnected, so one slow phone never blocks the hub.
func (c *Client) Send(msg []byte) {
	select {
	case <-c.done:
	case c.send <- msg:
	default:
		logger.NewCustomLog("ws_slow_client", "buffer full, disconnecting "+c.User.UserName, "warn")
		c.Close()
	}
}

// Close is safe to call many times, from any goroutine.
func (c *Client) Close() {
	c.once.Do(func() {
		close(c.done) // writePump sees this, sends a close frame, closes the socket
		c.hub.remove(c)
	})
}

// readPump runs in the handler goroutine and returns when the connection ends.
func (c *Client) readPump() {
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.NewCustomLog("ws_read", err.Error(), "debug")
			}
			return
		}

		var m Message
		if err := json.Unmarshal(raw, &m); err != nil {
			c.Send(errMsg("invalid_json"))
			continue
		}
		c.hub.handle(c, m)
	}
}

// writePump is the ONLY place that writes to the connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.conn.Close()

	for {
		select {
		case msg := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.Close()
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Close()
				return
			}

		case <-c.done:
			_ = c.conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseGoingAway, ""),
				time.Now().Add(writeWait),
			)
			return
		}
	}
}
