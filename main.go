package main

import (
	"flag"
	"fmt"

	"github.com/orderbynull/lottip/lottip"
)

var proxyAddr = flag.String("listen", "127.0.0.1:4040", "Proxy address")
var mysqlAddr = flag.String("mysql", "127.0.0.1:3306", "MySQL address")
var guiAddr = flag.String("gui", "127.0.0.1:8080", "Gui address")
var verbose = flag.Bool("verbose", false, "Verbose mode")
var useLocal = flag.Bool("local", false, "Use local gui files")
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
	l := lottip.New(*proxyAddr, *mysqlAddr, *guiAddr, *verbose, *useLocal)
	l.Run()
}
