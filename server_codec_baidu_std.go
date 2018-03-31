package brpc

import (
	baidu_std "./src/protocol"
	"errors"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"io"
	"sync"
	"sync/atomic"
)

type BaiduStdServerCodec struct {
	r       io.Reader
	rwc     io.ReadWriteCloser
	nextSeq uint64
	reqMap  sync.Map
	currReq *baidu_std.BaiduRpcStdProtocol
	closed  bool
}

func (c *BaiduStdServerCodec) AddRequest(req *baidu_std.BaiduRpcStdProtocol) uint64 {
	var seq = atomic.AddUint64(&c.nextSeq, 1)
	c.reqMap.Store(seq, req)
	c.currReq = req
	return seq
}

func (c *BaiduStdServerCodec) GetRequest(seq uint64) *baidu_std.BaiduRpcStdProtocol {
	reqInter, ok := c.reqMap.Load(seq)
	if !ok {
		return nil
	}

	if req, ok := reqInter.(*baidu_std.BaiduRpcStdProtocol); ok {
		return req
	} else {
		return nil
	}
}

func (c *BaiduStdServerCodec) ReadRequestHeader(r *Request) error {
	var req = &baidu_std.BaiduRpcStdProtocol{}

	var l = req.Header.GetHeaderLen()
	var buf = make([]byte, l)
	if n, err := io.ReadAtLeast(c.r, buf, l); err != nil || n < l {
		return err
	}

	if err := req.Header.Unmarshal(buf); err != nil {
		return err
	}

	var metaSize = req.Header.GetMetaSize()
	buf = make([]byte, metaSize)
	if n, err := io.ReadAtLeast(c.r, buf, metaSize); err != nil || n < metaSize {
		return err
	}

	if err := proto.Unmarshal(buf, &req.Meta); err != nil {
		return err
	}

	request := req.Meta.GetRequest()
	if request == nil {
		return errors.New("Bad Request")
	}
	r.ServiceMethod = fmt.Sprintf("%s.%s",
		request.GetServiceName(), request.GetMethodName())
	r.Seq = c.AddRequest(req)

	return nil
}

func (c *BaiduStdServerCodec) ReadRequestBody(request interface{}) error {
	var req = c.currReq
	var bodySize = req.Header.GetBodySize()
	var buf = make([]byte, bodySize)
	if n, err := io.ReadAtLeast(c.r, buf, bodySize); err != nil || n < bodySize {
		return err
	}

	var msg proto.Message
	if err := proto.Unmarshal(buf, msg); err != nil {
		return err
	}

	return nil
}

func (c *BaiduStdServerCodec) WriteResponse(resp *Response, body interface{}) error {
	var seq = resp.Seq
	var reqInter, ok = c.reqMap.Load(seq)
	var req = reqInter.(baidu_std.BaiduRpcStdProtocol)

	if !ok {
		return errors.New("Bad Response")
	}

	var respRpcMeta = baidu_std.RpcMeta{
		Response: &baidu_std.RpcResponseMeta{
			ErrorCode: proto.Int32(0),
		},
		CorrelationId: req.Meta.CorrelationId,
	}

	metaBuf, err := proto.Marshal(&respRpcMeta)
	if err != nil {
		return err
	}

	var msg = body.(proto.Message)
	bodyBuf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	var respHeader = baidu_std.BaiduRpcStdProtocolHeader{}
	respHeader.SetBodySize(len(bodyBuf))
	respHeader.SetMetaSize(len(metaBuf))

	headBuf, err := respHeader.Marshal()
	if err != nil {
		return err
	}

	c.rwc.Write(headBuf)
	c.rwc.Write(metaBuf)
	c.rwc.Write(bodyBuf)
	return nil
}

func (c *BaiduStdServerCodec) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.rwc.Close()
	return nil
}
