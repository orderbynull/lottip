package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	// "golang.org/x/text"
)

type GuiData struct {
	Query string
}

var proxyAddr = flag.String("listen", "127.0.0.1:4040", "proxy address")
var mysqlAddr = flag.String("mysql", "127.0.0.1:3306", "mysql address")
var guiAddr = flag.String("addr", "127.0.0.1:8080", "http service address")
var verbose = flag.Bool("verbose", true, "mysql address")
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

//Lottip defines application
type Lottip struct {
	wg        *sync.WaitGroup
	gui       chan string
	leftAddr  string
	rightAddr string
	verbose   bool
}

//New creates new Lottip application
func New(wg *sync.WaitGroup, leftAddr string, rightAddr string, verbose bool) *Lottip {
	l := &Lottip{}
	l.wg = wg
	l.gui = make(chan string, 10)
	l.leftAddr = leftAddr
	l.rightAddr = rightAddr
	l.verbose = verbose

	return l
}

//Run fires up application
func (l *Lottip) Run() {
	go l.startWebsocket()
	go l.startProxy()
}

func (l *Lottip) startWebsocket() {
	http.Handle("/", http.FileServer(FS(false)))
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		data := GuiData{}

		for {
			data.Query = <-l.gui

			json, _ := json.Marshal(data)

			err = c.WriteMessage(1, json)
			if err != nil {
				break
			}
		}
	})

	log.Fatal(http.ListenAndServe(*guiAddr, nil))
}

func (l *Lottip) startProxy() {
	defer l.wg.Done()

	sourceListener, err := net.Listen("tcp", l.leftAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer sourceListener.Close()

	for {
		l.log("Waiting for LEFT connection")

		leftConn, err := sourceListener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		l.log("Received LEFT connection")

		go func(conn net.Conn) {
			defer conn.Close()

			var wg sync.WaitGroup

			l.log("Handling LEFT connection")

			rightConn, err := net.Dial("tcp", l.rightAddr)
			if err != nil {
				log.Println(err)
				return
			}
			defer rightConn.Close()

			wg.Add(2)
			go func(){
				defer wg.Done()
				l.leftToRight(conn, rightConn)
			}()
			
			go func(){
				defer wg.Done()
				l.rightToLeft(rightConn, conn)
			}()
			wg.Wait()

			l.log("End of handling LEFT connection")
		}(leftConn)
	}
}

func (l *Lottip) log(msg string) {
	if l.verbose {
		log.Println(msg)
	}
}

func (l *Lottip) rightToLeft(left, right net.Conn) {
	for {
		buf := make([]byte, 65535)
		n, err := left.Read(buf)
		if err != nil {
			log.Println("Error: " + err.Error())
			break
		}

		right.Write(buf[:n])
	}
}

func (l *Lottip) leftToRight(left, right net.Conn) {
	for {
		header := []byte{0, 0, 0, 0}

		_, err := left.Read(header)
		if err == io.EOF {
			break
		}
		if err != nil {
			l.log("ERROR: " + err.Error())
			break
		}

		length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

		buf := make([]byte, length)

		bn, err := left.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			l.log("ERROR: " + err.Error())
			break
		}

		_, err = right.Write(append(header, buf[0:bn]...))
		if err != nil {
			break
		}

		if buf[0] == 0x03 {
			select {
			case l.gui <- string(buf[1:bn]):
			default:
			}
		}
	}
}

func main() {
	var wg sync.WaitGroup

	flag.Parse()

	wg.Add(1)

	lottip := New(&wg, *proxyAddr, *mysqlAddr, *verbose)
	lottip.Run()

	wg.Wait()
}