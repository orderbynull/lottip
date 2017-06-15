package main

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	pingPeriod          = time.Millisecond * 5000
	writeDeadlinePeriod = time.Second * 2
)

// client represents client connected via websocket
type client struct {
	ws       *websocket.Conn
	hub      *hub
	dataChan chan []byte
}

// newClient creates new client instance
func newClient(ws *websocket.Conn, hub *hub) *client {
	return &client{
		ws:       ws,
		hub:      hub,
		dataChan: make(chan []byte),
	}
}

// process ...
func (c *client) process() {
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
