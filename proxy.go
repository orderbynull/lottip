package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"lottip/chat"
	"lottip/protocol"
	"net"
	"strings"
	"time"
)

type ServerHandshakeV10 struct {
	ProtocolVersion byte
	ServerVersion   string
	ConnectionID    uint32
	AuthPluginData  []byte
	CapabilityFlags uint32
	CharacterSet    byte
	StatusFlags     uint16
	AuthPluginName  string
}

type ConnectionInfo struct {
	ConnId               string
	User                 string
	ClientAddress        string
	ClientPort           int
	ServerAddress        string
	ServerPort           int
	QueryId              int
	timer                *time.Time
	clientPacketFragment *[]byte
	serverPacketFragment *[]byte
	fsm                  *MySQLProtocolFSM
	serverHandshake      ServerHandshakeV10
}

func createConnectionInfo(id string, client string, clientPort int, server string, serverPort int, clientConn net.Conn, serverConn net.Conn, cmdChan chan chat.Cmd, resultChan chan chat.CmdResult, stateChan chan chat.ConnState) ConnectionInfo {
	timer := time.Now()
	ci := ConnectionInfo{}
	ci.ConnId = id
	ci.ClientAddress = client
	ci.ClientPort = clientPort
	ci.ServerAddress = server
	ci.ServerPort = serverPort
	ci.timer = &timer
	ci.clientPacketFragment = &[]byte{}
	ci.serverPacketFragment = &[]byte{}

	ci.fsm = CreateStateMachine(&ci, clientConn, serverConn, cmdChan, resultChan, stateChan)

	return ci
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

// handleConnection ...
func (p *MySQLProxyServer) handleConnection(client net.Conn) {
	defer client.Close()

	// New connection to MySQL is made per each incoming TCP request to MySQLProxyServer server.
	if !strings.Contains(p.mysqlHost, ":") {
		p.mysqlHost += ":3306"
	}
	server, err := net.Dial("tcp", p.mysqlHost)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to MySQL at " + p.mysqlHost)
		return
	}
	defer server.Close()

	connId := fmt.Sprintf("%s => %s", client.RemoteAddr().String(), server.RemoteAddr().String())

	defer func() { p.connStateChan <- chat.ConnState{connId, protocol.ConnStateFinished} }()

	clientAddress := client.RemoteAddr().String()
	clientPort := -1
	if addr, ok := client.RemoteAddr().(*net.TCPAddr); ok {
		clientAddress = addr.IP.String()
		clientPort = addr.Port
	}
	serverAddress := server.RemoteAddr().String()
	serverPort := -1
	if addr, ok := server.RemoteAddr().(*net.TCPAddr); ok {
		serverAddress = addr.IP.String()
		serverPort = addr.Port
	}

	connInfo := createConnectionInfo(connId, clientAddress, clientPort, serverAddress, serverPort, client, server, p.cmdChan, p.cmdResultChan, p.connStateChan)

	// Copy bytes from client to server and requestParser
	go io.Copy(io.Writer(&ClientToServerHandler{&connInfo, p.cmdChan, p.connStateChan, server}), client)

	// Copy bytes from server to client and responseParser
	io.Copy(io.Writer(&ServerToClientHandler{&connInfo, p.cmdResultChan, client}), server)
}

// run starts accepting TCP connection and forwarding it to MySQL server.
// Each incoming TCP connection is handled in own goroutine.
func (p *MySQLProxyServer) run() {
	listener, err := net.Listen("tcp", p.proxyHost)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not listen on TCP at " + p.proxyHost)
	}
	defer listener.Close()

	go func() {
		p.appReadyChan <- true
		close(p.appReadyChan)
	}()

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("Could not accept connection")
		}

		go p.handleConnection(client)
	}
}

// ReadLenEncode used to read variable length.
func readLenEncodedInt(packet []byte, offset uint32) (value uint64, newOffset uint32) {
	var u8 uint8
	u8 = packet[offset]

	switch u8 {
	case 0xfb:
		// nil value
		// we set the length to maxuint64.
		value = ^uint64(0)
		return value, offset + 1

	case 0xfc:
		value = uint64(packet[offset+1]) | uint64(packet[offset+2])<<8
		return value, offset + 3

	case 0xfd:
		value = uint64(packet[offset+1]) | uint64(packet[offset+2])<<8 | uint64(packet[offset+3])<<16
		return value, offset + 4

	case 0xfe:
		value = uint64(packet[offset]) | uint64(packet[offset+1])<<8 |
			uint64(packet[offset+2])<<16 | uint64(packet[offset+3])<<24 |
			uint64(packet[offset+4])<<32 | uint64(packet[offset+5])<<40 |
			uint64(packet[offset+6])<<48 | uint64(packet[offset+7])<<56
		return value, offset + 8

	default:
		return uint64(u8), offset + 1
	}
}

func GetPacketType(p []byte) byte {
	return p[4]
}
