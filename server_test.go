package brpc

import (
	example "./example"
	baidu_std "./src/protocol"
	"context"
	"errors"
	"fmt"
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

		t.Run("ClientSend", func(t *testing.T) {
			clientSend(addr)
		})
		t.Run("ChannelCallMethod", func(t *testing.T) {
			var addr = "list://127.0.0.1:8080"
			var lb = "random"
			var options = NewChannelOptions()
			var channel = NewChannel()
			if err := channel.Init(addr, lb, options); err != nil {
				t.Errorf("Channel.Init(%s, %s, %v) = %v",
					addr, lb, options, err)
			}

			var service = "EchoService"
			var method = "Echo"
			var md = MethodDescriptor{
				service: service,
				method:  method,
			}
			var cntl = Controller{}
			var request = example.EchoRequest{
				Message: proto.String("Hello, World"),
			}
			var response = example.EchoResponse{}
			var done = make(chan bool)
			var err = channel.CallMethod(
				&md,
				&cntl,
				&request,
				&response,
				done)
			log.Printf("channel.CallMethod() = %v", err)
			<-done
			fmt.Printf("EchoResponse=%v", response)
		})
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
		req.Header.Unmarshal(buf)
		log.Printf("[header:%s]", req.Header.String())

		var metaSize = req.Header.GetMetaSize()
		buf = make([]byte, metaSize)
		if n, err := io.ReadAtLeast(conn, buf, metaSize); err != nil || n < metaSize {
			log.Printf("[err:%v]", err)
		}

		if err := proto.Unmarshal(buf, &req.Meta); err != nil {
			log.Printf("[err:%v]", err)
		}
		log.Printf("[meta:%v]", req.Meta)

		var bodySize = req.Header.GetBodySize()
		buf = make([]byte, bodySize)
		if n, err := io.ReadAtLeast(conn, buf, bodySize); err != nil || n < bodySize {
			log.Printf("[ReadMetaSize.err:%v]", err)
		}

		log.Printf("[body:%s]", string(buf))
		var echoRes = &example.EchoResponse{}
		if err := UnmarshalEchoResponse(buf, echoRes); err != nil {
			log.Printf("[UnmarshalEchoResponse.err:%v]", err)
		}
	}
	return nil, nil
}

func UnmarshalEchoResponse(data []byte, v interface{}) error {
	msg := v.(proto.Message)
	msg.Reset()

	err := proto.Unmarshal(data, msg)
	return err
}
