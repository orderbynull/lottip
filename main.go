package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"lottip/chat"
	"time"
)

var (
	proxyAddr          = flag.String("proxy", "127.0.0.1:4041", "Proxy <host>:<port>")
	LogRequests        = flag.Bool("log-requests", false, "Enable logging of requests")
	LogResponses       = flag.Bool("log-responses", false, "Enable logging of responses")
	LogResponsePackets = flag.Bool("log-response-packets", false, "Enable logging of response packets")
	LogAll             = flag.Bool("log-all", false, "Enable logging of requests, responses, and other events")
	LogJSON            = flag.Bool("log-format-json", false, "Log entries as JSON")
	mysqlAddr          = flag.String("mysql", "127.0.0.1:3306", "MySQL <host>:<port>")
	guiAddr            = flag.String("gui-addr", "127.0.0.1:9999", "Web UI <host>:<port>")
	guiEnabled         = flag.Bool("gui-enabled", true, "Enable the web-gui server")
	useLocalUI         = flag.Bool("use-local", false, "Use local UI instead of embed")
	mysqlDsn           = flag.String("mysql-dsn", "", "MySQL DSN for query execution capabilities")
)

func appReadyInfo(appReadyChan chan bool) {
	<-appReadyChan
	time.Sleep(1 * time.Second)
	fmt.Printf("Forwarding queries from `%s` to `%s` \n", *proxyAddr, *mysqlAddr)
	fmt.Printf("Web gui available at `http://%s` \n", *guiAddr)
}

func main() {
	flag.Parse()
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zerolog.TimestampFieldName = "logTimestamp"

	cmdChan := make(chan chat.Cmd)
	cmdResultChan := make(chan chat.CmdResult)
	connStateChan := make(chan chat.ConnState)
	appReadyChan := make(chan bool)

	hub := chat.NewHub(cmdChan, cmdResultChan, connStateChan)

	go hub.Run()
	if guiEnabled != nil && *guiEnabled {
		go runHttpServer(hub)
	}
	go appReadyInfo(appReadyChan)

	p := MySQLProxyServer{cmdChan, cmdResultChan, connStateChan, appReadyChan, *mysqlAddr, *proxyAddr}
	p.run()
}
