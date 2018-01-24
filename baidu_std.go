package brpc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"io"
	"sync"
	"sync/atomic"
)

/**
 * | Header   |  Body |
 * Header => |---PRPC---|---BodySize---|---MetaSize---|
 * Body   => | MetaData | Data | Attachment |
 */

const (
	BAIDU_STD_RPC_MSG_HEADER_LEN = 12
	BAIDU_STD_MAGIC_STRING       = "PRPC"
)

type BaiduRpcStdProtocolHeader struct {
	BodySize uint32
	MetaSize uint32
}

func (h *BaiduRpcStdProtocolHeader) GetHeaderLen() int {
	return BAIDU_STD_RPC_MSG_HEADER_LEN
}

func (h *BaiduRpcStdProtocolHeader) GetMetaSize() int {
	return int(h.MetaSize)
}

func (h *BaiduRpcStdProtocolHeader) GetBodySize() int {
	return int(h.BodySize)
}

func (h *BaiduRpcStdProtocolHeader) Marshal() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, BAIDU_STD_RPC_MSG_HEADER_LEN))

	binary.Write(buf, binary.LittleEndian, []byte(BAIDU_STD_MAGIC_STRING))
	binary.Write(buf, binary.BigEndian, h.BodySize)
	binary.Write(buf, binary.BigEndian, h.MetaSize)

	return buf.Bytes(), nil
}

func (h *BaiduRpcStdProtocolHeader) Unmarshal(data []byte) error {
	if len(data) < BAIDU_STD_RPC_MSG_HEADER_LEN {
		return errors.New("Bad RPC Header Length")
	}

	if string(data[:4]) != BAIDU_STD_MAGIC_STRING {
		return errors.New("Bad Magic String")
	}

	h.BodySize = binary.BigEndian.Uint32(data[4:8])
	h.MetaSize = binary.BigEndian.Uint32(data[8:12])

	return nil
}

type BaiduRpcStdProtocol struct {
	Header BaiduRpcStdProtocolHeader
	Meta   RpcMeta
	Data   []byte
}

func (p *BaiduRpcStdProtocol) Marshal() ([]byte, error) {
	var s [][]byte
	var sep []byte

	headerBuf, err := p.Header.Marshal()
	if err != nil {
		return nil, err
	}
	s = append(s, headerBuf)

	metaBuf, err := proto.Marshal(&p.Meta)
	if err != nil {
		return nil, err
	}
	s = append(s, metaBuf)

	s = append(s, p.Data)

	return bytes.Join(s, sep), nil
}

func (p *BaiduRpcStdProtocol) Unmarshal(data []byte) error {
	if err := p.Header.Unmarshal(data); err != nil {
		return err
	}

	if len(data) < p.Header.GetHeaderLen()+p.Header.GetMetaSize() {
		return errors.New("Bad Meta Len")
	}
	var start = p.Header.GetHeaderLen()
	var end = p.Header.GetHeaderLen() + p.Header.GetMetaSize()
	metaBuf := data[start:end]

	if err := proto.Unmarshal(metaBuf, &p.Meta); err != nil {
		return err
	}

	p.Data = data[end:]
	return nil
}

type BaiduStdServerCodec struct {
	rwc     io.ReadWriteCloser
	nextSeq uint64
	reqMap  sync.Map
	currReq *BaiduRpcStdProtocol
	closed  bool
}

func (c *BaiduStdServerCodec) AddRequest(req *BaiduRpcStdProtocol) uint64 {
	var seq = atomic.AddUint64(&c.nextSeq, 1)
	c.reqMap.Store(seq, req)
	c.currReq = req
	return seq
}

func (c *BaiduStdServerCodec) GetRequest(seq uint64) *BaiduRpcStdProtocol {
	reqInter, ok := c.reqMap.Load(seq)
	if !ok {
		return nil
	}

	if req, ok := reqInter.(*BaiduRpcStdProtocol); ok {
		return req
	} else {
		return nil
	}
}

func (c *BaiduStdServerCodec) ReadRequestHeader(r *Request) error {
	var req = &BaiduRpcStdProtocol{}

	var l = req.Header.GetHeaderLen()
	var buf = make([]byte, 0, l)
	if n, err := io.ReadAtLeast(c.rwc, buf, l); err != nil || n < l {
		return err
	}

	if err := req.Header.Unmarshal(buf); err != nil {
		return err
	}

	var metaSize = req.Header.GetMetaSize()
	buf = make([]byte, 0, metaSize)
	if n, err := io.ReadAtLeast(c.rwc, buf, metaSize); err != nil || n < metaSize {
		return err
	}

	if err := proto.Unmarshal(buf, &req.Meta); err != nil {
		return err
	}

	request := req.Meta.GetRequest()
	if request == nil {
		return errors.New("Bad Request")
	}
	r.ServiceMethod = fmt.Sprintf("%s.%s", request.ServiceName, request.MethodName)
	r.Seq = c.AddRequest(req)

	return nil
}

func (c *BaiduStdServerCodec) ReadRequestBody(request interface{}) error {
	var req = c.currReq
	var bodySize = req.Header.GetBodySize()
	var buf = make([]byte, 0, bodySize)
	if n, err := io.ReadAtLeast(c.rwc, buf, bodySize); err != nil || n < bodySize {
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
	var req = reqInter.(BaiduRpcStdProtocol)

	if !ok {
		return errors.New("Bad Response")
	}

	var respRpcMeta = RpcMeta{
		Response: &RpcResponseMeta{
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

	var respHeader = BaiduRpcStdProtocolHeader{
		BodySize: uint32(len(bodyBuf)),
		MetaSize: uint32(len(metaBuf)),
	}

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
}
