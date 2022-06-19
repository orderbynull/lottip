package main

import (
	"github.com/pubnative/mysqlproto-go"
	"io"
	"lottip/chat"
	"math"
)

type ServerToClientHandler struct {
	connInfo        *ConnectionInfo
	queryResultChan chan chat.CmdResult
	client          io.Writer
}

// From SERVER => Client
func (pp *ServerToClientHandler) Write(buffer []byte) (n int, err error) {
	//duration := fmt.Sprintf("%.3f", time.Since(*pp.connInfo.timer).Seconds())

	extractPacketsFromBuffer(pp.connInfo, pp.connInfo.serverPacketFragment, buffer, func(packet []byte) {
		fsm := pp.connInfo.fsm
		if ok, _ := fsm.IsInState(StateIdle); ok {
			// Server's initial response asking for login info
			//pp.parseServerHandshakeV10(packet)
			//logging.LogResponse(pp.connInfo, "Connect", "Client connecting - Server Handshake/Challenge response (forcing uncompressed protocol)")
			fsm.Fire(MsgServerHello, packet)
		} else if ok, _ := fsm.IsInState(StateAuthSent); ok {
			// This must be the OK packet
			_, err := mysqlproto.ParseOKPacket(packet[4:], pp.connInfo.serverHandshake.CapabilityFlags)
			if err != nil {
				_, err := mysqlproto.ParseERRPacket(packet[4:], pp.connInfo.serverHandshake.CapabilityFlags)
				if err != nil {
					LogInvalid(pp.connInfo, "ERROR: Response is not OK or ERROR packet", packet)
				} else {
					fsm.Fire(MsgERROR, packet)
				}
			} else {
				fsm.Fire(MsgOK, packet)
			}
		} else {
			_, err := mysqlproto.ParseOKPacket(packet[4:], pp.connInfo.serverHandshake.CapabilityFlags)
			if err != nil {
				_, err := mysqlproto.ParseERRPacket(packet[4:], pp.connInfo.serverHandshake.CapabilityFlags)
				if err != nil {
					fsm.Fire(MsgServerToClient, packet)
				} else {
					fsm.Fire(MsgERROR, packet)
				}
			} else {
				fsm.Fire(MsgOK, packet)
			}
		}
	})
	//// Switch based on the state
	//case ConnStateInit:
	//// Server's initial response asking for loging info -- passthrough
	//buffer[25] = buffer[25] & 0xDF
	//logEntry(pp.connInfo, "Connect", "Client connecting - Server Handshake/Challenge response (forcing uncompressed protocol)")
	//pp.connInfo.connectionState = ConnStateAuthInfoRequested
	//
	//case ConnStateAuthInfoSent:
	//pp.extractPacketsFromBuffer(buffer, func () {
	//	pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseOk, "", duration}
	//	logResponse(pp.connInfo, "Login:OK")
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (errCode int32, errMessage string) {
	//	pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseErr, fmt.Sprintf("Code: %d => %s", errCode, errMessage), duration}
	//	logResponse(pp.connInfo, "Login:ERROR", "Code: %d => %s", errCode, errMessage)
	//	pp.connInfo.connectionState = ConnStateUnauthorized
	//}, func (seqNumber int) {
	//	logResponse(pp.connInfo, "Login:EOF", "Packet %d", seqNumber)
	//	pp.connInfo.connectionState = ConnStateUnauthorized
	//}, func (packet []byte) {
	//})
	//
	//case ConnStateAuthorized:
	//pp.extractPacketsFromBuffer(buffer, func () {
	//	logResponse(pp.connInfo, "OK")
	//}, func (errCode int32, errMessage string) {
	//	logResponse(pp.connInfo, "ERROR", "Code: %d => %s", errCode, errMessage)
	//}, func (seqNumber int) {
	//	logResponse(pp.connInfo, "EOF", "Packet %d", seqNumber)
	//}, func (packet []byte) {
	//})
	//
	//case ConnStateQueryActive:
	//pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseOk, "", duration}
	//pp.extractPacketsFromBuffer(buffer, func () {
	//	logResponse(pp.connInfo, "Query:OK")
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (errCode int32, errMessage string) {
	//	logResponse(pp.connInfo, "Query:ERR", "Code: %d => %s", errCode, errMessage)
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (seqNumber int) {
	//	logResponse(pp.connInfo, "Query:EOF", "Packet %d", seqNumber)
	//	pp.connInfo.connectionState = ConnStateQueryRows
	//}, func (packet []byte) {
	//	columns, _ := readLenEncodedInt(packet, 4)
	//	fmt.Println("Expecting", columns, "column definitions")
	//	pp.connInfo.columnCount = int(columns)
	//	pp.connInfo.connectionState = ConnStateQueryColumnDefs
	//})
	//
	//case ConnStateQueryColumnDefs:
	//pp.extractPacketsFromBuffer(buffer, func () {
	//	logResponse(pp.connInfo, "Query:OK", "Received this in the unexpectedly in the middle of column-defs")
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (errCode int32, errMessage string) {
	//	logResponse(pp.connInfo, "Query:ERR", "Code: %d => %s", errCode, errMessage)
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (seqNumber int) {
	//	logResponse(pp.connInfo, "Query:EOF", "Packet %d", seqNumber)
	//	pp.connInfo.connectionState = ConnStateQueryRows
	//}, func (packet []byte) {
	//	pp.connInfo.columnCount = pp.connInfo.columnCount - 1
	//	fmt.Println("Received a column def packet -", pp.connInfo.columnCount, "packets to go")
	//})
	//
	//case ConnStateQueryRows:
	//pp.queryResultChan <- chat.CmdResult{pp.connInfo.connId, pp.connInfo.queryId, protocol.ResponseOk, "", duration}
	//pp.extractPacketsFromBuffer(buffer, func () {
	//	logResponse(pp.connInfo, "Query:OK")
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (errCode int32, errMessage string) {
	//	logResponse(pp.connInfo, "Query:ERR", "Code: %d => %s", errCode, errMessage)
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (seqNumber int) {
	//	logResponse(pp.connInfo, "Query:EOF", "Packet %d", seqNumber)
	//	pp.connInfo.connectionState = ConnStateAuthorized
	//}, func (packet []byte) {
	//})
	//}

	//return pp.client.Write(buffer)
	return len(buffer), nil
}

func (pp *ServerToClientHandler) parseServerHandshakeV10(packet []byte) {
	_, end := getInt24(packet, 0)
	// sequenceNumber := packet[end]
	end++
	pp.connInfo.serverHandshake.ProtocolVersion = packet[end]
	end++
	start := end
	for ; packet[end] != 0; end++ {
	}
	pp.connInfo.serverHandshake.ServerVersion = string(packet[start:end])
	end++
	pp.connInfo.serverHandshake.ConnectionID, end = getInt32(packet, end)
	// Skip the plugin data
	pp.connInfo.serverHandshake.AuthPluginData = packet[end : end+7]
	end += 8
	// Skip filler
	end++
	pp.connInfo.serverHandshake.CapabilityFlags = uint32(packet[end]) | uint32(packet[end+1])<<8
	end += 2
	if end < len(packet) {
		pp.connInfo.serverHandshake.CharacterSet = packet[end]
		end++
		pp.connInfo.serverHandshake.StatusFlags = uint16(packet[end]) | uint16(packet[end+1])<<8
		end += 2
		pp.connInfo.serverHandshake.CapabilityFlags = pp.connInfo.serverHandshake.CapabilityFlags | (uint32(packet[end])|uint32(packet[end+1])<<8)<<16
		end += 2
		pluginDataLength := 0
		if pp.connInfo.serverHandshake.CapabilityFlags&mysqlproto.CLIENT_PLUGIN_AUTH == mysqlproto.CLIENT_PLUGIN_AUTH {
			pluginDataLength = int(packet[end])
		}
		end++
		// Skip reserved block
		end += 10
		if pp.connInfo.serverHandshake.CapabilityFlags&mysqlproto.CLIENT_SECURE_CONNECTION == mysqlproto.CLIENT_SECURE_CONNECTION {
			pluginDataLength := int(math.Max(13, float64(pluginDataLength)-8))
			pp.connInfo.serverHandshake.AuthPluginData = append(pp.connInfo.serverHandshake.AuthPluginData, packet[end:end+pluginDataLength]...)
			end += pluginDataLength
			if pp.connInfo.serverHandshake.CapabilityFlags&mysqlproto.CLIENT_PLUGIN_AUTH == mysqlproto.CLIENT_PLUGIN_AUTH {
				start = end
				for ; packet[end] != 0; end++ {
				}
				pp.connInfo.serverHandshake.AuthPluginName = string(packet[start:end])
			}
		}
	}
}

func getInt32(packet []byte, offset int) (uint32, int) {
	v := uint32(packet[offset]) << uint32(packet[offset+1]) << 8 << uint32(packet[offset+2]) << 16 << uint32(packet[offset+3]) << 24
	return v, offset + 4
}

func getInt24(packet []byte, offset int) (uint32, int) {
	v := uint32(packet[offset]) << uint32(packet[offset+1]) << 8 << uint32(packet[offset+2]) << 16
	return v, offset + 3
}
