package mysql

import (
	"fmt"
	"io"
	"net"
)

//Packet represents MySQL packet
type Packet struct {
	Payload []byte
	Type    byte
	Query   string
}

//ReadPacket reads MySQL packet into Packet struct
func ReadPacket(left net.Conn) (*Packet, error) {
	header := []byte{0, 0, 0, 0}

	if _, err := io.ReadFull(left, header); err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, err
	}

	bodyLength := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	body := make([]byte, bodyLength)

	n, err := io.ReadFull(left, body)
	if err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, err
	}

	return &Packet{Payload: append(header, body[0:n]...), Type: body[0], Query: string(body[1:n])}, nil
}

//WritePacket writes packet to connection
func WritePacket(pkt *Packet, conn net.Conn) (int, error) {
	n, err := conn.Write(pkt.Payload)
	if err != nil {
		return 0, fmt.Errorf("Error while writing packet payload: %s", err.Error())
	}

	return n, nil
}

//ProxyPacket is a shortcut for ReadPacket and then WritePacket
func ProxyPacket(left, right net.Conn) (*Packet, error) {
	packet, err := ReadPacket(left)
	if err != nil {
		return nil, err
	}

	_, err = WritePacket(packet, right)
	if err != nil {
		return nil, err
	}

	return packet, nil
}
