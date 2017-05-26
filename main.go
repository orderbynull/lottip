package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orderbynull/lottip/proxy"
	"github.com/orderbynull/lottip/pubsub"
	"github.com/orderbynull/lottip/static"
)

const (
	wsRoute     = "/ws"
	staticRoute = "/"
)

var (
	proxyAddr  = flag.String("proxy", "127.0.0.1:4041", "Proxy <host>:<port>")
	mysqlAddr  = flag.String("mysql", "127.0.0.1:3306", "MySQL <host>:<port>")
	guiAddr    = flag.String("gui", "127.0.0.1:9999", "Web UI <host>:<port>")
	useLocalUI = flag.Bool("use-local", false, "Use local UI instead of embed")
)

func runWsServer(hub *pubsub.Hub) {
	http.HandleFunc(wsRoute, func(w http.ResponseWriter, r *http.Request) {
		upgr := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		conn, err := upgr.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		//Proper handling 'close' message from the peer
		//https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages
		go func() {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
			}
		}()

		client := pubsub.NewClient(conn, hub)

		hub.RegisterClient(client)

		go client.Process()
	})

	log.Fatal(http.ListenAndServe(*guiAddr, nil))
}

func runStaticServer() {
	http.Handle(staticRoute, http.FileServer(static.FS(*useLocalUI)))
}

func appReadyInfo(appReadyChan chan bool) {
	<-appReadyChan
	time.Sleep(1 * time.Second)
	fmt.Printf("Forwarding queries from `%s` to `%s` \n", *proxyAddr, *mysqlAddr)
	fmt.Printf("Web gui available at `http://%s` \n", *guiAddr)
}

func main() {
	flag.Parse()

	cmdChan := make(chan proxy.Cmd)
	cmdResultChan := make(chan proxy.CmdResult)
	connStateChan := make(chan proxy.ConnState)
	appReadyChan := make(chan bool)

	hub := pubsub.NewHub(cmdChan, cmdResultChan, connStateChan)

	go hub.Run()
	go runWsServer(hub)
	go runStaticServer()
	go appReadyInfo(appReadyChan)

	p, _ := proxy.NewProxyServer(*proxyAddr, *mysqlAddr)
	p.SetChannels(cmdChan, cmdResultChan, connStateChan, appReadyChan)
	p.Run()
}
