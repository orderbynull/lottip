package main

import (
	"fmt"
	"github.com/pubnative/mysqlproto-go"
	"io"
	"log"
	"lottip/chat"
	"lottip/protocol"
	"net"
	"reflect"
	"time"
)

const dateFormat = "2006-01-02 15:04:05.000"

func logEntry(info *ConnectionInfo, entryType string, args ...interface{}) {
	prefix := fmt.Sprintf("%s [%s][%s][%s][%d][%s]", time.Now().Format(dateFormat), info.client, info.server, info.user, info.queryId, entryType)
	if len(args) > 0 {
		if len(args) == 1 {
			if reflect.TypeOf(args[0]).Kind() == reflect.String {
				prefix += " " + args[0].(string)
			}
		} else {
			prefix += " " + fmt.Sprintf(args[0].(string), args[1:]...)
		}
	}

	fmt.Println(prefix)
}

type RequestPacketParser struct {
	connInfo      *ConnectionInfo
	queryChan     chan chat.Cmd
	connStateChan chan chat.ConnState
	server        io.Writer
}

func (pp *RequestPacketParser) Write(p []byte) (n int, err error) {
	//fmt.Printf(">> ")
	//for i, b := range p {
	//	fmt.Printf("%02x ", b)
	//	if i > 12 {
	//		break
	//	}
	//}
	//fmt.Println()
	// Switch based on the state
	switch pp.connInfo.connectionState {
	case ConnStateInit:
		fmt.Printf("Initializing connection -- we shouldn't be sending anything here")
		pp.connStateChan <- chat.ConnState{pp.connInfo.connId, protocol.ConnStateFinished}
		return
	case ConnStateAuthInfoRequested:
		// Login request
		start := 3 + 1 + 2 + 2 + 4 + 1
		for ; p[start] == 0; start++ {
		}
		stop := start + 1
		for ; p[stop] != 0; stop++ {
		}
		pp.connInfo.user = string(p[start:stop])

		logEntry(pp.connInfo, "Login", "Authorizing as user: '%s' (forcing uncompressed protocol)", pp.connInfo.user)

		// Disable compression
		p[4] = p[4] & 0xDF
		pp.connInfo.connectionState = ConnStateAuthInfoSent
	case ConnStateUnauthorized:
		logEntry(pp.connInfo, "ERROR", "Unauthorized state - can't send commands/queries")
		pp.connStateChan <- chat.ConnState{pp.connInfo.connId, protocol.ConnStateFinished}
		return
	case ConnStateAuthorized:
		if len(p) < 4 {
			// ERROR
			fmt.Printf("[Request] connId: `%s`, queryId: `%d`, packetType: %d\n", pp.connInfo.connId, pp.connInfo.queryId, int(getPacketType(p)))
			fmt.Printf(">> ")
			for _, b := range p {
				fmt.Printf("%02x ", b)
			}
			fmt.Println()
			for _, b := range p {
				fmt.Printf("%c", b)
			}
			fmt.Println()

		} else if getPacketType(p) == protocol.Ping {
			// Filter out pings
			logEntry(pp.connInfo, "Ping")
		} else {
			// We are processing queries
			pp.connInfo.queryId++
			*pp.connInfo.timer = time.Now()

			switch getPacketType(p) {
			case protocol.ComStmtPrepare:
				query := string(p[5:])
				logEntry(pp.connInfo, "Prepare", query)
				pp.queryChan <- chat.Cmd{pp.connInfo.connId, pp.connInfo.queryId, "", query, nil, false}
			case protocol.ComQuery:
				query := string(p[5:])
				logEntry(pp.connInfo, "Query", query)
				pp.queryChan <- chat.Cmd{pp.connInfo.connId, pp.connInfo.queryId, "", query, nil, false}
				pp.connInfo.connectionState = ConnStateQueryFields
			case protocol.ComQuit:
				logEntry(pp.connInfo, "Quit")
				pp.connStateChan <- chat.ConnState{pp.connInfo.connId, protocol.ConnStateFinished}
			}
		}
	}

	return pp.server.Write(p)
}

func getPacketType(p []byte) byte {
	return p[4]
}

type ResponsePacketParser struct {
	connInfo        *ConnectionInfo
	queryResultChan chan chat.CmdResult
	client          io.Writer
}

func (pp *ResponsePacketParser) Write(p []byte) (n int, err error) {
	duration := fmt.Sprintf("%.3f", time.Since(*pp.connInfo.timer).Seconds())

	// Switch based on the state
	switch pp.connInfo.connectionState {
	case ConnStateInit:
		// Server's initial response asking for loging info -- passthrough
		p[25] = p[25] & 0xDF
		logEntry(pp.connInfo, "Connect", "Client connecting - Server Handshake/Challenge response (forcing uncompressed protocol)")
		pp.connInfo.connectionState = ConnStateAuthInfoRequested

	case ConnStateAuthInfoSent:
		pp.processPacket(p, func() {
			pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseOk, "", duration}
			logEntry(pp.connInfo, "Login:OK")
			pp.connInfo.connectionState = ConnStateAuthorized
		}, func(errCode int32, errMessage string) {
			pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseErr, fmt.Sprintf("Code: %d => %s", errCode, errMessage), duration}
			logEntry(pp.connInfo, "Login:ERROR", "Code: %d => %s", errCode, errMessage)
			pp.connInfo.connectionState = ConnStateUnauthorized
		}, func(seqNumber int) {
			logEntry(pp.connInfo, "Login:EOF", "Packet %d", seqNumber)
			pp.connInfo.connectionState = ConnStateUnauthorized
		})

	case ConnStateAuthorized:
		pp.processPacket(p, func() {
			logEntry(pp.connInfo, "OK")
		}, func(errCode int32, errMessage string) {
			logEntry(pp.connInfo, "ERROR", "Code: %d => %s", errCode, errMessage)
		}, func(seqNumber int) {
			logEntry(pp.connInfo, "EOF", "Packet %d", seqNumber)
		})

	case ConnStateQueryFields:
		pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseOk, "", duration}
		pp.processPacket(p, func() {
			logEntry(pp.connInfo, "Query:OK")
		}, func(errCode int32, errMessage string) {
			logEntry(pp.connInfo, "Query:ERR", "Code: %d => %s", errCode, errMessage)
			pp.connInfo.connectionState = ConnStateAuthorized
		}, func(seqNumber int) {
			logEntry(pp.connInfo, "Query:EOF", "Packet %d", seqNumber)
			pp.connInfo.connectionState = ConnStateQueryRows
		})

	case ConnStateQueryRows:
		pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseOk, "", duration}
		pp.processPacket(p, func() {
			logEntry(pp.connInfo, "Query:OK")
		}, func(errCode int32, errMessage string) {
			logEntry(pp.connInfo, "Query:ERR", "Code: %d => %s", errCode, errMessage)
			pp.connInfo.connectionState = ConnStateAuthorized
		}, func(seqNumber int) {
			logEntry(pp.connInfo, "Query:EOF", "Packet %d", seqNumber)
			pp.connInfo.connectionState = ConnStateAuthorized
		})
	}

	return pp.client.Write(p)
}

func (pp *ResponsePacketParser) processPacket(incomingPacket []byte, okPacket func(), errorPacket func(errCode int32, errMessage string), eofPacket func(seqNumber int)) {
	var packet []byte
	if len(pp.connInfo.packetFragment) > 0 {
		// logEntry(pp.connInfo, "INFO", "Packet fragment %v", pp.connInfo.packetFragment)
		packet = append(pp.connInfo.packetFragment, incomingPacket...)
		pp.connInfo.packetFragment = []byte{}
	} else {
		packet = incomingPacket
	}

	offset := uint32(0)
	bufferLen := uint32(len(packet))
	for {
		if bufferLen == offset {
			// Nothing else
			break
		} else if offset < bufferLen && bufferLen-offset >= 4 {
			packetSize := uint32(packet[offset+0]) | uint32(packet[offset+1])<<8 | uint32(packet[offset+2])<<16
			if bufferLen >= offset+packetSize+4 {
				temp := packet[offset : offset+3+1+packetSize]
				seqNum := int(temp[3])
				if temp[4] == mysqlproto.OK_PACKET && temp[5] == 0 {
					okPacket()
				} else if temp[4] == mysqlproto.ERR_PACKET {
					errCode := int32(temp[5]) | int32(temp[6])<<8
					errorPacket(errCode, string(temp[7:]))
				} else if temp[4] == mysqlproto.EOF_PACKET {
					eofPacket(seqNum)
				}
				offset += packetSize + 4
				continue
			} else {
				//	fmt.Printf("Not enough data for next packet\n")
			}
		} else {
			//fmt.Printf("Not enough data for next packet header\n")
		}

		if bufferLen-offset > 0 {
			//fmt.Printf("<< ")
			//fmt.Printf("bufferLen: %d, offset: %d | ", bufferLen, offset)
			//for _, b := range packet[offset:bufferLen] {
			//	fmt.Printf("%02x ", b)
			//}
			//fmt.Println()
			// If we reach here, that means we need more bytes
			pp.connInfo.packetFragment = make([]byte, bufferLen-offset)
			copy(pp.connInfo.packetFragment, packet[offset:bufferLen])
			// logEntry(pp.connInfo, "INFO", "Will wait for next TCP packet since %d bytes are remaining %v", len(pp.connInfo.packetFragment), pp.connInfo.packetFragment)
		} else {
			// We are done
			pp.connInfo.packetFragment = []byte{}
		}
		break
	}
}

type ConnState int

const (
	ConnStateInit ConnState = iota + 1
	ConnStateAuthInfoRequested
	ConnStateAuthInfoSent
	ConnStateAuthorized
	ConnStateQueryFields
	ConnStateQueryRows
	ConnStateUnauthorized
)

// ConnectionInfo stores connection level details
type ConnectionInfo struct {
	connId          string
	connectionState ConnState
	user            string
	client          string
	server          string
	queryId         int
	idle            bool
	timer           *time.Time
	packetFragment  []byte
}

func createConnectionInfo(id string, client string, server string) ConnectionInfo {
	timer := time.Now()
	return ConnectionInfo{id, ConnStateInit, "", client, server, 0, true, &timer, []byte{}}
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

	connInfo := createConnectionInfo(connId, client.RemoteAddr().String(), server.RemoteAddr().String())

	// Copy bytes from client to server and requestParser
	go io.Copy(io.Writer(&RequestPacketParser{&connInfo, p.cmdChan, p.connStateChan, server}), client)

	// Copy bytes from server to client and responseParser
	io.Copy(io.Writer(&ResponsePacketParser{&connInfo, p.cmdResultChan, client}), server)
}
