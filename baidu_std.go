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

type BaiduStdRpcHeader struct {
	BodySize uint32
	MetaSize uint32
}

func (h *BaiduStdRpcHeader) GetHeaderLen() int {
	return BAIDU_STD_RPC_MSG_HEADER_LEN
}

func (h *BaiduStdRpcHeader) GetMetaSize() int {
	return int(h.MetaSize)
}

func (h *BaiduStdRpcHeader) GetBodySize() int {
	return int(h.BodySize)
}

func (h *BaiduStdRpcHeader) Marshal() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, BAIDU_STD_RPC_MSG_HEADER_LEN))

	binary.Write(buf, binary.LittleEndian, []byte(BAIDU_STD_MAGIC_STRING))
	binary.Write(buf, binary.BigEndian, h.BodySize)
	binary.Write(buf, binary.BigEndian, h.MetaSize)

	return buf.Bytes(), nil
}

func (h *BaiduStdRpcHeader) Unmarshal(data []byte) error {
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

type BaiduStdRpcProtocol struct {
	Header BaiduStdRpcHeader
	Meta   RpcMeta
	Data   []byte
}

type BaiduStdServerCodec struct {
	rwc     io.ReadWriteCloser
	nextSeq uint64
	reqMap  sync.Map
	currReq *BaiduStdRpcProtocol
}

func (c *BaiduStdServerCodec) AddRequest(req *BaiduStdRpcProtocol) uint64 {
	var seq = atomic.AddUint64(&c.nextSeq, 1)
	c.reqMap.Store(seq, req)
	c.currReq = req
	return seq
}

func (c *BaiduStdServerCodec) ReadRequestHeader(r *Request) error {
	var req = &BaiduStdRpcProtocol{}

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
	var req = reqInter.(BaiduStdRpcProtocol)

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

	var respHeader = BaiduStdRpcHeader{
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
	return nil
}
