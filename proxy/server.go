package proxy

import (
	"fmt"
	"log"
	"net"
	"time"
)

// proxyServer implements server for capturing and forwarding MySQL traffic
type proxyServer struct {
	handshakes    map[int]*ConnSettings
	cmdChan       chan Cmd
	cmdStateChan  chan CmdResult
	connStateChan chan ConnState
	appReadyChan  chan bool
	mysqlHost     string
	proxyHost     string
}

// NewProxyServer returns new proxyServer with connections params for proxy and mysql hosts.
// Returns error if either proxyHost or mysqlHost not set.
func NewProxyServer(proxyHost string, mysqlHost string) (*proxyServer, error) {
	if proxyHost == "" || mysqlHost == "" {
		return nil, errInvalidProxyParams
	}

	return &proxyServer{handshakes: make(map[int]*ConnSettings), proxyHost: proxyHost, mysqlHost: mysqlHost}, nil
}

//...
//@TODO check for existence of ConnSettings
func (ps *proxyServer) getHandshake(connID int) *ConnSettings {
	return ps.handshakes[connID]
}

//...
func (ps *proxyServer) setHandshake(connID int, handshake *ConnSettings) {
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
func (ps *proxyServer) setCommand(
	connID int,
	cmdID int,
	database string,
	query string,
	params []PreparedParameter,
	executable bool,
) {
	var parametersSlice []string
	for _, parameter := range params {
		parametersSlice = append(parametersSlice, parameter.Value)
	}

	ps.cmdChan <- Cmd{
		ConnId:     connID,
		CmdId:      cmdID,
		Database:   database,
		Query:      query,
		Executable: executable,
		Parameters: parametersSlice,
	}
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
	var preparedQuery string
	var preparedParamsCount uint16
	var deprecateEof = ps.getHandshake(connID).DeprecateEOFSet()

	for {
		//Client query --> $requestPacket --> mysql
		requestPacket, err := ReadPacket(conn)
		if err != nil {
			break
		}

		cmdId++

		// There're packets which have zero length payload
		// and there's no need to analyze such packets.
		if len(requestPacket) < 5 {
			WritePacket(requestPacket, mysql)
			pkt, _ := ReadPacket(mysql)
			WritePacket(pkt, conn)

			continue
		}

		switch requestPacket[4] {

		// Received COM_QUERY from client
		case requestComQuery:
			decoded, _ := DecodeQueryRequest(requestPacket)

			selectedDb := haventYetDecidedFuncName(decoded.Query)
			if len(selectedDb) > 0 {
				ps.getHandshake(connID).SelectedDb = selectedDb
			}

			ps.setCommand(
				connID,
				cmdId,
				ps.getHandshake(connID).SelectedDb,
				decoded.Query,
				[]PreparedParameter{},
				true,
			)

			start := time.Now()
			WritePacket(requestPacket, mysql)

			response, result, err := ReadResponse(mysql, deprecateEof)
			if err == nil {
				if result == responseErr {
					ps.setCommandResult(connID, cmdId, result, readErrMessage(response), time.Since(start))
				} else {
					ps.setCommandResult(connID, cmdId, result, "", time.Since(start))
				}

				WritePacket(response, conn)
			}

		// Received COM_STMT_PREPARE from client
		case requestComStmtPrepare:
			decoded, _ := DecodeQueryRequest(requestPacket)

			selectedDb := haventYetDecidedFuncName(decoded.Query)
			if len(selectedDb) > 0 {
				ps.getHandshake(connID).SelectedDb = selectedDb
			}

			WritePacket(requestPacket, mysql)

			response, _, err := readPrepareResponse(mysql)
			if err == nil {
				decodedResponse, err := DecodeComStmtPrepareOkResponse(response)
				if err == nil {
					preparedParamsCount = decodedResponse.ParametersNum
					preparedQuery = decoded.Query
				}

				WritePacket(response, conn)
			}

		// Received requestComStmtExecute from MySQL client
		case requestComStmtExecute:
			var preparedParameters []PreparedParameter
			var executable bool

			decodedRequest, err := DecodeComStmtExecuteRequest(requestPacket, preparedParamsCount)
			if err == nil {
				preparedParameters = decodedRequest.PreparedParameters
				executable = true
			}

			ps.setCommand(
				connID,
				cmdId,
				ps.getHandshake(connID).SelectedDb,
				preparedQuery,
				preparedParameters,
				executable,
			)

			start := time.Now()

			WritePacket(requestPacket, mysql)
			response, result, _ := ReadResponse(mysql, deprecateEof)
			ps.setCommandResult(connID, cmdId, result, "", time.Since(start))
			WritePacket(response, conn)

		case requestComShowFields:
			WritePacket(requestPacket, mysql)
			response, _, _ := readShowFieldsResponse(mysql)
			WritePacket(response, conn)

		// Received COM_STMT_CLOSE from client
		case requestComStmtClose:
			continue

		default:
			WritePacket(requestPacket, mysql)
			pkt, _ := ReadPacket(mysql)
			WritePacket(pkt, conn)
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
