package brpc

import (
	"net"
	"testing"
)

func TestNewServer(t *testing.T) {
	var options = ServerOptions{
		Protocol: PROTO_BAIDU_STD,
		Addr:     ":8080",
	}
	var srv = NewServer(options)

	if srv == nil {
		t.Errorf("NewServer(%v) = nil", options)
	}
}

type Arith int

type Args struct {
	A int
	B int
}

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func TestRegister(t *testing.T) {
	var options = ServerOptions{
		Protocol: PROTO_BAIDU_STD,
		Addr:     ":8080",
	}
	var srv = NewServer(options)

	var arith Arith
	srv.Register(arith)

	var listener, err = net.Listen("tcp", "localhost:8080")
	if err != nil {
		t.Error(err)
	}
	srv.Accept(listener)
}
