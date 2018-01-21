// A clone of brpc which implemented by golang
package brpc

import (
	"errors"
	proto "github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

type ProtocolType string

const (
	BADIU_STD ProtocolType = "BAIDU_STD"
)

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	ArgType    reflect.Type
	ReplyType  reflect.Type
	numCalls   uint
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

type ServerOptions struct {
	Protocol string
	Addr     string
	logger   log.Logger
}

func (options *ServerOptions) Valid() bool {
	return true
}

type Server struct {
	// serviceMap sync.Map   // map[string]*service
	reqLock  sync.Mutex // protects freeReq
	respLock sync.Mutex // protects freeResp

	options  ServerOptions
	listener net.Listener
	closed   chan struct{}
}

func NewServer(options ServerOptions) *Server {
	var s = &Server{
		options: options,
		closed:  make(chan struct{}),
	}
	return s
}

func (s *Server) Init(options ServerOptions) error {
	if !options.Valid() {
		return errors.New("Bad Options")
	}
	s.options = options
	return nil
}

func (srv *Server) Register(rcvr interface{}) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {
		return errors.New("Bad Service Name")
	}
	return nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.options.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	go s.loop()
	return nil
}

func (s *Server) loop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(c net.Conn) {
	defer c.Close()
	for {
		var req = &BaiduStdRpcProtocol{}
		var headerLen = req.Header.HeaderLen()
		var buf = make([]byte, 0, headerLen)

		n, err := io.ReadAtLeast(c, buf, headerLen)
		if err != nil ||
			n != headerLen {
			return
		}
		if err := req.Header.Unmarshal(buf); err != nil {
			return
		}
		var bodySize = req.Header.BodySize
		buf = make([]byte, 0, bodySize)
		n, err = io.ReadAtLeast(c, buf, int(bodySize))
		if err != nil ||
			n != int(bodySize) {
			return
		}

		var metaSize = req.Header.MetaSize
		var meta = buf[:metaSize]
		if err := proto.Unmarshal(meta, &req.Meta); err != nil {
			return
		}
		/*
			var serviceName = req.Meta.GetRequest().GetServiceName()
			var methodName = req.Meta.GetRequest().GetMethodName()
		*/
	}

}

func (s *Server) Stop() error {
	return nil
}
