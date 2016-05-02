package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"fmt"

	"github.com/gorilla/websocket"
	// "golang.org/x/text"
)

//----------------------
func ReadPacket(c net.Conn) ([]byte, error) {
	var payload []byte
	for {
		// Read packet header
		var head = make([]byte, 4)
		_, err := c.Read(head)
		if err != nil {
			return nil, fmt.Errorf("Cannot read header")
		}

		// Packet Length [24 bit]
		pktLen := int(uint32(head[0]) | uint32(head[1])<<8 | uint32(head[2])<<16)
		log.Printf("PKTLEN = %d", pktLen)
		if pktLen < 1 {
			return nil, fmt.Errorf("pktLen < 1")
		}

		// Read packet body [pktLen bytes]
		var data = make([]byte, pktLen)

		n, err := c.Read(data)
		if err != nil {
			return nil, fmt.Errorf("Cannot read payload")
		}
		log.Printf("READ = %d", n)

		isLastPacket := (pktLen < MaxPayloadLen)

		// Zero allocations for non-splitting packets
		if isLastPacket && payload == nil {
			log.Print("Non-splitting packet")
			data = append(head, data...)
			return data, nil
		}

		payload = append(payload, head...)
		payload = append(payload, data...)

		if isLastPacket {
			log.Print("Splitted packet")
			return payload, nil
		}
	}
}

//----------------------

type GuiData struct {
	Query string
}

const MaxPayloadLen int = 1<<24 - 1

var proxyAddr = flag.String("listen", "127.0.0.1:4040", "proxy address")
var mysqlAddr = flag.String("mysql", "127.0.0.1:3306", "mysql address")
var guiAddr = flag.String("addr", "127.0.0.1:8080", "http service address")
var verbose = flag.Bool("verbose", true, "mysql address")
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type Lottip struct {
	wg        *sync.WaitGroup
	gui       chan string
	leftAddr  string
	rightAddr string
	verbose   bool
}

func New(wg *sync.WaitGroup, leftAddr string, rightAddr string, verbose bool) *Lottip {
	l := &Lottip{}
	l.wg = wg
	l.gui = make(chan string, 10)
	l.leftAddr = leftAddr
	l.rightAddr = rightAddr
	l.verbose = verbose

	return l
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
			go l.leftToRight(conn, rightConn, &wg, true)
			go l.leftToRight(rightConn, conn, &wg, false)
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

func (l *Lottip) rightToLeft(right, left net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	buf := make([]byte, 65535)
	for {
		n, err := left.Read(buf)
		if n == 0 {
			break
		}
		if err == io.EOF {
			l.log("*** EOF ***")
			break
		}
		if err != nil {
			l.log("ERROR: " + err.Error())
			break
		}

		_, err = right.Write(buf[0:n])
		if err != nil {
			l.log("3:" + err.Error())
			break
		}
	}
}

func (l *Lottip) leftToRight(left, right net.Conn, wg *sync.WaitGroup, xx bool) {
	defer wg.Done()

	if xx {
		for {
			header := []byte{0, 0, 0, 0}

			l.log("Waiting for data from SRC")
			n, err := left.Read(header)
			if n == 0 {
				break
			}
			if err == io.EOF {
				l.log("*** EOF ***")
				break
			}
			if err != nil {
				l.log("ERROR: " + err.Error())
				break
			}

			sequence := uint8(header[3])
			log.Printf("SEQ = %d", sequence)

			length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

			buf := make([]byte, length)

			bn, err := left.Read(buf)
			if bn == 0 {
				break
			}
			if err == io.EOF {
				l.log("*** EOF ***")
				break
			}
			if err != nil {
				l.log("ERROR: " + err.Error())
				break
			}

			// log.Printf("READ = %d", bn)
			// if bn != length {
			// 	l.log("!!!!!!!!!!!!!!")
			// }

			_, err = right.Write(append(header, buf[0:bn]...))
			if err != nil {
				l.log("3:" + err.Error())
				break
			}

			if buf[0] == 0x03 && xx {

				select {
				case l.gui <- string(buf[1:bn]):
				default:
				}
			}
		}
	} else {
		for {
			buf := make([]byte, 65535)
			n, err := left.Read(buf)
			if err != nil {
				log.Println("Error: "+err.Error())
				break
			}
			
			right.Write(buf[:n])
		}
	}
}

func (l *Lottip) Run() {
	go l.startWebsocket()
	go l.startProxy()
}

func main() {
	var wg sync.WaitGroup

	flag.Parse()

	wg.Add(1)

	lottip := New(&wg, *proxyAddr, *mysqlAddr, *verbose)
	lottip.Run()

	wg.Wait()
}
