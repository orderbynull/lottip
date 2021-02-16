package main

import (
	"fmt"
	"github.com/brainz80/lottip/chat"
	"github.com/brainz80/lottip/protocol"
	"io"
	"log"
	"net"
	"time"
)

type RequestPacketParser struct {
	connId        string
	queryId       *int
	queryChan     chan chat.Cmd
	connStateChan chan chat.ConnState
	timer         *time.Time
}

func (pp *RequestPacketParser) Write(p []byte) (n int, err error) {
	*pp.queryId++
	*pp.timer = time.Now()

	switch protocol.GetPacketType(p) {
	case protocol.ComStmtPrepare:
	case protocol.ComQuery:
		decoded, err := protocol.DecodeQueryRequest(p)
		if err == nil {
			pp.queryChan <- chat.Cmd{pp.connId, *pp.queryId, "", decoded.Query, nil, false}
		}
	case protocol.ComQuit:
		pp.connStateChan <- chat.ConnState{pp.connId, protocol.ConnStateFinished}
	}

	return len(p), nil
}

type ResponsePacketParser struct {
	connId          string
	queryId         *int
	queryResultChan chan chat.CmdResult
	timer           *time.Time
}

func (pp *ResponsePacketParser) Write(p []byte) (n int, err error) {
	duration := fmt.Sprintf("%.3f", time.Since(*pp.timer).Seconds())

	switch protocol.GetPacketType(p) {
	case protocol.ResponseErr:
		decoded, _ := protocol.DecodeErrResponse(p)
		pp.queryResultChan <- chat.CmdResult{pp.connId, *pp.queryId, protocol.ResponseErr, decoded, duration}
	default:
		pp.queryResultChan <- chat.CmdResult{pp.connId, *pp.queryId, protocol.ResponseOk, "", duration}
	}

	return len(p), nil
}

// MySQLProxyServer implements server for capturing and forwarding MySQL traffic.
type MySQLProxyServer struct {
	cmdChan       chan chat.Cmd
	cmdResultChan chan chat.CmdResult
	connStateChan chan chat.ConnState
	appReadyChan  chan bool
	mysqlHost     string
	proxyHost     string
}

// run starts accepting TCP connection and forwarding it to MySQL server.
// Each incoming TCP connection is handled in own goroutine.
func (p *MySQLProxyServer) run() {
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
func (p *MySQLProxyServer) handleConnection(client net.Conn) {
	defer client.Close()

	// New connection to MySQL is made per each incoming TCP request to MySQLProxyServer server.
	server, err := net.Dial("tcp", p.mysqlHost)
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer server.Close()

	connId := fmt.Sprintf("%s => %s", client.RemoteAddr().String(), server.RemoteAddr().String())

	defer func() { p.connStateChan <- chat.ConnState{connId, protocol.ConnStateFinished} }()

	var queryId int
	var timer time.Time

	// Copy bytes from client to server and requestParser
	go io.Copy(io.MultiWriter(server, &RequestPacketParser{connId, &queryId, p.cmdChan, p.connStateChan, &timer}), client)

	// Copy bytes from server to client and responseParser
	io.Copy(io.MultiWriter(client, &ResponsePacketParser{connId, &queryId, p.cmdResultChan, &timer}), server)
}
