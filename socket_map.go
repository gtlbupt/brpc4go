package brpc

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Socket struct {
	sync.RWMutex
	writeQueue chan *RPCRequest

	closed chan bool
	codec  ProtocolCodec
	conn   net.Conn

	// stubs
	stubs map[int64]*RPCStub
}

func (s *Socket) WriteAsync(req *RPCRequest) {
	s.writeQueue <- req
}

func (s *Socket) WriteGoroutine() {
	var stopped = false

	for !stopped {
		select {
		case <-s.closed:
			stopped = true
		case req := <-s.writeQueue:
			var data, err = s.codec.Encode(req)
			if err != nil {
			}
			s.Lock()
			s.conn.Write(data)
			s.Unlock()
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
		var resp = RPCResponse{}
		err := s.codec.Decode(s.conn, &resp)
		if err != nil {
			break
		}

		var id = resp.CorelationId
		s.Lock()
		stub := s.stubs[id]
		delete(s.stubs, id)
		s.Unlock()
		switch {
		case stub == nil:
			continue
		default:
			stub.Done()
		}
	}

	if err == io.EOF {
	}
	s.closed <- true
}

func NewSocket(conn net.Conn, codec ProtocolCodec) *Socket {
	return &Socket{
		codec: codec,
		conn:  conn,
	}
}

func CreateSocket(node ServerNode, codec ProtocolCodec, options *ChannelOptions) (*Socket, error) {

	if options.IsTLS() {
		return nil, fmt.Errorf("(TLS) not support now")
	} else {
		var addr = fmt.Sprintf("%s:port", node.Ip, node.Port)
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
