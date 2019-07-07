package main

import (
	"flag"
	"fmt"
	"github.com/orderbynull/lottip/http"
	"github.com/orderbynull/lottip/util"
	"srv-config/cfg"
	"time"

	"github.com/orderbynull/lottip/chat"
)

var (
	proxyAddr  = flag.String("proxy", "127.0.0.1:5555", "Proxy <host>:<port>")
	mysqlAddr  = flag.String("mysql", "180.167.115.58:10107", "MySQL <host>:<port>")
	guiAddr    = flag.String("gui", "127.0.0.1:9999", "Web UI <host>:<port>")
	useLocalUI = flag.Bool("use-local", false, "Use local UI instead of embed")
	mysqlDsn   = flag.String("mysql-dsn", "", "MySQL DSN for query execution capabilities")
)

func AppReadyInfo(appReadyChan chan bool) {
	<-appReadyChan
	time.Sleep(1 * time.Second)
	fmt.Printf("Forwarding queries from `%s` to `%s` \n", *proxyAddr, *mysqlAddr)
	fmt.Printf("Web gui available at `http://%s` \n", *guiAddr)
}

func main() {
	flag.Parse()
	cfg.RotationSystemLog("sql-proxy","/Users/zy/GolandProjects/src/github.com/orderbynull/lottip/")
	cmdChan := make(chan chat.Cmd)
	cmdResultChan := make(chan chat.CmdResult)
	connStateChan := make(chan chat.ConnState)
	appReadyChan := make(chan bool)

	hub := chat.NewHub(cmdChan, cmdResultChan, connStateChan)

	go hub.Run()
	go http.RunHttpServer(hub,mysqlDsn,useLocalUI,guiAddr)
	go AppReadyInfo(appReadyChan)

	p := util.MySQLProxyServer{
		cmdChan,
		cmdResultChan,
		connStateChan,
		appReadyChan,
		*mysqlAddr,
		*proxyAddr}
	p.Run()
}
