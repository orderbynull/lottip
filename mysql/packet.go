package mysql

import "errors"

//ErrMalformedPacket means packet is malformed or cannot be parsed via selected function
var ErrMalformedPacket = errors.New("malformed packet")

//ErrUnknownPacket means packet type cannot be detected
var ErrUnknownPacket = errors.New("unknown packet")

//GetResponsePktType returns response packet type without parsing it into struct
func GetResponsePktType(pkt []byte) (byte, error) {
	if len(pkt) < 4 {
		return 0, ErrMalformedPacket
	}

	switch pkt[4] {
	case
		ResponseOkPacket,
		ResponseErrPacket:
		return pkt[4], nil
	}

	return 0, ErrUnknownPacket
}

//ComQueryPkt represents COM_QUERY request packet
type ComQueryPkt struct {
	Query string
}

//ParseComQuery extracts sql query from COM_QUERY command
func ParseComQuery(pkt []byte) (*ComQueryPkt, error) {
	if len(pkt) < 6 {
		return nil, ErrMalformedPacket
	}

	if pkt[4] != ComQuery {
		return nil, ErrMalformedPacket
	}

	return &ComQueryPkt{Query: string(pkt[5:])}, nil
}

//OkPkt represents response OK_Packet
type OkPkt struct {
	AffectedRows int
}

//ParseOk parses OK_Packet
func ParseOk(pkt []byte) (*OkPkt, error) {
	if len(pkt) < 7 {
		return nil, ErrMalformedPacket
	}

	if pkt[4] != ResponseOkPacket {
		return nil, ErrMalformedPacket
	}

	return &OkPkt{AffectedRows: int(pkt[5])}, nil
}

//ErrPkt represents response ERR_Packet
type ErrPkt struct {
	ErrorCode    int
	ErrorMessage string
}

//ParseErr parses ERR_Packet
func ParseErr(pkt []byte) (*ErrPkt, error) {
	if len(pkt) < 13 {
		return nil, ErrMalformedPacket
	}

	if pkt[4] != ResponseErrPacket {
		return nil, ErrMalformedPacket
	}

	return &ErrPkt{ErrorCode: int(uint32(pkt[5]) | uint32(pkt[6])<<8), ErrorMessage: string(pkt[13:])}, nil
}
