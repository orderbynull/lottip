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

// GuiData holds data sent to browser via websocket
type GuiData struct {
	Query      string
	NewSession bool
	SessionID  int
	Type       string
}

//SessionState ...
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

//Lottip defines application structure
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

		//Init websocket connection
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		//Holds data to be sent via websocket
		var data []byte

		for {
			//Reading data came for gui and preparing to send
			select {
			case q := <-l.gui:
				data, _ = json.Marshal(q)
				break
			case s := <-l.sessions:
				data, _ = json.Marshal(s)
				l.log("State received")
				break
			}

			//Pushing data to gui
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

	//Listening for connection from client
	sourceListener, err := net.Listen("tcp", l.leftAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer sourceListener.Close()

	//Start counting sessions
	sessionID := 1

	for {
		//Accepting connection from client
		leftConn, err := sourceListener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		//Parse and proxy packets to server
		go l.HandleConnection(leftConn, sessionID)

		//Each new connection from client means new session
		sessionID++
	}
}

//HandleConnection ...
func (l *Lottip) HandleConnection(leftConn net.Conn, sessionID int) {
	defer leftConn.Close()
	defer l.SessionEnded(sessionID)

	//Send "session started" event to gui
	l.SessionStarted(sessionID)

	var wg sync.WaitGroup

	//Trying to connect to server
	rightConn, err := net.Dial("tcp", l.rightAddr)
	if err != nil {
		return
	}
	defer rightConn.Close()

	wg.Add(2)
	//Start passing packets from client to server
	go func() {
		defer wg.Done()
		l.LeftToRight(leftConn, rightConn, sessionID)
	}()

	//Start passing packets from server to client
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

//SessionStarted sends "session started" event to gui
func (l *Lottip) SessionStarted(sessID int) {
	l.sessions <- SessionState{SessionID: sessID, State: true, Type: "State"}
}

//SessionEnded sends "session ended" event to gui
func (l *Lottip) SessionEnded(sessID int) {
	l.sessions <- SessionState{SessionID: sessID, State: false, Type: "State"}
}

//RightToLeft passes packets from server to client
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

//LeftToRight passes packets from client to server
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
