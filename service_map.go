package brpc

import (
	"./src/util"
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  int64
}

func (m *methodType) NumCalls() int64 {
	return atomic.LoadInt64(&m.numCalls)
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

func NewService() *service {
	return &service{
		method: make(map[string]*methodType),
	}
}

func (s *service) GetName() string {
	return s.name
}

func (s *service) Install(v interface{}) error {
	s.typ = reflect.TypeOf(v)
	s.rcvr = reflect.ValueOf(v)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {
		return errors.New("Bad Service Name")
	}
	s.name = sname

	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		return errors.New("Type Has Not Export Method")
	}

	return nil
}

func (s *service) call(server *Server, sending *sync.Mutex, wg *sync.WaitGroup, mtype *methodType, req *Request, argv, replyv reflect.Value, codec ServerCodec) {
	if wg != nil {
		defer wg.Done()
	}

	atomic.AddInt64(&mtype.numCalls, 1)

	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply
	var ctx = reflect.ValueOf(context.TODO())
	returnValues := function.Call([]reflect.Value{s.rcvr, ctx, argv})
	// The return value for the method is an error.
	replyv = returnValues[0]
	errInter := returnValues[1].Interface()
	errMsg := ""
	if errInter != nil {
		errMsg = errInter.(error).Error()
	}
	server.sendResponse(sending, req, replyv.Interface(), codec, errMsg)
	server.freeRequest(req)
}

func OldSuitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
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

		if !util.IsExportedOrBuiltinType(argType) {
			continue
		}

		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}

		// ReplyType Must be exported
		if !util.IsExportedOrBuiltinType(replyType) {
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

		// Method needs three ins: receiver, context.Context, *args,
		if mtype.NumIn() != 3 {
			continue
		}

		ctxType := mtype.In(1)
		if !util.IsContextType(ctxType) {
			continue
		}

		argType := mtype.In(2)

		if !util.IsExportedOrBuiltinType(argType) {
			continue
		}

		// ReplayType, ReturnType
		if mtype.NumOut() != 2 {
			continue
		}

		replyType := mtype.Out(0)
		if replyType.Kind() != reflect.Ptr {
			continue
		}

		// ReplyType Must be exported
		if !util.IsExportedOrBuiltinType(replyType) {
			continue
		}

		if returnType := mtype.Out(1); returnType != typeOfError {
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}
