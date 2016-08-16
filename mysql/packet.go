package mysql

import "errors"

//ErrMalformedPacket means packet is malformed or cannot be parsed via selected function
var ErrMalformedPacket = errors.New("Malformed packet")

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
	LastInsertID int
}

//ParseOk parses OK_Packet
func ParseOk(pkt []byte) (*OkPkt, error) {
	if len(pkt) < 7 {
		return nil, ErrMalformedPacket
	}

	if pkt[4] != ResponseOkPacket {
		return nil, ErrMalformedPacket
	}

	return &OkPkt{AffectedRows: int(pkt[5]), LastInsertID: int(pkt[6])}, nil
}
