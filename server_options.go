package brpc

import (
	"log"
	"net"
)

type ProtocolType string

const (
	PROTO_BAIDU_STD ProtocolType = "BAIDU_STD"
)

type ServerOptions struct {
	Protocol ProtocolType
	Lis      net.Listener
	logger   log.Logger
}
