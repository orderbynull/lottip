package proxy

import (
	"encoding/binary"
	"io"
	"net"
)

//...
type ConnSettings struct {
	clientCapabilities uint32
	serverCapabilities uint32
	SelectedDb         string
}

//...
func (h *ConnSettings) DeprecateEOFSet() bool {
	return ((clientDeprecateEOF & h.serverCapabilities) != 0) &&
		((clientDeprecateEOF & h.clientCapabilities) != 0)
}

// processHandshake handles ConnSettings between client and MySQL server.
// When client connects MySQL server for the first time "ConnSettings"
// packet is sent by MySQL server so it just should be delivered without analyzing.
// Returns extended server and client capabilities flags
func processHandshake(app net.Conn, mysql net.Conn) *ConnSettings {
	serverPacket, _ := ProxyPacket(mysql, app)
	clientPacket, _ := ProxyPacket(app, mysql)
	ProxyPacket(mysql, app)

	return &ConnSettings{
		clientCapabilities: uint32(binary.LittleEndian.Uint16(clientPacket[6 : 6+2])),
		serverCapabilities: uint32(binary.LittleEndian.Uint16(serverPacket[30 : 30+2])),
	}
}

// readPrepareResponse reads response from MySQL server for COM_STMT_PREPARE
// query issued by client.
// ...
func readPrepareResponse(conn net.Conn) ([]byte, byte, error) {
	pkt, err := ReadPacket(conn)
	if err != nil {
		return []byte{}, 0, err
	}

	numParams := binary.LittleEndian.Uint16(pkt[9:11])
	numColumns := binary.LittleEndian.Uint16(pkt[11:13])
	packetsExpected := 0

	if numParams > 0 {
		packetsExpected += int(numParams) + 1
	}

	if numColumns > 0 {
		packetsExpected += int(numColumns) + 1
	}

	switch pkt[4] {
	case responsePrepareOk:
		var data []byte
		var eofCnt int

		data = append(data, pkt...)

		for i := 1; i <= packetsExpected; i++ {
			eofCnt++
			pkt, err = ReadPacket(conn)
			if err != nil {
				return []byte{}, 0, err
			}

			data = append(data, pkt...)
		}

		return data, responseOk, nil

	case responseErr:
		return pkt, responseErr, nil
	}

	return []byte{}, 0, nil
}

func readErrMessage(errPacket []byte) string {
	return string(errPacket[13:])
}

func readShowFieldsResponse(conn net.Conn) ([]byte, byte, error) {
	return ReadResponse(conn, true)
}

// ReadResponse ...
func ReadResponse(conn net.Conn, deprecateEof bool) ([]byte, byte, error) {
	pkt, err := ReadPacket(conn)
	if err != nil {
		return []byte{}, 0, err
	}

	switch pkt[4] {
	case responseOk:
		return pkt, responseOk, nil

	case responseErr:
		return pkt, responseErr, nil

	case responseLocalinfile:
	}

	var data []byte

	data = append(data, pkt...)

	if !deprecateEof {
		columns, _ := ReadLenEncodedInteger(pkt[4:])

		toRead := int(columns) + 1

		for i := 0; i < toRead; i++ {
			pkt, err := ReadPacket(conn)
			if err != nil {
				return []byte{}, 0, err
			}

			data = append(data, pkt...)
		}
	}

	for {
		pkt, err := ReadPacket(conn)
		if err != nil {
			return []byte{}, 0, err
		}

		data = append(data, pkt...)

		if pkt[4] == responseEof {
			break
		}
	}

	return data, responseResultset, nil
}

// ReadPacket ...
func ReadPacket(conn net.Conn) ([]byte, error) {

	// Read packet header
	header := []byte{0, 0, 0, 0}
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	// Calculate packet body length
	bodyLen := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	// Read packet body
	body := make([]byte, bodyLen)
	n, err := io.ReadFull(conn, body)
	if err != nil {
		return nil, err
	}

	return append(header, body[0:n]...), nil
}

// WritePacket ...
func WritePacket(pkt []byte, conn net.Conn) (int, error) {
	n, err := conn.Write(pkt)
	if err != nil {
		return 0, errWritePacket
	}

	return n, nil
}

// ProxyPacket ...
func ProxyPacket(src, dst net.Conn) ([]byte, error) {
	pkt, err := ReadPacket(src)
	if err != nil {
		return nil, err
	}

	_, err = WritePacket(pkt, dst)
	if err != nil {
		return nil, err
	}

	return pkt, nil
}
