package mysql

import "fmt"

//ComQueryPkt represents COM_QUERY command packet
type ComQueryPkt struct {
	Query string
}

//ParseComQuery extracts sql query from COM_QUERY command
func ParseComQuery(pkt []byte) (*ComQueryPkt, error) {
	if pkt[4] != ComQuery {
		return nil, fmt.Errorf("Cannot parse COM_QUERY")
	}

	return &ComQueryPkt{Query: string(pkt[5:])}, nil
}

//OkPkt represents OK_Packet
type OkPkt struct {
	AffectedRows int
	LastInsertID int
}

//ParseOk parses OK_Packet
func ParseOk(pkt []byte) (*OkPkt, error) {
	if pkt[4] != ResponseOkPacket {
		return nil, fmt.Errorf("Cannot parse OK_Packet")
	}

	return &OkPkt{AffectedRows: int(pkt[5]), LastInsertID: int(pkt[6])}, nil
}
