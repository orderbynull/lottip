package main

import (
	"io"
	"lottip/chat"
	"lottip/protocol"
)

type ClientToServerHandler struct {
	connInfo      *ConnectionInfo
	queryChan     chan chat.Cmd
	connStateChan chan chat.ConnState
	server        io.Writer
}

// CLIENT to SERVER
func (pp *ClientToServerHandler) Write(buffer []byte) (n int, err error) {
	extractPacketsFromBuffer(pp.connInfo, pp.connInfo.clientPacketFragment, buffer, func(packet []byte) {
		// Switch based on the state
		fsm := pp.connInfo.fsm

		if ok, _ := fsm.IsInState(StateIdle); ok {
			pp.connStateChan <- chat.ConnState{pp.connInfo.ConnId, protocol.ConnStateFinished}
		} else if ok, _ := fsm.IsInState(StateAuthRequested); ok {
			fsm.Fire(MsgLogin, packet)
		} else {
			if len(packet) < 4 {
				LogInvalid(pp.connInfo, "?Request?", packet)
			} else if GetPacketType(packet) == protocol.ComPing {
				fsm.Fire(GetPacketType(packet), packet)
			} else {
				// We are processing queries
				fsm.Fire(GetPacketType(packet), packet)
			}
		}
	})

	return len(buffer), nil
}
