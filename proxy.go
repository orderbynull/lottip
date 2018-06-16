package main

import (
	"fmt"
	"github.com/orderbynull/lottip/chat"
	"github.com/orderbynull/lottip/protocol"
	"io"
	"log"
	"net"
)

type RequestPacketParser struct {
	connId        string
	cmdId         *int
	cmdChan       chan chat.Cmd
	connStateChan chan chat.ConnState
}

func (pp *RequestPacketParser) Write(p []byte) (n int, err error) {
	*pp.cmdId++

	switch protocol.GetPacketType(p) {
	case protocol.ComStmtPrepare:
	case protocol.ComQuery:
		decoded, err := protocol.DecodeQueryRequest(p)
		if err == nil {
			pp.cmdChan <- chat.Cmd{pp.connId, *pp.cmdId, "", decoded.Query, nil, false}
		}
	case protocol.ComQuit:
		pp.connStateChan <- chat.ConnState{pp.connId, protocol.ConnStateFinished}
	}

	return len(p), nil
}

type ResponsePacketParser struct {
	connId        string
	cmdId         *int
	cmdResultChan chan chat.CmdResult
}

func (pp *ResponsePacketParser) Write(p []byte) (n int, err error) {
	switch protocol.GetPacketType(p) {
	case protocol.ResponseErr:
		pp.cmdResultChan <- chat.CmdResult{pp.connId, *pp.cmdId, protocol.ResponseErr, "Fuck!", "1s"}
	default:
		pp.cmdResultChan <- chat.CmdResult{pp.connId, *pp.cmdId, protocol.ResponseOk, "", "1s"}
	}

	return len(p), nil
}

// proxy implements server for capturing and forwarding MySQL traffic.
type proxy struct {
	cmdChan       chan chat.Cmd
	cmdResultChan chan chat.CmdResult
	connStateChan chan chat.ConnState
	appReadyChan  chan bool
	mysqlHost     string
	proxyHost     string
}

// run starts accepting TCP connection and forwarding it to MySQL server.
// Each incoming TCP connection is handled in own goroutine.
func (p *proxy) run() {
	listener, err := net.Listen("tcp", p.proxyHost)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()

	go func() {
		p.appReadyChan <- true
		close(p.appReadyChan)
	}()

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Print(err.Error())
		}

		go p.handleConnection(client)
	}
}

// handleConnection ...
func (p *proxy) handleConnection(client net.Conn) {
	defer client.Close()

	// New connection to MySQL is made per each incoming TCP request to proxy server.
	server, err := net.Dial("tcp", p.mysqlHost)
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer server.Close()

	connId := fmt.Sprintf("%s => %s", client.RemoteAddr().String(), server.RemoteAddr().String())

	defer func() { p.connStateChan <- chat.ConnState{connId, protocol.ConnStateFinished} }()

	var cmdId int

	// Copy bytes from client to server and requestParser
	go io.Copy(io.MultiWriter(server, &RequestPacketParser{connId, &cmdId, p.cmdChan, p.connStateChan}), client)

	// Copy bytes from server to client and responseParser
	io.Copy(io.MultiWriter(client, &ResponsePacketParser{connId, &cmdId, p.cmdResultChan}), server)
}
