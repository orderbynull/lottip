package chat

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	pingPeriod          = time.Millisecond * 5000
	writeDeadlinePeriod = time.Second * 2
)

// Client represents client(browser) connected via websocket
type Client struct {
	ws       *websocket.Conn
	hub      *Hub
	dataChan chan []byte
}

// NewClient creates new Client instance
func NewClient(ws *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ws:       ws,
		hub:      hub,
		dataChan: make(chan []byte),
	}
}

// Process ...
func (c *Client) Process() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		c.ws.Close()
		c.hub.deregister <- c
	}()

	for {
		select {
		case data, ok := <-c.dataChan:
			c.ws.SetWriteDeadline(time.Now().Add(writeDeadlinePeriod))
			if !ok {
				return
			}

			if err := c.ws.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeDeadlinePeriod))
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
