package http

import (
	"fmt"
	"github.com/orderbynull/lottip/util"
	"log"
	"net/http"

	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/olekukonko/tablewriter"
	"github.com/orderbynull/lottip/chat"
)

const (
	websocketRoute = "/ws"
	webRoute       = "/"
)

func RunHttpServer(hub *chat.Hub, mysqlDsn *string, useLocalUI *bool, guiAddr *string) {
	// Websockets endpoint
	http.HandleFunc(websocketRoute, func(w http.ResponseWriter, r *http.Request) {
		upgr := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		conn, err := upgr.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// Proper handling 'close' message from the peer
		// See https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages for details
		go func() {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
			}
		}()

		client := chat.NewClient(conn, hub)

		hub.RegisterClient(client)

		go client.Process()
	})

	// Query execution endpoint
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}

		type Data struct {
			Database   string
			Query      string
			Parameters []string
		}

		var parsedData Data
		data := r.PostFormValue("data")
		json.Unmarshal([]byte(data), &parsedData)

		columns, rows, err := util.GetQueryResults(parsedData.Database, parsedData.Query, parsedData.Parameters, *mysqlDsn)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		if len(columns) > 0 {
			table := tablewriter.NewWriter(w)
			table.SetAutoFormatHeaders(false)
			table.SetColWidth(1000)
			table.SetHeader(columns)
			table.AppendBulk(rows)
			table.Render()
		} else {
			fmt.Fprint(w, "Empty response")
		}
	})

	http.Handle(webRoute, http.FileServer(util.FS(*useLocalUI)))

	log.Fatal(http.ListenAndServe(*guiAddr, nil))
}