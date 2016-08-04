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

	_, err := io.ReadFull(left, header)
	if err != nil {
		return nil, fmt.Errorf("Error while reading packet header: %s", err.Error())
	}

	bodyLength := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	body := make([]byte, bodyLength)

	bn, err := io.ReadFull(left, body)
	if err != nil {
		return nil, fmt.Errorf("Error while reading packet body: %s", err.Error())
	}

	return &Packet{Payload: append(header, body[0:bn]...), Type: body[0], Query: string(body[1:bn])}, nil
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
