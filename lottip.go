package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"

	// "golang.org/x/text"
	"github.com/gorilla/websocket"
)

func pipe(src, dst io.ReadWriter, dir bool, errsig chan bool, guichan chan string, guiReady chan bool) {
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			//log.Print(err)
			errsig <- true
			return
		}
		b := buff[:n]

		//write out result
		n, err = dst.Write(b)
		if err != nil {
			//log.Print(err)
			errsig <- true
			return
		}

		if dir == true {
			cmd := buff[4]
			if cmd == 0x03 {
				//log.Printf("\n************************ \n\n %s \n\n************************ \n", buff)

				select {
				case ready := <-guiReady:
					if ready == true {
						//log.Println("Sending to channel...")
						guichan <- string(buff[5:n])
					}
				default:

				}
			}

			if cmd == 0x01 {
				log.Print("COM_QUIT")
			}
		}
	}
}

type GuiStruct struct {
	Query string
}

var guiAddr = flag.String("gui", ":8080", "gui address")
var proxyAddr = flag.String("listen", "127.0.0.1:4040", "proxy address")
var mysqlAddr = flag.String("mysql", "127.0.0.1:3306", "mysql address")

var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func echo(w http.ResponseWriter, r *http.Request, guiReady chan bool, guiChan chan string) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	data := GuiStruct{Query: ""}
	for {
		guiReady <- true

		data.Query = ""

		data.Query = <-guiChan

		json, _ := json.Marshal(data)

		err = c.WriteMessage(1, json)
		if err != nil {
			log.Println("write:", err)
			break
		}

	}
}

func main() {
	flag.Parse()

	guiChan := make(chan string, 1000)
	guiReady := make(chan bool)

	go func() {

		log.Println("Waiting for connection request for MySQL...")

		sourceListener, err := net.Listen("tcp", *proxyAddr)
		if err != nil {
			log.Fatal(err)
		}

		for {
			sourceConnection, err := sourceListener.Accept()
			if err != nil {
				log.Print(err)
				continue
			}

			log.Println("Received connection request for MySQL")

			errsig := make(chan bool)

			func(sconnection net.Conn) {
				defer sconnection.Close()

				// log.Println("Dialing MySQL...");

				destinationConnection, err := net.Dial("tcp", *mysqlAddr)
				if err != nil {
					log.Println(err)
                    return
				}

				//log.Println("Connection to MySQL succedded")

				defer destinationConnection.Close()

				go pipe(sconnection, destinationConnection, true, errsig, guiChan, guiReady)
				go pipe(destinationConnection, sconnection, false, errsig, guiChan, guiReady)

				<-errsig
			}(sourceConnection)
		}
	}()

    http.Handle("/", http.FileServer(FS(false)))
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		echo(w, r, guiReady, guiChan)
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}
