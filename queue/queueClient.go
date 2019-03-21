package queue

import (
	"fmt"
	"net"
	"time"
	"strconv"
	"log"
)


//Client ... one client
type Client struct {
	Rch chan []byte
	Wch chan []byte
	Dch chan bool
}


// ClientCluster ...
type ClientCluster struct {
	cMap              ConcurrentMap   // map contains all clients, key -> uuid, value -> Client
	globalDch         chan string     // done channel
	concurrentClients int             // init param, concurrent clients count
	reportStatistics  bool            // init param, report statistics or not
}

// CreateCluster ...
func CreateCluster(
	concurrentClients int,
	reportStatistics  bool,
	targetIp          string,
	targetPort        int,) {
	clientCluster := &ClientCluster{
		cMap:    New(),
		globalDch: make(chan string),
		concurrentClients: concurrentClients,
		reportStatistics: reportStatistics,
	}

	for i := 0; i < concurrentClients; i++ {
		// generate uid 
		//uid := RandStringRunes(10)
		uid := "asdfasdfasdfa" + string(i)
		clientCluster.createClient(targetIp, targetPort, uid)
	}

	if reportStatistics {
		go clientCluster.report()
	}
	
	for {
		select {
			case uid := <- clientCluster.globalDch:
				fmt.Println(uid, "connection cloosed")
				clientCluster.cMap.Remove(uid)

				if clientCluster.cMap.Count() == 0 {
					goto ForEnd
			}
		}
	}
	ForEnd:
	log.Println("done")
}

func (cl *ClientCluster) report() {
	// report interval 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <- cl.globalDch:
			return 

		case <-ticker.C:
			activeClients := cl.cMap.Count()
			fmt.Printf("report statistics: %d clients total, %d connections waiting, %d connections received token \n", cl.concurrentClients, 
						activeClients, cl.concurrentClients - activeClients)
		}
	}
}

func (cl *ClientCluster) createClient(targetIp string, targetPort int,uid string) {
	client := &Client{
		Rch : make(chan []byte),
		Wch : make(chan []byte),
		Dch : make(chan bool),
	}
	go client.start(cl, targetIp, targetPort, uid)
}

func (cl *ClientCluster) registerClient(c *Client, uid string) {
	cl.cMap.Set(uid, c)
}


func (c *Client) start(cl *ClientCluster, targetIp string, targetPort int, uid string) {
	addr, err := net.ResolveTCPAddr("tcp", targetIp + ":" + strconv.Itoa(targetPort))
	conn, err := net.DialTCP("tcp", nil, addr)
	defer func() { 
		if conn != nil {
			conn.Close()
		}
	}()

	if err != nil {
		fmt.Println("connect to server failed:", err.Error())
		return
	}

	cl.registerClient(c, uid)

	conn.Write(AssemblePacket(Req_REGISTER, []byte(uid)))

	go c.rHandler(conn, uid)

	// no need to use write handle in this case
	//go c.wHandler(conn, uid)

	select {
		case <- c.Dch:
			fmt.Println(uid, "connection closed")
			cl.globalDch <- uid
	}
	
}

func (c *Client) rHandler(conn *net.TCPConn,  uid string) {

	for {
		//fmt.Println("process rHandler", uid)
		data := make([]byte, 128)

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		i, err := conn.Read(data)

		if err != nil {	
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				//timeout, go on
				continue
			} else {
				// network error
				c.Dch <- true
				return 
			}
		}

		if i != 0 {
			for {
				var index uint8
				var content []byte
				data, index, content = DisassemblePacket(data)
	
				if index == 0 {
					// for simplicity, any bytes left are all discarded
					break
				}
	
				if index == Res_CURRENT_POS {
					fmt.Println("report current positon", uid, BytesToInt32(content))
				} else if index == Res_GET_TOKEN {
					fmt.Println("get token", uid, string(content))
	
					//TODO: do next things, server will close the connection, so we don't do anything here
				} 
	
			}
		}
		
	}
}

/*
func (c *Client) wHandler(conn net.Conn,  uid string) {
	for {
		select {
			case msg := <- c.Wch:
				fmt.Println((msg[0]))
				//fmt.Println("send data after: " + string(msg[1:]))
				conn.Write(msg)
			case <- c.Dch :
				return
		}
	}

}
*/

