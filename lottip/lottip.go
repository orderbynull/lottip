package lottip

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/orderbynull/lottip/embed"
	"github.com/orderbynull/lottip/mysql"
)

//gQuery holds query and it's session sent via websocket
type gQuery struct {
	Query     string
	SessionID int
	Type      string
}

//gState holds sessions state sent via websocket
type gState struct {
	SessionID int
	State     bool
	Type      string
}

//Lottip defines application structure
type Lottip struct {
	wg         *sync.WaitGroup
	gQueryChan chan gQuery
	gStateChan chan gState
	leftAddr   string
	rightAddr  string
	guiAddr    string
	verbose    bool
	useLocal   bool
}

//New creates new Lottip application
func New(leftAddr string, rightAddr string, guiAddr string, verbose bool, useLocal bool) *Lottip {
	l := &Lottip{}
	l.wg = &sync.WaitGroup{}
	l.gQueryChan = make(chan gQuery)
	l.gStateChan = make(chan gState)
	l.leftAddr = leftAddr
	l.rightAddr = rightAddr
	l.guiAddr = guiAddr
	l.verbose = verbose
	l.useLocal = useLocal

	return l
}

//Run fires up application
func (l *Lottip) Run() {
	l.wg.Add(2)
	go l.StartWebsocket()
	go l.StartProxy()
	l.wg.Wait()
}

//StartWebsocket starts listening for WS connection
func (l *Lottip) StartWebsocket() {
	defer l.wg.Done()

	http.Handle("/", http.FileServer(embed.FS(l.useLocal)))

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {

		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		//Init websocket connection
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		//Proper handling 'close' message from the peer
		//https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages
		go func() {
			if _, _, err := c.NextReader(); err != nil {
				c.Close()
			}
		}()

		//Holds data to be sent via websocket
		var data []byte

		for {
			//Reading data came for gui and preparing to send
			select {
			case q := <-l.gQueryChan:
				data, _ = json.Marshal(q)
				break
			case s := <-l.gStateChan:
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

	log.Fatal(http.ListenAndServe(l.guiAddr, nil))
}

//StartProxy starts listening for MySQL connection
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
		go l.handleConnection(leftConn, sessionID)

		//Each new connection from client means new session
		sessionID++
	}
}

//handleConnection ...
func (l *Lottip) handleConnection(leftConn net.Conn, sessionID int) {
	defer leftConn.Close()
	defer l.sessionState(sessionID, false)

	//Send "session started" event to gui
	l.sessionState(sessionID, true)

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
func (l *Lottip) sessionState(sessID int, started bool) {
	select {
	case l.gStateChan <- gState{SessionID: sessID, State: started, Type: "State"}:
	default:
	}
}

//RightToLeft passes packets from server to client
func (l *Lottip) RightToLeft(right, left net.Conn) {
	for {
		_, err := mysql.ProxyPacket(right, left)
		if err != nil {
			break
		}

		//pktType, err := mysql.GetResponsePktType(pkt)
		//if err == nil {
		//	switch pktType {
		//	case mysql.ResponseOkPacket:
		//		_, err := mysql.ParseOk(pkt)
		//		if err == nil {
		//			//l.log("OK received")
		//		}
		//	case mysql.ResponseErrPacket:
		//		_, err := mysql.ParseErr(pkt)
		//		if err == nil {
		//			//l.log("ERR received")
		//		}
		//	}
		//} else {
		//	//l.log("Unknown packet")
		//}
	}
}

//LeftToRight passes packets from client to server
func (l *Lottip) LeftToRight(left, right net.Conn, sessID int) {

	for {
		pkt, err := mysql.ProxyPacket(left, right)
		if err != nil {
			break
		}

		query, err := mysql.GetQuery(pkt)
		if err == nil {
			l.PushToWebSocket(query, sessID)
		}

		query = ""
	}
}

//PushToWebSocket ...
func (l *Lottip) PushToWebSocket(query string, sessID int) {
	select {
	case l.gQueryChan <- gQuery{Query: query, SessionID: sessID, Type: "Query"}:
	default:
	}
}
