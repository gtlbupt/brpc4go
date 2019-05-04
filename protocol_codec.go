package brpc

import (
	"fmt"
	"net"
)

type ProtocolCodec interface {
	Encode(*RPCRequest) ([]byte, error)
	Decode(net.Conn, *RPCResponse) error
}

type BaiduStdCodec struct {
}

func (c *BaiduStdCodec) Encode(req *RPCRequest) ([]byte, error) {
	return nil, nil
}

func (c *BaiduStdCodec) Decode(conn net.Conn, resp *RPCResponse) error {
	return nil
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
