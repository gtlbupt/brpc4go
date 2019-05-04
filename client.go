package brpc

import (
	baidu_std "./src/protocol"
	"errors"
	proto "github.com/golang/protobuf/proto"
	"io"
	"sync"
)

type ClientCodec interface {
	// WriteRequest must be safe for concurrent use by multiple goroutine
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
