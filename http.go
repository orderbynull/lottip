package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/olekukonko/tablewriter"
)

const (
	websocketRoute = "/ws"
	webRoute       = "/"
)

func runHttpServer(hub *Hub) {

	// Websockets endpoint
	http.HandleFunc(websocketRoute, func(w http.ResponseWriter, r *http.Request) {
		upgr := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		conn, err := upgr.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		//Proper handling 'close' message from the peer
		//https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages
		go func() {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
			}
		}()

		client := newClient(conn, hub)

		hub.registerClient(client)

		go client.Process()
	})

	// Query execution endpoint
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}

		query := r.PostFormValue("query")

		columns, rows, err := getQueryResults(query)
		if err != nil {
			log.Println(err)
			return
		}

		if len(columns) > 0 {
			table := tablewriter.NewWriter(w)
			table.SetHeader(columns)
			table.AppendBulk(rows)
			table.Render()
		} else {
			fmt.Fprint(w, "Got empty response")
		}
	})

	http.Handle(webRoute, http.FileServer(FS(*useLocalUI)))

	log.Fatal(http.ListenAndServe(*guiAddr, nil))
}
