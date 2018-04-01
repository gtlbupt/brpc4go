package brpc

import (
	baidu_std "./src/protocol"
	"context"
	"errors"
	proto "github.com/golang/protobuf/proto"
	"io"
	"sync"
	"time"
)

type ChannelOptions struct {
	addr               string
	lb                 string
	connect_timeout_ms time.Duration
	timeout_ms         time.Duration
	max_retry          int
}

type Channel struct {
	options ChannelOptions
}

func NewChannel(options ChannelOptions) *Channel {
	return &Channel{
		options: options,
	}
}

type MethodDescriptor struct {
	service string
	method  string
}

type Controller struct {
	ctx context.Context
}

func (c *Channel) CallMethod(
	md MethodDescriptor,
	cntl Controller,
	request *proto.Message,
	response *proto.Message,
	done chan<- struct{}) {
}

type ClientCodec interface {
	WriteRequest(*Request, interface{}) error
	ReadResponseHeader(*Response) error
	ReadResponseBody(interface{}) error

	Close() error
}

type BaiduStdClientCodec struct {
	r        io.Reader
	rwc      io.ReadWriteCloser
	nextSeq  uint64
	resMap   sync.Map
	currResp *baidu_std.BaiduRpcStdProtocol
	closed   bool
}

func (codec *BaiduStdClientCodec) WriteRequest(req *Request, v interface{}) error {
	return nil
}

func (codec *BaiduStdClientCodec) ReadResponseHeader(r *Response) error {
	var res = &baidu_std.BaiduRpcStdProtocol{}

	var l = res.Header.GetHeaderLen()
	var buf = make([]byte, l)
	if n, err := io.ReadAtLeast(codec.r, buf, l); err != nil || n < l {
		return err
	}

	if err := res.Header.Unmarshal(buf); err != nil {
		return err
	}

	var metaSize = res.Header.GetMetaSize()
	buf = make([]byte, metaSize)
	if n, err := io.ReadAtLeast(codec.r, buf, metaSize); err != nil || n < metaSize {
		return err
	}

	if err := proto.Unmarshal(buf, &res.Meta); err != nil {
		return err
	}

	response := res.Meta.GetResponse()
	if response == nil {
		return errors.New("Bad Response")
	}

	return nil
}

func (codec *BaiduStdClientCodec) ReadResponseBody(interface{}) error {
	return nil
}
