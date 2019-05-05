package brpc

import (
	baidu_std "./src/protocol"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
)

type ProtocolCodec interface {
	Encode(*RPCStub) ([]byte, error)
	Decode(net.Conn) (*RPCStub, error)
}

type BaiduStdCodec struct {
}

func (c *BaiduStdCodec) Encode(stub *RPCStub) ([]byte, error) {
	bodyBuf, err := proto.Marshal(stub.Request())
	if err != nil {
		return nil, err
	}
	var meta = baidu_std.RpcMeta{
		Request: &baidu_std.RpcRequestMeta{
			ServiceName: proto.String(stub.MD().ServiceName()),
			MethodName:  proto.String(stub.MD().MethodName()),
		},
		CorrelationId: proto.Int64(stub.CorrelationId()),
	}

	var protocol = baidu_std.BaiduRpcStdProtocol{
		Meta: meta,
		Data: bodyBuf,
	}

	data, err := protocol.Marshal()
	return data, err
}

func (c *BaiduStdCodec) Decode(conn net.Conn) (*RPCStub, error) {

	var req = &baidu_std.BaiduRpcStdProtocol{}

	var l = req.Header.GetHeaderLen()
	var buf = make([]byte, l)
	if n, err := io.ReadAtLeast(conn, buf, l); err != nil || n < l {
		log.Printf("[err:%v]", err)
		return nil, err
	}
	req.Header.Unmarshal(buf)
	log.Printf("[header:%s]", req.Header.String())

	var metaSize = req.Header.GetMetaSize()
	buf = make([]byte, metaSize)
	if n, err := io.ReadAtLeast(conn, buf, metaSize); err != nil || n < metaSize {
		log.Printf("[err:%v]", err)
		return nil, err
	}

	if err := proto.Unmarshal(buf, &req.Meta); err != nil {
		log.Printf("[err:%v]", err)
		return nil, err
	}
	log.Printf("[meta:%v]", req.Meta)

	var bodySize = req.Header.GetBodySize()
	buf = make([]byte, bodySize)
	if n, err := io.ReadAtLeast(conn, buf, bodySize); err != nil || n < bodySize {
		log.Printf("[ReadMetaSize.err:%v]", err)
		return nil, err
	}

	var id = req.Meta.GetCorrelationId()
	var stub = gRPCStubMgr.Get(id)
	if stub == nil {
		return nil, nil
	}

	log.Printf("[body:%s]", string(buf))
	if err := proto.Unmarshal(buf, stub.Response()); err != nil {
		log.Printf("[UnmarshalEchoResponse.err:%v]", err)
		return nil, err
	}
	return stub, nil
}

func NewBaiduStdCodec() *BaiduStdCodec {
	return &BaiduStdCodec{}
}

func NewProtocolCodec(protocol string) (ProtocolCodec, error) {
	if protocol == "baidu_std" {
		return NewBaiduStdCodec(), nil
	} else {
		return nil, fmt.Errorf("(%s) don't support", protocol)
	}
}
