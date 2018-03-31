package brpc

import (
	example "./example"
	baidu_std "./src/protocol"
	"context"
	"errors"
	proto "github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

type Args struct {
	A int
	B int
}

type Quotient struct {
	Quo int
	Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(ctx context.Context, args *Args) (quo *Quotient, err error) {
	if args.B == 0 {
		err = errors.New("divide by zero")
		return
	}
	quo = &Quotient{}

	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B

	return
}

func newFakeServer(addr string) (*Server, error) {
	var protocol = "tcp"
	ln, err := net.Listen(protocol, addr)
	if err != nil {
		return nil, err
	}

	var options = ServerOptions{
		Protocol: PROTO_BAIDU_STD,
		Lis:      ln,
	}
	var srv = NewServer(options)

	return srv, nil
}

type EchoService struct {
}

func (s *EchoService) Echo(ctx context.Context, req *example.EchoRequest) (*example.EchoResponse, error) {
	log.Printf("[message:%s]", req.GetMessage())
	var resp = &example.EchoResponse{}
	resp.Message = req.Message
	return resp, nil
}

func TestNewServer(t *testing.T) {
	var addr = ":8080"
	var srv, err = newFakeServer(addr)
	if err != nil {
		t.Errorf("newFakeServer() = %v", err)
	}
	defer srv.Close()
	if srv == nil {
		t.Errorf("NewServer() = nil")
	}
	t.Run("AddService", func(t *testing.T) {
		var serviceImpl = new(EchoService)
		var err = srv.AddService(serviceImpl)
		if err != nil {
			t.Errorf("server.AddService(AddServiceImpl) = %v", err)
		}
	})
	t.Run("AddServiceDup", func(t *testing.T) {
		var serviceImpl = new(EchoService)
		var err = srv.AddService(serviceImpl)
		if err == nil {
			t.Errorf("server.AddService(AddServiceImpl) = %v, expect = DupService", err)
		}
	})
	t.Run("Start", func(t *testing.T) {
		var err = srv.Start()
		if err != nil {
			t.Errorf("server.Start() = %v", err)
		}
		clientSend(addr)
		time.Sleep(1 * time.Second)
	})
}

func clientSend(addr string) ([]byte, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	var serviceName = "EchoService"
	var methodName = "Echo"
	body, err := proto.Marshal(
		&example.EchoResponse{
			Message: proto.String("Hello, World!!!"),
		})
	var data = baidu_std.MakeTestRequest(
		serviceName,
		methodName,
		body,
	)
	n, err := conn.Write(data)
	if err != nil || n != len(data) {
		return nil, err
	}
	{
		var req = &baidu_std.BaiduRpcStdProtocol{}

		var l = req.Header.GetHeaderLen()
		var buf = make([]byte, l)
		if n, err := io.ReadAtLeast(conn, buf, l); err != nil || n < l {
			log.Printf("[err:%v]", err)
		}
		var metaSize = req.Header.GetMetaSize()
		buf = make([]byte, metaSize)
		if n, err := io.ReadAtLeast(conn, buf, metaSize); err != nil || n < metaSize {
			log.Printf("[err:%v]", err)
		}

		if err := proto.Unmarshal(buf, &req.Meta); err != nil {
			log.Printf("[err:%v]", err)
		}
		log.Printf("[meta:%v]", req.Meta)
	}
	return nil, nil
}
