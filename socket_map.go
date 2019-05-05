package brpc

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Socket struct {
	sync.RWMutex
	writeQueue chan *RPCStub

	closed chan bool
	codec  ProtocolCodec
	conn   net.Conn

	// stubs
	stubs map[int64]*RPCStub
}

func (s *Socket) WriteAsync(req *RPCStub) {
	s.writeQueue <- req
}

func (s *Socket) WriteGoroutine() {
	var stopped = false

	for !stopped {
		select {
		case <-s.closed:
			stopped = true
		case stub := <-s.writeQueue:
			var data, err = s.codec.Encode(stub)
			if err != nil {
				stub.Controller().SetFailed(err)
			} else {
				s.Lock()
				s.conn.Write(data)
				s.Unlock()
			}
			log.Printf("Socket.WriteAsync()")
		}
	}
}

// 1. Read RPC.Response...
// 2. Close Event
// 3. Timeout Event.
func (s *Socket) ReadGoroutine() {

	var err error
	var stopped = false

	for !stopped {
		stub, err := s.codec.Decode(s.conn)
		if err != nil {
			break
		}
		stub.Done()
	}

	if err == io.EOF {
	}
	s.closed <- true
}

func NewSocket(conn net.Conn, codec ProtocolCodec) *Socket {
	return &Socket{
		writeQueue: make(chan *RPCStub, 10000),
		closed:     make(chan bool),
		codec:      codec,
		conn:       conn,
		stubs:      make(map[int64]*RPCStub),
	}
}

func CreateSocket(node ServerNode, codec ProtocolCodec, options *ChannelOptions) (*Socket, error) {

	if options.IsTLS() {
		return nil, fmt.Errorf("(TLS) not support now")
	} else {
		var addr = fmt.Sprintf("%s:%d", node.Ip, node.Port)
		var conn, err = net.DialTimeout("tcp", addr, options.timeout_ms*time.Millisecond)
		if err != nil {
			return nil, err
		}
		var socket = NewSocket(conn, codec)
		go socket.ReadGoroutine()
		go socket.WriteGoroutine()
		return socket, nil
	}
}
