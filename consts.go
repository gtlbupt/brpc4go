package brpc

import (
	"fmt"
	proto "github.com/golang/protobuf/proto"
)

const (
	ProtocolDefault  = ""
	ProtocolBaiduStd = "baidu_std"
	ProtocolHttp     = "http"
	ProtocolH2c      = "h2c"
	ProtocolH2       = "h2"
)

const (
	ConnTypeDefault = ""
	ConnTypeSingle  = "single"
	ConnTypePooled  = "pooled"
	ConnTypeShort   = "short"
)

// support load balance algorithm
const (
	LoadBalanceRandom     = "random"
	LoadBalanceRoundRobin = "rr"
	LoadBalanceWeightRR   = "wrr"
	LoadBalanceMurMurHash = "c_murmurhash"
	LoadBalanceCMd5Hash   = "c_md5"
)

const (
	NameServiceBNS  = "bns"
	NameServiceFile = "file"
	NameServiceList = "list"
)

type AdaptiveProtocolType string
type AdaptiveConnectionType string

type ServerNode struct {
	Ip   string
	Port int
	Tag  string
}

func (n *ServerNode) ToString() string {
	return fmt.Sprintf("%s:%d:%s", n.Ip, n.Port, n.Tag)
}

type MethodDescriptor struct {
	service string
	method  string
}

type RPCRequest struct {
	ServiceAndMethod string
	corelation_id    int64
}

type RPCResponse struct {
	CorelationId int64
	Data         []byte
}

type RPCStub struct {
	ServiceAndMethod string
	CorelationId     int64
	Request          *proto.Message
	Response         *proto.Message
	done             chan bool
}

func (s *RPCStub) Done() {
	s.done <- true
}

func IsSupportProtocolOpt(protocol string, conn_type string) error {
	return nil
}
