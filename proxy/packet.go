package proxy

import (
	"encoding/binary"
	"io"
	"net"
)

// processHandshake handles handshake between client and MySQL server.
// When client connects MySQL server for the first time "handshake"
// packet is sent by MySQL server so it just should be delivered without analyzing.
func processHandshake(app net.Conn, mysql net.Conn) {
	proxyPacket(mysql, app)
}

// readPrepareResponse reads response from MySQL server for COM_STMT_PREPARE
// query issued by client.
// ...
func readPrepareResponse(conn net.Conn) ([]byte, byte, error) {
	pkt, err := readPacket(conn)
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
			pkt, err = readPacket(conn)
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

// readResponse ...
func readResponse(conn net.Conn) ([]byte, byte, error) {

	pkt, err := readPacket(conn)
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

	var eofCnt int
	var data []byte

	data = append(data, pkt...)

	for {
		pkt, err := readPacket(conn)
		if err != nil {
			return []byte{}, 0, err
		}

		data = append(data, pkt...)

		if pkt[4] == responseEof {
			eofCnt++
		}

		if eofCnt == 2 {
			break
		}
	}

	return data, responseResultset, nil
}

// readPacket ...
func readPacket(conn net.Conn) ([]byte, error) {

	header := []byte{0, 0, 0, 0}

	if _, err := io.ReadFull(conn, header); err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, err
	}

	bodyLength := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	body := make([]byte, bodyLength)

	n, err := io.ReadFull(conn, body)
	if err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, err
	}

	return append(header, body[0:n]...), nil
}

// writePacket ...
func writePacket(pkt []byte, conn net.Conn) (int, error) {
	n, err := conn.Write(pkt)
	if err != nil {
		return 0, ErrWritePacket
	}

	return n, nil
}

// proxyPacket ...
func proxyPacket(src, dst net.Conn) ([]byte, error) {
	pkt, err := readPacket(src)
	if err != nil {
		return nil, err
	}

	_, err = writePacket(pkt, dst)
	if err != nil {
		return nil, err
	}

	return pkt, nil
}

// getQueryString ...
func getQueryString(pkt []byte) (string, error) {
	if len(pkt) > 5 {
		return string(pkt[5:]), nil
	}

	return "", ErrNoQueryPacket
}