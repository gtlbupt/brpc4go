// A clone of brpc which implemented by golang
package brpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

type ProtocolType string

const (
	BADIU_STD ProtocolType = "BAIDU_STD"
)

type ServerOptions struct {
	Protocol string
	Addr     string
	logger   log.Logger
}

func (options *ServerOptions) Valid() bool {
	return true
}

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  int64
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

// Request is a header written before every RPC call. It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
type Request struct {
	ServiceMethod string   // format: "Service.Method"
	Seq           uint64   // sequence number by client
	next          *Request // for free list in Server
}

// Response is a header written before every RPC return. It is used intenrnally
// but documented here as an aid to debugging, such as when anylyzing
// network traffic.
type Response struct {
	ServiceMethod string    // echoes that of the Request
	Seq           uint64    // echoes that of the Request
	Error         string    // error, if any
	next          *Response // for free list in Server
}

// Server represents an RPC Server
type Server struct {
	serviceMap sync.Map // map[string]*service

	reqLock sync.Mutex // protects freeReq
	freeReq *Request

	respLock sync.Mutex // protects freeResp
	freeResp *Response

	options ServerOptions
	closed  chan struct{}
}

// NewServer returns a new Server.
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

// @DONE.
func (srv *Server) Register(rcvr interface{}) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {
		return errors.New("Bad Service Name")
	}
	s.name = sname

	// install the methods

	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		return errors.New("Type Has Not Export Method")
	}

	if _, dup := srv.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("rpc: service already defined")
	}

	return nil
}

func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		// Method must be exported
		if method.PkgPath != "" {
			continue
		}

		// Method needs three ins: receiver, *args, *replay
		if mtype.NumIn() != 3 {
			continue
		}

		argType := mtype.In(1)

		if !isExportedOrBuiltinType(argType) {
			continue
		}

		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}

		// ReplyType Must be exported
		if !isExportedOrBuiltinType(replyType) {
			continue
		}

		if mtype.NumOut() != 1 {
			continue
		}

		if returnType := mtype.Out(0); returnType != typeOfError {
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

// A value sent as a placeholder for the server's response value when the server
// receives an invalid request. Is is never decoded by the client since the Response
// contains an error when it is used.
var invalidRequest = struct{}{}

func (s *Server) sendResponse(sending *sync.Mutex, req *Request, reply interface{}, codec ServerCodec, errmsg string) {
	resp := s.getResponse()
	// Encode the response header
	resp.ServiceMethod = req.ServiceMethod
	if errmsg != "" {
		resp.Error = errmsg
		reply = invalidRequest
	}
	resp.Seq = req.Seq
	sending.Lock()
	err := codec.WriteResponse(resp, reply)
	if err != nil {
		log.Println("rpc: writing response:", err)
	}
	sending.Unlock()

	s.freeResponse(resp)
}

func (m *methodType) NumCalls() int64 {
	return atomic.LoadInt64(&m.numCalls)
}

func (s *service) call(server *Server, sending *sync.Mutex, wg *sync.WaitGroup, mtype *methodType, req *Request, argv, replyv reflect.Value, codec ServerCodec) {
	if wg != nil {
		defer wg.Done()
	}

	atomic.AddInt64(&mtype.numCalls, 1)

	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	errMsg := ""
	if errInter != nil {
		errMsg = errInter.(error).Error()
	}
	server.sendResponse(sending, req, replyv.Interface(), codec, errMsg)
	server.freeRequest(req)
}

type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *gobServerCodec) ReadRequestHeader(r *Request) error {
	return c.dec.Decode(r)
}

func (c *gobServerCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobServerCodec) WriteResponse(r *Response, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *gobServerCodec) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

// ServeConn runs the Server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection. To Use an alternate codec, use ServeCodec
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	buf := bufio.NewWriter(conn)
	srv := &gobServerCodec{
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
	}
	server.ServeCodec(srv)
}

// ServerCodec is like ServeConn but uses the specified codec to
// decode requests and encode responses.
func (server *Server) ServeCodec(codec ServerCodec) {
	sending := new(sync.Mutex)
	var wg sync.WaitGroup

	for {
		service, mtype, req, argv, replyv, keepReading, err := server.readRequest(codec)
		if err != nil {
			if !keepReading {
				break
			}

			// send a response if we actually managed to read a header
			if req != nil {
				server.sendResponse(sending, req, invalidRequest, codec, err.Error())
				server.freeRequest(req)
			}
			continue
		}
		wg.Add(1)
		go service.call(server, sending, &wg, mtype, req, argv, replyv, codec)
	}

	wg.Wait()
	codec.Close()
}

// ServeRequest is like ServeCodec but synchronously serves a single request.
// It does not close the codec upon completion.
func (server *Server) ServeRequest(codec ServerCodec) error {
	sending := new(sync.Mutex)
	service, mtype, req, argv, replyv, keepReading, err := server.readRequest(codec)
	if err != nil {
		if !keepReading {
			return err
		}

		if req != nil {
			server.sendResponse(sending, req, invalidRequest, codec, err.Error())
			server.freeRequest(req)
		}
		return err
	}
	service.call(server, sending, nil, mtype, req, argv, replyv, codec)
	return nil
}

func (server *Server) getRequest() *Request {
	server.reqLock.Lock()
	req := server.freeReq
	if req == nil {
		req = new(Request)
	} else {
		server.freeReq = req.next
		*req = Request{}
	}
	server.reqLock.Unlock()
	return req
}

func (server *Server) freeRequest(req *Request) {
	server.reqLock.Lock()
	req.next = server.freeReq
	server.freeReq = req
	server.reqLock.Unlock()
}

func (server *Server) getResponse() *Response {
	server.respLock.Lock()
	resp := server.freeResp
	if resp == nil {
		resp = new(Response)
	} else {
		server.freeResp = resp.next
		*resp = Response{}
	}
	server.respLock.Unlock()
	return resp
}

func (server *Server) freeResponse(resp *Response) {
	server.respLock.Lock()
	resp.next = server.freeResp
	server.freeResp = resp
	server.respLock.Unlock()
}

func (server *Server) readRequest(codec ServerCodec) (service *service, mtype *methodType, req *Request, argv, replyv reflect.Value, keepReading bool, err error) {
	service, mtype, req, keepReading, err = server.readRequestHeader(codec)
	if err != nil {
		if !keepReading {
			return
		}
		codec.ReadRequestBody(nil)
		return
	}

	// Decode the argument value.
	argIsValue := false
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	// argv guarantedd to be a pointer now.
	if err = codec.ReadRequestBody(argv.Interface()); err != nil {
		return
	}

	if argIsValue {
		argv = argv.Elem()
	}

	replyv = reflect.New(mtype.ReplyType.Elem())

	switch mtype.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
	}

	return
}

func (server *Server) readRequestHeader(codec ServerCodec) (svc *service, mtype *methodType, req *Request, keepReading bool, err error) {
	// Grap the request header.
	req = server.getRequest()
	err = codec.ReadRequestHeader(req)
	if err != nil {
		req = nil
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}
		err = errors.New("rpc: server cannot decode request: " + err.Error())
		return
	}

	keepReading = true

	dot := strings.LastIndex(req.ServiceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc: service/Method request ill-formed: " + req.ServiceMethod)
		return
	}
	serviceName := req.ServiceMethod[:dot]
	methodName := req.ServiceMethod[dot+1:]

	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc: cann't find service " + req.ServiceMethod)
		return
	}

	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc: can't find method " + req.ServiceMethod)
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection. Accept blocks until the listener
// return a non-nil error. The caller typically invokes Accept in
// a go statement.
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return
		}
		go s.ServeConn(conn)
	}
}

// A ServerCodec implements reading of RPC requests and writing of
// RPC responses for the server side of an RPC session.
// The server calls ReadRequestHeader and ReadRequestBody in pairs
// to read requests from the connection, and it calls WriteResponse to
// Write a response back. The server calls Close when finished with the
// connection. ReadRequestBody may be called with a nil
// argument to force the body of the request to be read and discarded.
type ServerCodec interface {
	// ReadRequest*
	ReadRequestHeader(*Request) error
	ReadRequestBody(interface{}) error

	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(*Response, interface{}) error

	Close() error
}
