package main

import (
	"flag"
	"fmt"
	"lottip/chat"
	"time"
)

var (
	proxyAddr  = flag.String("proxy", "127.0.0.1:4041", "Proxy <host>:<port>")
	mysqlAddr  = flag.String("mysql", "127.0.0.1:3306", "MySQL <host>:<port>")
	guiAddr    = flag.String("gui", "127.0.0.1:9999", "Web UI <host>:<port>")
	useLocalUI = flag.Bool("use-local", false, "Use local UI instead of embed")
	mysqlDsn   = flag.String("mysql-dsn", "", "MySQL DSN for query execution capabilities")
)

func appReadyInfo(appReadyChan chan bool) {
	<-appReadyChan
	time.Sleep(1 * time.Second)
	fmt.Printf("Forwarding queries from `%s` to `%s` \n", *proxyAddr, *mysqlAddr)
	fmt.Printf("Web gui available at `http://%s` \n", *guiAddr)
}

func main() {
	flag.Parse()

	cmdChan := make(chan chat.Cmd)
	cmdResultChan := make(chan chat.CmdResult)
	connStateChan := make(chan chat.ConnState)
	appReadyChan := make(chan bool)

	hub := chat.NewHub(cmdChan, cmdResultChan, connStateChan)

	go hub.Run()
	go runHttpServer(hub)
	go appReadyInfo(appReadyChan)

	p := MySQLProxyServer{cmdChan, cmdResultChan, connStateChan, appReadyChan, *mysqlAddr, *proxyAddr}
	p.run()
}
