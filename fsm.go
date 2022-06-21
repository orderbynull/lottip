package main

import (
	"context"
	"fmt"
	"github.com/qmuntal/stateless"
	"lottip/chat"
	"lottip/protocol"
	"net"
	"reflect"
	"time"
)

type MySQLProtocolFSM struct {
	*stateless.StateMachine
	connectionInfo *ConnectionInfo
	clientConn     net.Conn
	serverConn     net.Conn
}

const (
	StateIdle          = "Idle"
	StateAuthRequested = "AuthRequested"
	StateAuthSent      = "AuthSent"
	StateAuthorized    = "stateAuthorized"
	StateUnauthorized  = "stateUnauthorized"
)

const (
	PacketReceived    = "PacketReceived"
	MsgServerHello    = "ServerHello"
	MsgLogin          = "Login"
	MsgOK             = "OK"
	MsgERROR          = "ERROR"
	MsgServerToClient = "MsgServerToClient"
)

func CreateStateMachine(ci *ConnectionInfo, clientConn net.Conn, serverConn net.Conn, cmdChan chan chat.Cmd, resultChan chan chat.CmdResult, stateChan chan chat.ConnState) *MySQLProtocolFSM {
	fsm := stateless.NewStateMachine(StateIdle)

	fsm.SetTriggerParameters(PacketReceived, reflect.TypeOf([]byte{}))
	fsm.OnUnhandledTrigger(func(ctx context.Context, state stateless.State, trigger stateless.Trigger, unmetGuards []string) error {
		LogOther(ci, "fsm - Unhandled event", "%d in state %s", trigger, state)
		return nil
	})

	fsm.Configure(StateIdle).
		Permit(MsgServerHello, StateAuthRequested, func(ctx context.Context, args ...interface{}) bool {
			// Server's initial response asking for loging info -- passthrough after changing compression flag
			packet := args[0].([]byte)
			packet[25] = packet[25] & 0xDF
			LogResponse(ci, "Handshake", "Server Handshake/Challenge response (forcing uncompressed protocol)")

			//LogOther(ci, "fsm - Writing to client", "% x", packet)
			clientConn.Write(packet)
			return true
		})

	fsm.Configure(StateAuthRequested).
		Permit(MsgLogin, StateAuthSent, func(ctx context.Context, args ...interface{}) bool {
			// Client's Auth Info
			packet := args[0].([]byte)

			start := 3 + 1 + 2 + 2 + 4 + 1
			for ; packet[start] == 0; start++ {
			}
			stop := start + 1
			for ; packet[stop] != 0; stop++ {
			}
			ci.User = string(packet[start:stop])
			LogRequest(ci, "Login", "Authorizing as user: '%s' (forcing uncompressed protocol)", ci.User)

			// Disable compression
			packet[4] = packet[4] & 0xDF

			//LogOther(ci, "fsm - Writing to server", "% x", packet)
			serverConn.Write(packet)
			return true
		})

	fsm.Configure(StateAuthSent).
		Permit(MsgOK, StateAuthorized, func(ctx context.Context, args ...interface{}) bool {
			LogResponse(ci, "Authorized", "Client auth successful")
			packet := args[0].([]byte)
			//LogOther(ci, "fsm - Writing to client", "% x", packet)
			clientConn.Write(packet)
			return true
		}).
		Permit(MsgERROR, StateUnauthorized, func(ctx context.Context, args ...interface{}) bool {
			// Server's initial response asking for loging info -- passthrough after changing compression flag
			packet := args[0].([]byte)
			LogResponse(ci, "Unauthorized", "Client auth failed")
			//LogOther(ci, "fsm - Writing to client", "% x", packet)
			clientConn.Write(packet)
			return true
		})

	fsm.Configure(StateAuthorized).
		InternalTransition(protocol.ComChangeUser, func(ctx context.Context, args ...interface{}) error {
			// Do not allow this!
			packet := args[0].([]byte)
			start := 3 + 1 + 2 + 2 + 4 + 1
			for ; packet[start] == 0; start++ {
			}
			stop := start + 1
			for ; packet[stop] != 0; stop++ {
			}
			LogRequest(ci, "ChangeUser", "Will reject request to change user to %s", string(packet[start:stop]))
			return nil
		}).
		InternalTransition(protocol.ComPing, func(ctx context.Context, args ...interface{}) error {
			packet := args[0].([]byte)
			LogOther(ci, "Ping")
			serverConn.Write(packet)
			return nil
		}).
		InternalTransition(protocol.ComCreateDB, logAndSendQueryToServer(ci, serverConn, "CreateDB", cmdChan)).
		InternalTransition(protocol.ComDropDB, logAndSendQueryToServer(ci, serverConn, "DropDB", cmdChan)).
		InternalTransition(protocol.ComShutdown, logAndDrop(ci, "Shutdown", cmdChan)).
		InternalTransition(protocol.ComProcessKill, logAndSendQueryToServer(ci, serverConn, "ProcessKill", cmdChan)).
		InternalTransition(protocol.ComQuery, logAndSendQueryToServer(ci, serverConn, "Query", cmdChan)).
		InternalTransition(protocol.ComStmtPrepare, logAndSendQueryToServer(ci, serverConn, "StmtPrepare", cmdChan)).
		InternalTransition(protocol.ComStmtExecute, logAndSendQueryToServer(ci, serverConn, "StmtExecute", cmdChan)).
		InternalTransition(protocol.ComStmtClose, logAndSendQueryToServer(ci, serverConn, "StmtClose", cmdChan)).
		InternalTransition(protocol.ComStmtSendLongData, logAndSendQueryToServer(ci, serverConn, "StmtSendLongData", cmdChan)).
		InternalTransition(protocol.ComStmtReset, logAndSendQueryToServer(ci, serverConn, "StmtReset", cmdChan)).
		InternalTransition(protocol.ComQuit, logAndSendQueryToServer(ci, serverConn, "Quit", cmdChan)).
		InternalTransition(protocol.ComBinlogDump, logAndDrop(ci, "BinlogDump", cmdChan)).
		InternalTransition(protocol.ComBinlogDump, logAndDrop(ci, "BinlogDump", cmdChan)).
		InternalTransition(protocol.ComTableDump, logAndDrop(ci, "TableDump", cmdChan)).
		InternalTransition(protocol.ComConnectOut, logAndDrop(ci, "ConnectOut", cmdChan)).
		InternalTransition(protocol.ResponseOk, logAndSendPacketToServer(ci, serverConn, "Ok")).
		InternalTransition(MsgOK, func(ctx context.Context, args ...interface{}) error {
			packet := args[0].([]byte)
			LogResponse(ci, "OK")
			duration := fmt.Sprintf("%.3f", time.Since(*ci.timer).Seconds())
			resultChan <- chat.CmdResult{ci.ConnId, ci.QueryId, protocol.ResponseOk, "", duration}
			clientConn.Write(packet)
			return nil
		}).
		InternalTransition(MsgERROR, func(ctx context.Context, args ...interface{}) error {
			packet := args[0].([]byte)
			LogResponse(ci, "ERROR")
			duration := fmt.Sprintf("%.3f", time.Since(*ci.timer).Seconds())
			resultChan <- chat.CmdResult{ci.ConnId, ci.QueryId, protocol.ResponseErr, "", duration}
			clientConn.Write(packet)
			return nil
		}).
		InternalTransition(MsgServerToClient, func(ctx context.Context, args ...interface{}) error {
			packet := args[0].([]byte)
			LogResponsePacket(ci, packet)
			clientConn.Write(packet)
			return nil
		})

	return &MySQLProtocolFSM{fsm, ci, clientConn, serverConn}
}

func logAndSendQueryToServer(ci *ConnectionInfo, serverConn net.Conn, command string, cmdChan chan chat.Cmd) func(ctx context.Context, args ...interface{}) error {
	return func(ctx context.Context, args ...interface{}) error {
		packet := args[0].([]byte)
		query := string(packet[5:])
		LogRequest(ci, command, query)

		ci.QueryId++
		*ci.timer = time.Now()

		cmdChan <- chat.Cmd{ci.ConnId, ci.QueryId, "", query, nil, false}
		//LogOther(ci, "fsm - Writing to server", "% x", packet)
		serverConn.Write(packet)
		return nil
	}
}

func logAndSendPacketToServer(ci *ConnectionInfo, serverConn net.Conn, command string) func(ctx context.Context, args ...interface{}) error {
	return func(ctx context.Context, args ...interface{}) error {
		packet := args[0].([]byte)
		LogOther(ci, "Sending raw packet to server for command "+command, "% x", packet)
		serverConn.Write(packet)
		return nil
	}
}

func logAndDrop(ci *ConnectionInfo, command string, cmdChan chan chat.Cmd) func(ctx context.Context, args ...interface{}) error {
	return func(ctx context.Context, args ...interface{}) error {
		packet := args[0].([]byte)
		query := string(packet[5:])
		LogRequest(ci, "BLOCKED:"+command, query)
		cmdChan <- chat.Cmd{ci.ConnId, ci.QueryId, "", query, nil, false}
		LogOther(ci, "fsm - Dropping packet", "% x", packet)
		return nil
	}
}

func (fsm *MySQLProtocolFSM) ProcessPacket(packet []byte) {
	fsm.Fire(MsgERROR)
}
