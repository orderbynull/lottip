package mysql

import "fmt"

//ComQueryPkt represents COM_QUERY command
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
