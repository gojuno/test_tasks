package junoKvServer

import (
	"net"
	"sync"
	"bufio"
	"errors"
	/** @todo #1:60m/DEV think use junoKVClient in future for write redis response protocol ? */
	"github.com/siddontang/goredis"
	"log"
	"runtime"
	"strings"
	"time"
)

var (
	errEmptyCommand = errors.New("empty command")
	errNotFound     = errors.New("command not found")
	errCmdParams    = errors.New("invalid command param")
	errValue        = errors.New("value is not an integer or out of range")
)

var (
	newLine = []byte("\r\n")

	nullBulk  = []byte("-1")
	nullArray = []byte("-1")

	okStatus = "OK"
)

// KVServer main structure with listener, client connections pool and synchronization primitives
type KVServer struct {
	connWait sync.WaitGroup
	listener net.Listener
	stoped   bool
	quit     chan struct{}

	/** @todo #1:60m/ARCH,DEV investigate how to implement multiple TCP clients without Mutex? */
	clientsMutex sync.Mutex
	clients      map[*client]struct{}
	data         *kvData
}

type client struct {
	server *KVServer
	conn   net.Conn

	cmd  string
	args [][]byte

	reqReader  *goredis.RespReader
	respWriter *respWriter
}

//NewKVServer create KVServer and initialize listen socket
func NewKVServer(addr string) (*KVServer, error) {
	if addr == "" {
		addr = "0.0.0.0:8379"
	}
	s := new(KVServer)
	s.quit = make(chan struct{})
	s.clientsMutex = sync.Mutex{}
	s.clients = make(map[*client]struct{})
	s.connWait = sync.WaitGroup{}
	s.data = newKVData()

	var err error
	/** @todo #1:15m/DEV need ipv6 support for listen */
	if s.listener, err = net.Listen("tcp", addr); err != nil {
		return nil, err
	}

	return s, nil
}

// Start main cycle with accept incoming connections with spawn response go-routine
func (server *KVServer) Start() {
	for {
		select {
		case <-server.quit:
			return
		default:
			conn, err := server.listener.Accept()
			if err != nil {
				/** @todo #1:60m/DEV,ARCH to be or not to be structured log ? maybe try https://github.com/rs/zerolog or https://github.com/oklog/oklog ? */
				log.Print(err.Error())
				continue
			}

			newClient(conn, server)
		}
	}

}

// Stop close listener and wait all response will be completed
func (server *KVServer) Stop() {
	if server.stoped {
		return
	}
	server.stoped = true

	close(server.quit)

	server.listener.Close()

	//wait all connection closed
	server.connWait.Wait()

}

func newClient(conn net.Conn, server *KVServer) {
	c := new(client)
	c.server = server
	c.conn = conn
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		/** @todo #1:60m/ARCH need config settings, what right constant value ? */
		tcpConn.SetReadBuffer(4096)
		tcpConn.SetWriteBuffer(4096)
	}
	br := bufio.NewReaderSize(conn, 4096)

	c.reqReader = goredis.NewRespReader(br)
	c.respWriter = newResponseWriter(conn, 4096)

	server.connWait.Add(1)

	server.addClient(c)

	go c.processResponse()

}

func (server *KVServer) addClient(c *client) {
	server.clientsMutex.Lock()
	server.clients[c] = struct{}{}
	server.clientsMutex.Unlock()
}

func (server *KVServer) deleteClient(c *client) {
	server.clientsMutex.Lock()
	delete(server.clients, c)
	server.clientsMutex.Unlock()
}

func (c *client) processResponse() {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			buf = buf[0:n]

			log.Fatalf("clientResponse run panic %s:%v", buf, e)
		}

		c.conn.Close()

		c.server.deleteClient(c)

		c.server.connWait.Done()
	}()

	select {
	case <-c.server.quit:
		//check server closed
		return
	default:
		break
	}

	kc := time.Duration(1) * time.Second
	for {
		if kc > 0 {
			c.conn.SetReadDeadline(time.Now().Add(kc))
		}

		c.cmd = ""
		c.args = nil

		reqData, err := c.reqReader.ParseRequest()
		if err == nil {
			err = c.handleRequest(reqData)
		}

		if err != nil {
			return
		}
	}
}

func (c *client) handleRequest(reqData [][]byte) error {
	if len(reqData) == 0 {
		c.cmd = ""
		c.args = reqData[0:0]
	} else {
		c.cmd = strings.ToLower(string(reqData[0]))
		c.args = reqData[1:]
	}

	c.handleCommand()

	return nil
}

func (c *client) handleCommand() {
	var err error

	if len(c.cmd) == 0 {
		err = errEmptyCommand
	} else if exeCmd, ok := cmdList[c.cmd]; !ok {
		err = errNotFound
	} else {
		err = exeCmd(c)
	}

	if err != nil {
		c.respWriter.writeError(err)
	}
	c.respWriter.flush()
	return
}
