package brpc

import (
	"context"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"sync"
	"sync/atomic"
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

type CallId struct {
}

type Controller struct {
	ctx    context.Context
	callId CallId
	err    error
}

func (cntl *Controller) SetFailed(err error) {
	cntl.err = err
}

func (cntl *Controller) Failed() bool {
	return cntl.err != nil
}

type MethodDescriptor struct {
	service string
	method  string
}

func (md *MethodDescriptor) ServiceName() string {
	return md.service
}

func (md *MethodDescriptor) MethodName() string {
	return md.method
}

type RPCRequest struct {
	ServiceAndMethod string
	corelation_id    int64
}

type RPCResponse struct {
	CorelationId int64
	Data         []byte
}

type RPCDoneFunctor func(args interface{}) error

type RPCDone struct {
	done     chan bool
	callback *RPCDoneFunctor
}

func (d *RPCDone) Done() {
	close(d.done)
}

func NewDefaultRPCDone() *RPCDone {
	return &RPCDone{}
}

func NewRPCDone(done chan bool, callback *RPCDoneFunctor) *RPCDone {
	return &RPCDone{
		done:     done,
		callback: callback,
	}
}

type RPCStub struct {
	md            MethodDescriptor
	correlationId int64
	cntl          *Controller
	request       proto.Message
	response      proto.Message
	done          RPCDone
}

func (stub *RPCStub) MD() *MethodDescriptor {
	return &stub.md
}

func (stub *RPCStub) CorrelationId() int64 {
	return stub.correlationId
}

func (stub *RPCStub) Controller() *Controller {
	return stub.cntl
}

func (stub *RPCStub) Request() proto.Message {
	return stub.request
}

func (stub *RPCStub) Response() proto.Message {
	return stub.response
}

func (s *RPCStub) Done() {
	s.done.Done()
}

type RPCStubMgr struct {
	next_id int64

	sync.RWMutex // guard RPCStubs
	stubs        map[int64]*RPCStub
}

func (mgr *RPCStubMgr) NextID() int64 {
	return atomic.AddInt64(&mgr.next_id, 1)
}

func (mgr *RPCStubMgr) Put(id int64, stub *RPCStub) {
	mgr.Lock()
	defer mgr.Unlock()

	mgr.put(id, stub)
}

func (mgr *RPCStubMgr) put(id int64, stub *RPCStub) {
	mgr.stubs[id] = stub
}

func (mgr *RPCStubMgr) Get(id int64) *RPCStub {
	mgr.Lock()
	defer mgr.Unlock()

	return mgr.get(id)
}

func (mgr *RPCStubMgr) get(id int64) *RPCStub {
	return mgr.stubs[id]
}

func NewRPCStubMgr() *RPCStubMgr {
	return &RPCStubMgr{
		stubs: make(map[int64]*RPCStub),
	}
}

var gRPCStubMgr = NewRPCStubMgr()

func NewRPCStub(
	md *MethodDescriptor,
	cntl *Controller,
	request proto.Message,
	response proto.Message,
	done chan bool) *RPCStub {

	var id = gRPCStubMgr.NextID()
	var stub = &RPCStub{
		md:            *md,
		correlationId: id,
		cntl:          cntl,
		request:       request,
		response:      response,
		done:          *NewRPCDone(done, nil),
	}
	gRPCStubMgr.Put(id, stub)
	return stub
}

func IsSupportProtocolOpt(protocol string, conn_type string) error {
	return nil
}
