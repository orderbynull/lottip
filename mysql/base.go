package mysql

import (
	"errors"
	"io"
	"net"
)

//ErrUnknownPacket means packet type cannot be detected
var ErrWritePacket = errors.New("error while writing packet payload")

//ReadPacket reads MySQL packet into Packet struct
func ReadPacket(conn net.Conn) ([]byte, error) {
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

//WritePacket writes packet to connection
func WritePacket(pkt []byte, conn net.Conn) (int, error) {
	n, err := conn.Write(pkt)
	if err != nil {
		return 0, ErrWritePacket
	}

	return n, nil
}

//ProxyPacket is a shortcut for ReadPacket + WritePacket
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
