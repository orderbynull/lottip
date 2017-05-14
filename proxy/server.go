package proxy

import (
	"fmt"
	"log"
	"net"
	"time"
)

// ProxyServer implements server for capturing and forwarding MySQL traffic
type ProxyServer struct {
	cmdChan       chan Cmd
	cmdStateChan  chan CmdResult
	connStateChan chan ConnState
	appReadyChan  chan bool
	mysqlHost     string
	proxyHost     string
}

// NewProxyServer returns new ProxyServer with connections params for proxy and mysql hosts.
// Returns error if either proxyHost or mysqlHost not set.
func NewProxyServer(proxyHost string, mysqlHost string) (*ProxyServer, error) {
	if proxyHost == "" || mysqlHost == "" {
		return nil, ErrInvalidProxyParams
	}

	return &ProxyServer{proxyHost: proxyHost, mysqlHost: mysqlHost}, nil
}

// SetChannels assigns user defined channels to ProxyServer.
// This channels are used to transfer captured command(query), command state and
// connection state to corresponding routine.
func (ps *ProxyServer) SetChannels(
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
func (ps *ProxyServer) setCommand(connId int, cmdId int, query string) {
	ps.cmdChan <- Cmd{ConnId: connId, CmdId: cmdId, Query: query}
}

// setCommandResult writes command execution result to command result channel
// provided by caller code via NewProxyServer routine
func (ps *ProxyServer) setCommandResult(connId int, cmdId int, cmdState byte, error string, duration time.Duration) {
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
func (ps *ProxyServer) setConnectionState(connId int, state byte) {
	ps.connStateChan <- ConnState{ConnId: connId, State: state}
}

// handleConnection ...
func (ps *ProxyServer) handleConnection(connId int, conn net.Conn) {
	defer conn.Close()

	// Establishing connection to MySQL server for proxying packets
	// New connection is made per each TCP request to proxy server
	mysql, err := net.Dial("tcp", ps.mysqlHost)
	if err != nil {
		log.Print(err.Error())
		return
	}

	// Both calls to setConnectionState used to update connection state
	// on connection open and close events
	defer ps.setConnectionState(connId, connStateFinished)
	ps.setConnectionState(connId, connStateStarted)

	processHandshake(conn, mysql)

	ps.extractAndForward(conn, mysql, connId)
}

// extractAndForward reads data from proxy client, extracts queries and forwards them to MySQL
func (ps *ProxyServer) extractAndForward(conn net.Conn, mysql net.Conn, connId int) {
	var cmdId int
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
			ps.setCommand(connId, cmdId, query)

			start := time.Now()
			writePacket(queryPacket, mysql)

			response, result, err := readResponse(mysql)
			if err == nil {
				if result == responseErr {
					ps.setCommandResult(connId, cmdId, result, readErrMessage(response), time.Since(start))
				} else {
					ps.setCommandResult(connId, cmdId, result, "", time.Since(start))
				}

				writePacket(response, conn)
			}

		// Received COM_STMT_PREPARE from client
		case requestCmdStmtPrepare:
			query, _ := getQueryString(queryPacket)
			ps.setCommand(connId, cmdId, query)

			start := time.Now()
			writePacket(queryPacket, mysql)

			response, result, err := readPrepareResponse(mysql)
			if err == nil {
				ps.setCommandResult(connId, cmdId, result, "", time.Since(start))

				writePacket(response, conn)
			}

		// Received COM_STMT_EXECUTE from MySQL client
		case requestCmdStmtExecute:
			writePacket(queryPacket, mysql)
			response, _, _ := readResponse(mysql)
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
func (ps *ProxyServer) Run() {
	listener, err := net.Listen("tcp", ps.proxyHost)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()

	go func() {
		ps.appReadyChan <- true
		close(ps.appReadyChan)
	}()

	var connectionId int
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err.Error())
		}

		connectionId++

		go ps.handleConnection(connectionId, conn)
	}
}
