package main

import (
	"encoding/json"
	"github.com/orderbynull/lottip/proxy"
)

// hub ...
type hub struct {
	clients       map[*client]bool
	register      chan *client
	deregister    chan *client
	cmdChan       chan proxy.Cmd
	cmdResultChan chan proxy.CmdResult
	connStateChan chan proxy.ConnState
}

// newHub ...
func newHub(
	cmdChan chan proxy.Cmd,
	cmdResultChan chan proxy.CmdResult,
	connStateChan chan proxy.ConnState,
) *hub {
	return &hub{
		clients:       make(map[*client]bool),
		register:      make(chan *client),
		deregister:    make(chan *client),
		cmdChan:       cmdChan,
		cmdResultChan: cmdResultChan,
		connStateChan: connStateChan,
	}
}

// registerClient...
func (h *hub) registerClient(client *client) {
	h.register <- client
}

// run ...
func (h *hub) run() {
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
