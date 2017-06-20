package proxy

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// proxyServer implements server for capturing and forwarding MySQL traffic
type proxyServer struct {
	//...
	handshakes map[int]*handshake

	//...
	cmdChan chan Cmd

	//...
	cmdStateChan chan CmdResult

	//...
	connStateChan chan ConnState

	//...
	appReadyChan chan bool

	//...
	mysqlHost string

	//...
	proxyHost string
}

// NewProxyServer returns new proxyServer with connections params for proxy and mysql hosts.
// Returns error if either proxyHost or mysqlHost not set.
func NewProxyServer(proxyHost string, mysqlHost string) (*proxyServer, error) {
	if proxyHost == "" || mysqlHost == "" {
		return nil, errInvalidProxyParams
	}

	return &proxyServer{handshakes: make(map[int]*handshake), proxyHost: proxyHost, mysqlHost: mysqlHost}, nil
}

//...
//@TODO check for existence of handshake
func (ps *proxyServer) getHandshake(connID int) *handshake {
	return ps.handshakes[connID]
}

//...
func (ps *proxyServer) setHandshake(connID int, handshake *handshake) {
	ps.handshakes[connID] = handshake
}

//...
func (ps *proxyServer) removeHandshake(connID int) {
	delete(ps.handshakes, connID)
}

// SetChannels assigns user defined channels to proxyServer.
// This channels are used to transfer captured command(query), command state and
// connection state to corresponding routine.
func (ps *proxyServer) SetChannels(
	cmdChan chan Cmd,
	cmdStateChan chan CmdResult,
	connStateChan chan ConnState,
	appReadyChan chan bool,
) {
	ps.cmdChan = cmdChan
	ps.cmdStateChan = cmdStateChan
	ps.connStateChan = connStateChan
	ps.appReadyChan = appReadyChan
}

// setCommand writes command string representation and it's id to command channel
// provided by caller code via NewProxyServer routine
func (ps *proxyServer) setCommand(connID int, cmdID int, query string, executable bool) {
	ps.cmdChan <- Cmd{ConnId: connID, CmdId: cmdID, Query: query, Executable: executable}
}

// setCommandResult writes command execution result to command result channel
// provided by caller code via NewProxyServer routine
func (ps *proxyServer) setCommandResult(connId int, cmdId int, cmdState byte, error string, duration time.Duration) {
	ps.cmdStateChan <- CmdResult{
		ConnId:   connId,
		CmdId:    cmdId,
		Result:   cmdState,
		Error:    error,
		Duration: fmt.Sprintf("%.2f", duration.Seconds()),
	}
}

// setConnectionState writes TCP connection state to connection state channel
// provided by caller code via NewProxyServer routine
func (ps *proxyServer) setConnectionState(connId int, state byte) {
	ps.connStateChan <- ConnState{ConnId: connId, State: state}
}

// handleConnection ...
func (ps *proxyServer) handleConnection(connID int, conn net.Conn) {
	defer conn.Close()
	defer ps.removeHandshake(connID)

	// Establishing connection to MySQL server for proxying packets
	// New connection is made per each TCP request to proxy server
	mysql, err := net.Dial("tcp", ps.mysqlHost)
	if err != nil {
		log.Print(err.Error())
		return
	}

	// Both calls to setConnectionState used to update connection state
	// on connection open and close events
	defer ps.setConnectionState(connID, connStateFinished)
	ps.setConnectionState(connID, connStateStarted)

	ps.setHandshake(connID, processHandshake(conn, mysql))

	ps.extractAndForward(conn, mysql, connID)
}

// extractAndForward reads data from proxy client, extracts queries and forwards them to MySQL
func (ps *proxyServer) extractAndForward(conn net.Conn, mysql net.Conn, connID int) {
	var cmdId int
	var deprecateEof = ps.getHandshake(connID).deprecateEOF()

	for {
		//Client query --> $queryPacket - -> mysql
		queryPacket, err := readPacket(conn)
		if err != nil {
			break
		}

		cmdId++

		// There're packets which have zero length payload
		// and there's no need to analyze such packets.
		if len(queryPacket) < 5 {
			writePacket(queryPacket, mysql)
			pkt, _ := readPacket(mysql)
			writePacket(pkt, conn)

			continue
		}

		switch queryPacket[4] {

		// Received COM_QUERY from client
		case requestCmdQuery:
			query, _ := getQueryString(queryPacket)

			selectedDb := haventYetDecidedFuncName(query)
			if len(selectedDb) > 0 {
				ps.getHandshake(connID).setSelectedDb(selectedDb)
			}

			ps.setCommand(connID, cmdId, query, true)

			start := time.Now()
			writePacket(queryPacket, mysql)

			response, result, err := readResponse(mysql, deprecateEof)
			if err == nil {
				if result == responseErr {
					ps.setCommandResult(connID, cmdId, result, readErrMessage(response), time.Since(start))
				} else {
					ps.setCommandResult(connID, cmdId, result, "", time.Since(start))
				}

				writePacket(response, conn)
			}

		// Received COM_STMT_PREPARE from client
		case requestCmdStmtPrepare:
			query, _ := getQueryString(queryPacket)

			selectedDb := haventYetDecidedFuncName(query)
			if len(selectedDb) > 0 {
				ps.getHandshake(connID).setSelectedDb(selectedDb)
			}

			ps.setCommand(connID, cmdId, query, false)

			start := time.Now()
			writePacket(queryPacket, mysql)

			response, result, err := readPrepareResponse(mysql)
			if err == nil {
				ps.setCommandResult(connID, cmdId, result, "", time.Since(start))

				writePacket(response, conn)
			}

		// Received COM_STMT_EXECUTE from MySQL client
		case requestCmdStmtExecute:
			writePacket(queryPacket, mysql)
			response, _, _ := readResponse(mysql, deprecateEof)
			writePacket(response, conn)

		case requestCmdShowFields:
			writePacket(queryPacket, mysql)
			response, _, _ := readShowFieldsResponse(mysql)
			writePacket(response, conn)

		// Received COM_STMT_CLOSE from client
		case requestCmdStmtClose:
			continue

		default:
			writePacket(queryPacket, mysql)
			pkt, _ := readPacket(mysql)
			writePacket(pkt, conn)
		}
	}
}

// Run starts accepting TCP connections and forwarding them to MySQL server.
// Each incoming TCP connection is handled in own goroutine.
func (ps *proxyServer) Run() {
	listener, err := net.Listen("tcp", ps.proxyHost)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()

	go func() {
		ps.appReadyChan <- true
		close(ps.appReadyChan)
	}()

	var connectionID int
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err.Error())
		}

		connectionID++

		go ps.handleConnection(connectionID, conn)
	}
}
