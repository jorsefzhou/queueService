package queue

//import _ "net/http/pprof"
//import  "net/http"

// server
import (
	"fmt"
	"net"
	"time"
	"sync/atomic"
)


type CS struct {
	Rch chan []byte
	Wch chan ChannelMsg
	Cch chan bool
	conn net.Conn
	uid   string
	t     int64
	pos   int64
	offline  bool
}

type ChannelMsg struct {
    req  byte
    data []byte
}

func NewCs(uid string, pos int64) *CS {
	return &CS { 
		Rch: make(chan []byte),  // read data channel
		Wch: make(chan ChannelMsg),  // write data channel
		Cch: make(chan bool),    // clear resource channel
		uid: uid, 
		pos : pos, 
		t: time.Now().Unix(),
		offline : false,
	}
}

// Server
type Server struct {
	activeClientsMap ConcurrentMap   // store all clients info
	queueChan  chan string           // queue chain
	currentQueueHeadPos   int64
	currentQueueTailPos int64

	listenPort   int
	bingdingIP   string
	passPeriod   int
	passNumber   int
	maxSeats     int
}


// CreateServer ...
func CreateServer(
	listenPort   int,
	bingdingIP   string,
	passPeriod   int,
	passNumber   int,
	maxSeats     int,) *Server {

//	go func() {
//			fmt.Println(http.ListenAndServe("localhost:6060", nil))
//	}()

	server := &Server{
		activeClientsMap: New(),
		queueChan:  make(chan string, maxSeats),

		listenPort: listenPort,
		bingdingIP: bingdingIP,
		passPeriod: passPeriod,
		passNumber: passNumber,
		maxSeats: maxSeats,
	}

	server.Start()
	return server
}


// Start ...
func (s *Server) Start() {
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP(s.bingdingIP), s.listenPort, ""})
	if err != nil {
		fmt.Println("listen failed:", err.Error())
		return
	}
	
	fmt.Println("listening, waiting for clients")

	go s.dispatchPass()

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			fmt.Println("client connect exception", err.Error())
			continue
		}
		go s.sHandler(conn)
	}	
}

func (s *Server) dispatchPass() {
	for {
		lastWindowEntry := time.Now().Unix()

		fmt.Println("call dispatchPass", time.Now().Unix())
		
		i := 0
		needBroadcast := false

		for {
			if (time.Now().Unix() - lastWindowEntry) > int64(s.passPeriod) {
				break
			}

			if i >= s.passNumber {
				// TODO: calculate sleep precise time interval
				time.Sleep(1000 * time.Millisecond)

				if needBroadcast {
					go s.broadcastPosition()
					fmt.Println(i, "users authenticated")
					needBroadcast = false
				}

				continue
			}

			select {
			case uid := <- s.queueChan:
				// TODO: generate a valid pass token
				token := "you are passed!"

				if cs, ok := s.activeClientsMap.Get(uid); !ok {
					//fmt.Println("conn die, close WHandler", uid)
				} else {
					tcs := cs.(*CS)

					if tcs.offline {
						s.setCurrentQueueHeadPosValue(tcs.pos)
						s.activeClientsMap.Remove(uid)	
						close(tcs.Cch)

						continue
					}
					
					// send channel msg
					tcs.Wch <- ChannelMsg {Res_GET_TOKEN, []byte(token)}
					
					// set current queue head pos
					s.setCurrentQueueHeadPosValue(tcs.pos)
					s.activeClientsMap.Remove(uid)
					i ++
					needBroadcast = true
				}
			default:
				time.Sleep(1000 * time.Millisecond)
				
				if needBroadcast {
					go s.broadcastPosition()
					fmt.Println(i, "users authenticated")
					needBroadcast = false
				}
			}
		}
	}
}

func (s *Server) addAndGetCurrentQueueTailPosValue(delta int64) int64 {
    for {
        v := atomic.LoadInt64(&s.currentQueueTailPos)
        if atomic.CompareAndSwapInt64(&s.currentQueueTailPos, v, (v + delta)){
			return v + delta
        }
	}
}

func (s *Server) setCurrentQueueHeadPosValue(val int64) {
	atomic.StoreInt64(&s.currentQueueHeadPos, val)
}


func (s *Server) broadcastPosition() {
	fmt.Println("broadcast queue position, current waiting user count: ", s.activeClientsMap.Count())

	// modify to prevent server deadlock
	items := s.activeClientsMap.Items()

	for _, value := range items {
		v := value.(*CS)

		if !v.offline {
			s.tellClientCurrentPos(v)
		}
	}

	/*
	items.IterCb(func(key string, val interface{}) {
		v := val.(*CS)

		if !v.offline {
			s.tellClientCurrentPos(v)
		}
	})*/
}


func (s *Server) sHandler(conn net.Conn) {
	defer func() { 
		if conn != nil {
			conn.Close()
		}
	}()

	data := make([]byte, 128)

	var C *CS
	for {
		if conn == nil {
			return
		}

		conn.Read(data)
		_, index, content := DisassemblePacket(data)

		if index == Req_REGISTER {
			uid := string(content)
			//TODO: verify uid or associated token (ignored here for simplicity) to ensure the client is valid

			fmt.Println("register client get uid", 	s.currentQueueHeadPos, s.currentQueueTailPos)	
			if s.activeClientsMap.Count() >= s.maxSeats {
				fmt.Println("reach max seat, return")
				conn.Close()
				conn = nil
				return
			}

			if cs, ok := s.activeClientsMap.Get(uid); ok {
				C = cs.(*CS)
				if C.offline {
					// reconnect
					C.offline = false
	
					C.conn = conn
					s.tellClientCurrentPos(C)
					
					return
				} else {
					// already connected, ignore
					conn.Close()
					conn = nil
				}
				
				continue			
			}

			// create new user into dict
			pos := s.addAndGetCurrentQueueTailPosValue(1)
			C = NewCs(uid, pos)
			s.queueChan <- uid
			s.activeClientsMap.Set(uid, C)

			C.conn = conn
			// go write handler and read handler
			go s.sWHandler(C)
			go s.sRHandler(C)
			
			// inform client current position
			s.tellClientCurrentPos(C)
			break
		} else{
			conn.Close()
			conn = nil
			return
		}
	}

	select {
	case <-C.Cch:
		//fmt.Println("close handler goroutine")
		return
	}
}

func (s *Server) tellClientCurrentPos(C *CS) {
	position := int32(C.pos - atomic.LoadInt64(&s.currentQueueHeadPos))
	C.Wch <- ChannelMsg {Res_CURRENT_POS, Int32ToBytes(position)}
}

func (s *Server) sWHandler(C *CS) {
	for {
		select {
		case msg := <-C.Wch:
			
			data := AssemblePacket(msg.req, msg.data)
			C.conn.Write(data)

			if msg.req == Res_GET_TOKEN {
				// work done, write data and close in one second
				go func() {
					time.Sleep(1 * time.Second)
					close(C.Cch)
				}()
				return 
			}

		case <- C.Cch:
			//fmt.Println("close sWHandler goroutine")
			return 
		}
	}
}


func (s *Server) sRHandler(C *CS) {
	for {
		// check if connection is broken
		data := make([]byte, 128)
		C.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err := C.conn.Read(data)

		if err != nil {	
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				//timeout, go on
			} else {
				//fmt.Println("connection is closed", C.uid)
				C.conn.Close()
				//fmt.Println("set offline", true)
				C.offline = true
				
				// sleep one second
				time.Sleep(1000 * time.Millisecond)
			}
		}
		
		select {
			case <- C.Cch:
				//fmt.Println("close sRHandler goroutine")
				return 
			default:
				continue
		}
	}	
}