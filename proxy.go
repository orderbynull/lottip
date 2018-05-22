package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/orderbynull/lottip/chat"
	"github.com/orderbynull/lottip/protocol"
)

var errInvalidProxyParams = errors.New("main: both proxy and mysql hosts must be set")

// proxy implements server for capturing and forwarding MySQL traffic
type proxy struct {
	rwmu          sync.RWMutex
	handshakes    map[uint32]*protocol.ConnSettings
	cmdChan       chan chat.Cmd
	cmdStateChan  chan chat.CmdResult
	connStateChan chan chat.ConnState
	appReadyChan  chan bool
	mysqlHost     string
	proxyHost     string
}

// newProxy returns new proxy with connections params for proxy and mysql hosts.
// Returns error if either proxyHost or mysqlHost not set.
func newProxy(proxyHost string, mysqlHost string) (*proxy, error) {
	if proxyHost == "" || mysqlHost == "" {
		return nil, errInvalidProxyParams
	}

	return &proxy{handshakes: make(map[uint32]*protocol.ConnSettings), proxyHost: proxyHost, mysqlHost: mysqlHost}, nil
}

// run starts accepting TCP connection and forwarding it to MySQL server.
// Each incoming TCP connection is handled in own goroutine.
func (ps *proxy) run() {
	listener, err := net.Listen("tcp", ps.proxyHost)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()

	go func() {
		ps.appReadyChan <- true
		close(ps.appReadyChan)
	}()

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Print(err.Error())
		}

		go ps.handleConnection(client)
	}
}

// setChannels assigns user defined channels to proxy.
// This channels are used to transfer captured command(query), command state and
// connection state to corresponding routine.
func (ps *proxy) setChannels(
	cmdChan chan chat.Cmd,
	cmdStateChan chan chat.CmdResult,
	connStateChan chan chat.ConnState,
	appReadyChan chan bool,
) {
	ps.cmdChan = cmdChan
	ps.cmdStateChan = cmdStateChan
	ps.connStateChan = connStateChan
	ps.appReadyChan = appReadyChan
}

// setCommand writes command string representation and it's id to command channel
// provided by caller code via newProxy routine
func (ps *proxy) setCommand(
	connID uint32,
	cmdID int,
	database string,
	query string,
	params []protocol.PreparedParameter,
	executable bool,
) {
	var parametersSlice []string
	for _, parameter := range params {
		parametersSlice = append(parametersSlice, parameter.Value)
	}

	ps.cmdChan <- chat.Cmd{
		ConnId:     connID,
		CmdId:      cmdID,
		Database:   database,
		Query:      query,
		Executable: executable,
		Parameters: parametersSlice,
	}
}

// setCommandResult writes command execution result to command result channel
// provided by caller code via newProxy routine
func (ps *proxy) setCommandResult(connId uint32, cmdId int, cmdState byte, error string, duration time.Duration) {
	ps.cmdStateChan <- chat.CmdResult{
		ConnId:   connId,
		CmdId:    cmdId,
		Result:   cmdState,
		Error:    error,
		Duration: fmt.Sprintf("%.2f", duration.Seconds()),
	}
}

// setConnectionState writes TCP connection state to connection state channel
// provided by caller code via newProxy routine
func (ps *proxy) setConnectionState(connId uint32, state byte) {
	ps.connStateChan <- chat.ConnState{ConnId: connId, State: state}
}

// handleConnection ...
func (ps *proxy) handleConnection(client net.Conn) {
	defer client.Close()

	// Establishing connection to MySQL server for forwarding packets.
	// New connection is made per each TCP request to proxy server.
	server, err := net.Dial("tcp", ps.mysqlHost)
	if err != nil {
		log.Print(err.Error())
		return
	}

	// Read server and client capabilities
	serverHandshake, clientHandshake, err := protocol.ProcessHandshake(client, server)
	if err != nil {
		println(err.Error())
		return
	}

	connSettings := &protocol.ConnSettings{
		ClientCapabilities: clientHandshake.ClientCapabilities,
		ServerCapabilities: serverHandshake.ServerCapabilities,
	}

	defer ps.setConnectionState(serverHandshake.ConnectionID, protocol.ConnStateFinished)
	ps.setConnectionState(serverHandshake.ConnectionID, protocol.ConnStateStarted)

	ps.rwmu.Lock()
	ps.handshakes[serverHandshake.ConnectionID] = connSettings
	ps.rwmu.Unlock()

	defer func() {
		ps.rwmu.Lock()
		delete(ps.handshakes, serverHandshake.ConnectionID)
		ps.rwmu.Unlock()
	}()

	ps.process(client, server, serverHandshake.ConnectionID)
}

// process reads data from proxy client, extracts queries and forwards them to MySQL
func (ps *proxy) process(client net.Conn, mysql net.Conn, connID uint32) {
	var cmdId int
	var preparedQuery string
	var preparedParamsCount uint16
	ps.rwmu.RLock()
	var deprecateEof = ps.handshakes[connID].DeprecateEOFSet()
	ps.rwmu.RUnlock()

	for {
		//Client query --> $requestPacket --> mysql
		requestPacket, err := protocol.ReadPacket(client)
		if err != nil {
			break
		}

		cmdId++

		// There're packets which have zero length payload
		// and there's no need to analyze such packets.
		if len(requestPacket) < 5 {
			protocol.WritePacket(requestPacket, mysql)
			pkt, _ := protocol.ReadPacket(mysql)
			protocol.WritePacket(pkt, client)

			continue
		}

		switch requestPacket[4] {

		// Received COM_QUERY from client
		case protocol.ComQuery:
			decoded, _ := protocol.DecodeQueryRequest(requestPacket)

			selectedDb := getUseDatabaseValue(decoded.Query)
			if len(selectedDb) > 0 {
				ps.rwmu.RLock()
				ps.handshakes[connID].SelectedDb = selectedDb
				ps.rwmu.RUnlock()
			}

			ps.rwmu.RLock()
			selectedDb = ps.handshakes[connID].SelectedDb
			ps.rwmu.RUnlock()
			ps.setCommand(
				connID,
				cmdId,
				selectedDb,
				decoded.Query,
				[]protocol.PreparedParameter{},
				true,
			)

			start := time.Now()
			protocol.WritePacket(requestPacket, mysql)

			response, result, err := protocol.ReadResponse(mysql, deprecateEof)
			if err == nil {
				if result == protocol.ResponseErr {
					ps.setCommandResult(connID, cmdId, result, protocol.ReadErrMessage(response), time.Since(start))
				} else {
					ps.setCommandResult(connID, cmdId, result, "", time.Since(start))
				}

				protocol.WritePacket(response, client)
			}

		// Received COM_STMT_PREPARE from client
		case protocol.ComStmtPrepare:
			decoded, _ := protocol.DecodeQueryRequest(requestPacket)

			selectedDb := getUseDatabaseValue(decoded.Query)
			if len(selectedDb) > 0 {
				ps.rwmu.RLock()
				ps.handshakes[connID].SelectedDb = selectedDb
				ps.rwmu.RUnlock()
			}

			protocol.WritePacket(requestPacket, mysql)

			response, _, err := protocol.ReadPrepareResponse(mysql)
			if err == nil {
				decodedResponse, err := protocol.DecodeComStmtPrepareOkResponse(response)
				if err == nil {
					preparedParamsCount = decodedResponse.ParametersNum
					preparedQuery = decoded.Query
				}

				protocol.WritePacket(response, client)
			}

		// Received comStmtExecute from MySQL client
		case protocol.ComStmtExecute:
			var preparedParameters []protocol.PreparedParameter
			var executable bool

			decodedRequest, err := protocol.DecodeComStmtExecuteRequest(requestPacket, preparedParamsCount)
			if err == nil {
				preparedParameters = decodedRequest.PreparedParameters
				executable = true
			}

			ps.rwmu.RLock()
			selectedDb := ps.handshakes[connID].SelectedDb
			ps.rwmu.RUnlock()

			ps.setCommand(
				connID,
				cmdId,
				selectedDb,
				preparedQuery,
				preparedParameters,
				executable,
			)

			start := time.Now()

			protocol.WritePacket(requestPacket, mysql)
			response, result, _ := protocol.ReadResponse(mysql, deprecateEof)
			ps.setCommandResult(connID, cmdId, result, "", time.Since(start))
			protocol.WritePacket(response, client)

		case protocol.ComFieldList:
			protocol.WritePacket(requestPacket, mysql)
			response, _, _ := protocol.ReadShowFieldsResponse(mysql)
			protocol.WritePacket(response, client)

		// Received COM_STMT_CLOSE from client
		case protocol.ComStmtClose:
			continue

		default:
			protocol.WritePacket(requestPacket, mysql)
			pkt, _ := protocol.ReadPacket(mysql)
			protocol.WritePacket(pkt, client)
		}
	}
}
