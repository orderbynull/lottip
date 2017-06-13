package main

import (
	"encoding/json"
	"github.com/orderbynull/lottip/proxy"
)

// Hub ...
type Hub struct {
	clients       map[*Client]bool
	register      chan *Client
	deregister    chan *Client
	cmdChan       chan proxy.Cmd
	cmdResultChan chan proxy.CmdResult
	connStateChan chan proxy.ConnState
}

// newHub ...
func newHub(
	cmdChan chan proxy.Cmd,
	cmdResultChan chan proxy.CmdResult,
	connStateChan chan proxy.ConnState,
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

// registerClient...
func (h *Hub) registerClient(client *Client) {
	h.register <- client
}

// run ...
func (h *Hub) run() {
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
