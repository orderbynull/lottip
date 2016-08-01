package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/orderbynull/lottip/lottip"
)

var proxyAddr = flag.String("listen", "127.0.0.1:4040", "proxy address")
var mysqlAddr = flag.String("mysql", "127.0.0.1:3306", "mysql address")
var guiAddr = flag.String("addr", "127.0.0.1:8080", "http service address")
var verbose = flag.Bool("verbose", true, "verbose mode")
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

	var wg sync.WaitGroup

	flag.Parse()

	wg.Add(1)

	l := lottip.New(&wg, *proxyAddr, *mysqlAddr, *guiAddr, *verbose)
	l.Run()

	wg.Wait()
}
