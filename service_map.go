package brpc

import (
	"errors"
	"reflect"
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
	return new(service)
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
