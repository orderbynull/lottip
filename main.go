package main

import (
	"flag"
	"fmt"

	"github.com/orderbynull/lottip/lottip"
)

var proxyAddr = flag.String("listen", "127.0.0.1:4040", "Proxy address")
var mysqlAddr = flag.String("mysql", "127.0.0.1", "MySQL address")
var mysqlPort = flag.String("port", "3306", "MySQL port")
var guiAddr = flag.String("addr", "127.0.0.1:8080", "HTTP service address")
var verbose = flag.Bool("verbose", true, "Verbose mode")
var art = `
                    ___                                               ___   
                   /  /\          ___         ___       ___          /  /\  
                  /  /::\        /  /\       /  /\     /  /\        /  /::\ 
  ___     ___    /  /:/\:\      /  /:/      /  /:/    /  /:/       /  /:/\:\
 /__/\   /  /\  /  /:/  \:\    /  /:/      /  /:/    /__/::\      /  /:/~/:/
 \  \:\ /  /:/ /__/:/ \__\:\  /  /::\     /  /::\    \__\/\:\__  /__/:/ /:/ 
  \  \:\  /:/  \  \:\ /  /:/ /__/:/\:\   /__/:/\:\      \  \:\/\ \  \:\/:/  
   \  \:\/:/    \  \:\  /:/  \__\/  \:\  \__\/  \:\      \__\::/  \  \::/   
    \  \::/      \  \:\/:/        \  \:\      \  \:\     /__/:/    \  \:\   
     \__\/        \  \::/          \__\/       \__\/     \__\/      \  \:\  
                   \__\/                                             \__\/                                                           
	`

func main() {
	fmt.Println(art)
	flag.Parse()
	l := lottip.New(*proxyAddr, fmt.Sprintf("%s:%s", *mysqlAddr, *mysqlPort), *guiAddr, *verbose)
	l.Run()
}
