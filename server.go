// A clone of brpc which implemented by golang
package brpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// Server represents an RPC Server
type Server struct {
	serviceMap sync.Map // map[string]*service

	freeReq RequestList

	freeResp ResponseList

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

func (srv *Server) Close() {
	srv.options.Lis.Close()
	close(srv.closed)
}

// @DONE.
func (srv *Server) AddService(rcvr interface{}) error {
	var service = NewService()
	service.Install(rcvr)

	if _, dup := srv.serviceMap.LoadOrStore(service.GetName(), service); dup {
		return errors.New("rpc: service already defined")
	}

	return nil
}

func (srv *Server) Start() error {
	go srv.startImpl()
	return nil
}

func (srv *Server) startImpl() {
	var lis = srv.options.Lis
	for {
		conn, err := lis.Accept()
		if err != nil {
			return
		}
		if srv.options.Protocol == PROTO_BAIDU_STD {
			var codec = &BaiduStdServerCodec{
				rwc: conn,
			}
			go srv.ServeCodec(codec)
		} else {
			go srv.ServeConn(conn)
		}
	}
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

func (srv *Server) ServeCodec(codec ServerCodec) {
	var server = srv
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		service, mtype, req, argv, replyv, keepReading, err := server.readRequest(codec)
		if err != nil {
			if !keepReading {
				break
			}
			if req != nil {
				server.sendResponse(
					sending, req, invalidRequest, codec, err.Error())
				server.freeRequest(req)
			}
			continue
		}
		wg.Add(1)
		go service.call(server, sending, wg, mtype, req, argv, replyv, codec)
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
	return server.freeReq.Get()
}

func (server *Server) freeRequest(req *Request) {
	server.freeReq.Put(req)
}

func (server *Server) getResponse() *Response {
	return server.freeResp.Get()
}

func (server *Server) freeResponse(resp *Response) {
	server.freeResp.Put(resp)
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
