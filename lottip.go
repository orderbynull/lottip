package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/orderbynull/lottip/mysql"
)

// GuiData holds data sent to browser
type GuiData struct {
	Query      string
	NewSession bool
	SessionID  int
	Type       string
}

type SessionState struct {
	SessionID int
	State     bool
	Type      string
}

var proxyAddr = flag.String("listen", "127.0.0.1:4040", "proxy address")
var mysqlAddr = flag.String("mysql", "127.0.0.1:3306", "mysql address")
var guiAddr = flag.String("addr", "127.0.0.1:8080", "http service address")
var verbose = flag.Bool("verbose", true, "verbose mode")
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

//Lottip defines application
type Lottip struct {
	wg        *sync.WaitGroup
	gui       chan GuiData
	sessions  chan SessionState
	leftAddr  string
	rightAddr string
	verbose   bool
}

//New creates new Lottip application
func New(wg *sync.WaitGroup, leftAddr string, rightAddr string, verbose bool) *Lottip {
	l := &Lottip{}
	l.wg = wg
	l.gui = make(chan GuiData)
	l.sessions = make(chan SessionState)
	l.leftAddr = leftAddr
	l.rightAddr = rightAddr
	l.verbose = verbose

	return l
}

//Run fires up application
func (l *Lottip) Run() {
	go l.StartWebsocket()
	go l.StartProxy()
}

//StartWebsocket ...
func (l *Lottip) StartWebsocket() {
	http.Handle("/", http.FileServer(FS(false)))
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		var data []byte

		for {
			select {
			case q := <-l.gui:
				data, _ = json.Marshal(q)
				break
			case s := <-l.sessions:
				data, _ = json.Marshal(s)
				l.log("State received")
				break
			}

			err = c.WriteMessage(1, data)
			if err != nil {
				l.log("Error writing to socket: " + err.Error())
				break
			}
		}
	})

	log.Fatal(http.ListenAndServe(*guiAddr, nil))
}

//StartProxy ...
func (l *Lottip) StartProxy() {
	defer l.wg.Done()

	sourceListener, err := net.Listen("tcp", l.leftAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer sourceListener.Close()

	sessionID := 1

	for {
		leftConn, err := sourceListener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go l.HandleConnection(leftConn, sessionID)

		sessionID++
	}
}

//HandleConnection ...
func (l *Lottip) HandleConnection(leftConn net.Conn, sessionID int) {
	defer leftConn.Close()
	defer l.SessionEnded(sessionID)

	l.SessionStarted(sessionID)

	var wg sync.WaitGroup

	rightConn, err := net.Dial("tcp", l.rightAddr)
	if err != nil {
		return
	}
	defer rightConn.Close()

	wg.Add(2)
	go func() {
		defer wg.Done()
		l.LeftToRight(leftConn, rightConn, sessionID)
	}()

	go func() {
		defer wg.Done()
		l.RightToLeft(rightConn, leftConn)
	}()
	wg.Wait()
}

func (l *Lottip) log(msg string) {
	if l.verbose {
		log.Println(msg)
	}
}

//SessionStarted ...
func (l *Lottip) SessionStarted(sessID int) {
	l.sessions <- SessionState{SessionID: sessID, State: true, Type: "State"}
}

//SessionEnded ...
func (l *Lottip) SessionEnded(sessID int) {
	l.sessions <- SessionState{SessionID: sessID, State: false, Type: "State"}
}

//RightToLeft ...
func (l *Lottip) RightToLeft(left, right net.Conn) {
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

//LeftToRight ...
func (l *Lottip) LeftToRight(left, right net.Conn, sessID int) {
	//Indicates first query in session
	//First query means session just started
	isNewSession := true

	for {
		packet, err := mysql.ProxyPacket(left, right)
		if err != nil {
			break
		}

		isNewSession = l.PushToWebSocket(packet, isNewSession, sessID)
	}
}

//PushToWebSocket ...
func (l *Lottip) PushToWebSocket(pkt *mysql.Packet, isNewSession bool, sessID int) bool {
	if pkt.Type == mysql.ComQuery {
		select {
		case l.gui <- GuiData{Query: pkt.Query, NewSession: isNewSession, SessionID: sessID, Type: "Query"}:
			return false
		default:
		}
	}

	return true
}

func main() {
	art := `
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

	fmt.Println(art)

	var wg sync.WaitGroup

	flag.Parse()

	wg.Add(1)

	lottip := New(&wg, *proxyAddr, *mysqlAddr, *verbose)
	lottip.Run()

	wg.Wait()
}
