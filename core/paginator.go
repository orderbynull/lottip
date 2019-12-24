package core

import (
	"fmt"
	"net/http"
)

type Paginator struct {
	NewerUrl string
	OlderUrl string

}

func NewPaginator(request *http.Request, rows []PgsqlPacket) *Paginator {
	var lastRowId int
	var firstRowId int

	if len(rows) > 0 {
		lastRowId = rows[len(rows)-1].Id
	}
	if len(rows) > 0 {
		firstRowId = rows[0].Id
	}

	return &Paginator{
		fmt.Sprintf("%s?newer=%d", request.URL.Path, firstRowId),
		fmt.Sprintf("%s?older=%d", request.URL.Path, lastRowId),
	}
}
