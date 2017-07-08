package chat

import (
	"encoding/json"
)

// Hub ...
type Hub struct {
	clients       map[*Client]bool
	register      chan *Client
	deregister    chan *Client
	cmdChan       chan Cmd
	cmdResultChan chan CmdResult
	connStateChan chan ConnState
}

// NewHub ...
func NewHub(
	cmdChan chan Cmd,
	cmdResultChan chan CmdResult,
	connStateChan chan ConnState,
) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		register:      make(chan *Client),
		deregister:    make(chan *Client),
		cmdChan:       cmdChan,
		cmdResultChan: cmdResultChan,
		connStateChan: connStateChan,
	}
}

// RegisterClient...
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// Run ...
func (h *Hub) Run() {
	var data []byte
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.deregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.dataChan)
			}

		case cmd := <-h.cmdChan:
			data, _ = json.Marshal(cmd)

		case cmdResult := <-h.cmdResultChan:
			data, _ = json.Marshal(cmdResult)

		case connState := <-h.connStateChan:
			data, _ = json.Marshal(connState)
		}

		for client := range h.clients {
			if len(data) > 0 {
				client.dataChan <- data
			}
		}
	}
}
