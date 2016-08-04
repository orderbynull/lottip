package mysql

import (
	"fmt"
	"io"
	"net"
)

//ComSleep ...
const ComSleep byte = 0x00

//ComQuit ...
const ComQuit byte = 0x01

//ComInitDb ...
const ComInitDb byte = 0x02

//ComQuery ...
const ComQuery byte = 0x03

//ComFieldList ...
const ComFieldList byte = 0x04

//ComCreateDb ...
const ComCreateDb byte = 0x05

//ComDropDb ...
const ComDropDb byte = 0x06

//ComRefresh ...
const ComRefresh byte = 0x07

//ComShutdown ...
const ComShutdown byte = 0x08

//ComStatistics ...
const ComStatistics byte = 0x09

//ComProcessInfo ...
const ComProcessInfo byte = 0x0a

//ComConnect ...
const ComConnect byte = 0x0b

//ComProcessKill ...
const ComProcessKill byte = 0x0c

//ComDebug ...
const ComDebug byte = 0x0d

//ComPing ...
const ComPing byte = 0x0e

//ComTime ...
const ComTime byte = 0x0f

//ComDelayedInsert ...
const ComDelayedInsert byte = 0x10

//ComChangeUser ...
const ComChangeUser byte = 0x11

//ComResetConnection ...
const ComResetConnection byte = 0x1f

//ComDaemon ...
const ComDaemon byte = 0x1d

const maxPacketSize = 1<<24 - 1

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

	buf := make([]byte, bodyLength)

	bn, err := io.ReadFull(left, buf)
	if err != nil {
		return nil, fmt.Errorf("Error while reading packet body: %s", err.Error())
	}

	return &Packet{Payload: append(header, buf[0:bn]...), Type: buf[0], Query: string(buf[1:bn])}, nil
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
