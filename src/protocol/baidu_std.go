package baidu_std

import (
	"bytes"
	"errors"
	proto "github.com/golang/protobuf/proto"
)

/**
 * | Header    |  Body									|
 *   Header => |---PRPC---|---BodySize---|---MetaSize---|
 *   Body   => | MetaData | Data | Attachment |
 */

type BaiduRpcStdProtocol struct {
	Header BaiduRpcStdProtocolHeader
	Meta   RpcMeta
	Data   []byte
}

func (p *BaiduRpcStdProtocol) Marshal() ([]byte, error) {
	var s [][]byte
	var sep []byte

	p.Header.SetBodySize(len(p.Data))

	metaBuf, err := proto.Marshal(&p.Meta)
	if err != nil {
		return nil, err
	}
	p.Header.SetMetaSize(len(metaBuf))

	headerBuf, err := p.Header.Marshal()
	if err != nil {
		return nil, err
	}

	s = append(s, headerBuf)
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

func makeTestRpcRequestMeta() RpcRequestMeta {
	var serviceName = "LogicService"
	var methodName = "JsonCmd"
	var logid = int64(102340)
	var traceId = int64(12345)
	var spanId = int64(23456)
	var parentSpanId = int64(4895445)
	var req = RpcRequestMeta{
		ServiceName:  proto.String(serviceName),
		MethodName:   proto.String(methodName),
		LogId:        proto.Int64(logid),
		TraceId:      proto.Int64(traceId),
		SpanId:       proto.Int64(spanId),
		ParentSpanId: proto.Int64(parentSpanId),
	}
	return req
}
func MakeTestRequest() []byte {
	var req = makeTestRpcRequestMeta()

	var rpcMeta = RpcMeta{
		Request: &req,
	}
	var p = &BaiduRpcStdProtocol{
		Meta: rpcMeta,
		Data: []byte("{}"),
	}
	var data, _ = p.Marshal()
	return data
}
