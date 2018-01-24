package brpc

import (
	"bytes"
	proto "github.com/golang/protobuf/proto"
	"testing"
)

func TestBaiduRpcStdProtocolHeader(t *testing.T) {
	var metaSize = 1024
	var bodySize = 2048
	var header = BaiduRpcStdProtocolHeader{
		MetaSize: uint32(metaSize),
		BodySize: uint32(bodySize),
	}

	t.Run("GetHeaderLen", func(t *testing.T) {
		if header.GetHeaderLen() != BAIDU_STD_RPC_MSG_HEADER_LEN {
			t.Errorf("h.GetHeaderLen() = %d, expect = %d",
				header.GetHeaderLen(), BAIDU_STD_RPC_MSG_HEADER_LEN)
		}
	})

	t.Run("GetMetaSize", func(t *testing.T) {
		if header.GetMetaSize() != metaSize {
			t.Errorf("h.GetMetaSize() = %d, expect = %d",
				header.GetMetaSize(), metaSize)
		}
	})

	t.Run("GetBodySize", func(t *testing.T) {
		if header.GetBodySize() != bodySize {
			t.Errorf("h.GetBodySize() = %d, expect = %d",
				header.GetBodySize(), bodySize)
		}
	})

	t.Run("MarshalUnmarshal", func(t *testing.T) {
		var metaSize = 1024
		var bodySize = 2048
		var header = BaiduRpcStdProtocolHeader{
			MetaSize: uint32(metaSize),
			BodySize: uint32(bodySize),
		}
		buf, err := header.Marshal()
		if err != nil || len(buf) != header.GetHeaderLen() {
			t.Errorf("(%v).Marshal() = %v",
				header, err)
		}

		var target = &BaiduRpcStdProtocolHeader{}
		if err := target.Unmarshal(buf); err != nil {
			t.Errorf("(%v).Unmarshl(%v) = %v",
				target, buf, err)
		}

		if target.GetMetaSize() != metaSize {
			t.Errorf("(%v).GetMetaSize() = %d, expect = %d",
				target, target.GetMetaSize(), metaSize)
		}
		if target.GetBodySize() != bodySize {
			t.Errorf("(%v).GetBodySize() = %d, expect = %d",
				target, target.GetBodySize(), bodySize)
		}
	})
}

func TestRpcRequestMeta(t *testing.T) {
	t.Run("Field", func(t *testing.T) {
		var serviceName = "LogicService"
		var methodName = "JsonCmd"
		var logid = int64(102340)
		var traceId = int64(12345)
		var spanId = int64(23456)
		var parentSpanId = int64(4895445)
		var req = &RpcRequestMeta{
			ServiceName:  proto.String(serviceName),
			MethodName:   proto.String(methodName),
			LogId:        proto.Int64(logid),
			TraceId:      proto.Int64(traceId),
			SpanId:       proto.Int64(spanId),
			ParentSpanId: proto.Int64(parentSpanId),
		}
		if req.GetServiceName() != serviceName ||
			req.GetMethodName() != methodName ||
			req.GetLogId() != logid ||
			req.GetTraceId() != traceId ||
			req.GetSpanId() != spanId ||
			req.GetParentSpanId() != parentSpanId {
			t.Errorf("req(%v).Field() Field", req)
		}

		{
			buf, err := proto.Marshal(req)
			if err != nil {
				t.Errorf("proto.Marshal(%v)", req)
			}
			var target = &RpcRequestMeta{}
			if err := proto.Unmarshal(buf, target); err != nil {
				t.Errorf("proto.Unmarshal(%v) failed = %v", buf, err)
			}

			if !rpcRequestMetaEqual(req, target) {
				t.Errorf("req(%v).Field() Field", req)
			}
		}
	})
}

func TestRpcResponseMeta(t *testing.T) {
	t.Run("ErrorCode/ErrorText", func(t *testing.T) {
		var errorCode = int32(1023)
		var errorText = "ErrorText"

		var resp = &RpcResponseMeta{
			ErrorCode: proto.Int32(errorCode),
			ErrorText: proto.String(errorText),
		}

		{
			if resp.GetErrorCode() != errorCode {
				t.Errorf("(%v).GetErrorCode() = %d, expect = %d",
					resp, resp.GetErrorCode(), errorCode)
			}

			if resp.GetErrorText() != errorText {
				t.Errorf("(%v).GetErrorText() = %s, expect = %s",
					resp, resp.GetErrorText(), errorText)

			}
		}

		buf, err := proto.Marshal(resp)
		if err != nil {
			t.Errorf("proto.Marshal(%v) = err(%v)",
				resp, err)
		}

		{
			var target = &RpcResponseMeta{}
			if err := proto.Unmarshal(buf, target); err != nil {
				t.Errorf("proto.Unmarshal() = err(%v)", err)
			}

			if resp.GetErrorCode() != errorCode {
				t.Errorf("(%v).GetErrorCode() = %d, expect = %d",
					resp, resp.GetErrorCode(), errorCode)
			}

			if resp.GetErrorText() != errorText {
				t.Errorf("(%v).GetErrorText() = %s, expect = %s",
					resp, resp.GetErrorText(), errorText)

			}
		}
	})
}

func TestBaiduStdRpcProtocolRpcMeta(t *testing.T) {
	t.Run("Request", func(t *testing.T) {
		var serviceName = "LogicService"
		var methodName = "JsonCmd"
		var logid = int64(102340)
		var traceId = int64(12345)
		var spanId = int64(23456)
		var parentSpanId = int64(4895445)
		var req = &RpcRequestMeta{
			ServiceName:  proto.String(serviceName),
			MethodName:   proto.String(methodName),
			LogId:        proto.Int64(logid),
			TraceId:      proto.Int64(traceId),
			SpanId:       proto.Int64(spanId),
			ParentSpanId: proto.Int64(parentSpanId),
		}

		var errorCode = int32(1023)
		var errorText = "ErrorText"

		var resp = &RpcResponseMeta{
			ErrorCode: proto.Int32(errorCode),
			ErrorText: proto.String(errorText),
		}

		var compressType = int32(1)
		var correlationId = int64(2)
		var attachmentSize = int32(3)

		var rpcMeta = &RpcMeta{
			Request:        req,
			Response:       resp,
			CompressType:   proto.Int32(compressType),
			CorrelationId:  proto.Int64(correlationId),
			AttachmentSize: proto.Int32(attachmentSize),
		}

		buf, err := proto.Marshal(rpcMeta)
		if err != nil {
			t.Errorf("proto.Marshal(%v) = %v", rpcMeta, err)
		}

		var target = &RpcMeta{}
		if err := proto.Unmarshal(buf, target); err != nil {
			t.Errorf("proto.Unmarshal(%v) = %v", buf, err)
		}

		if !rpcRequestMetaEqual(target.GetRequest(), rpcMeta.GetRequest()) ||
			!rpcResponseMetaEqual(target.GetResponse(), rpcMeta.GetResponse()) ||
			target.GetCompressType() != rpcMeta.GetCompressType() ||
			target.GetCorrelationId() != rpcMeta.GetCorrelationId() ||
			target.GetAttachmentSize() != rpcMeta.GetAttachmentSize() {
			t.Errorf("Unmarshal failed, src(%v) == dst(%v)",
				rpcMeta, target)
		}
	})
}

func TestBaiduStdRpcProtocol(t *testing.T) {
	var serviceName = "LogicService"
	var methodName = "JsonCmd"
	var logid = int64(102340)
	var traceId = int64(12345)
	var spanId = int64(23456)
	var parentSpanId = int64(4895445)
	var req = &RpcRequestMeta{
		ServiceName:  proto.String(serviceName),
		MethodName:   proto.String(methodName),
		LogId:        proto.Int64(logid),
		TraceId:      proto.Int64(traceId),
		SpanId:       proto.Int64(spanId),
		ParentSpanId: proto.Int64(parentSpanId),
	}

	var errorCode = int32(1023)
	var errorText = "ErrorText"

	var resp = &RpcResponseMeta{
		ErrorCode: proto.Int32(errorCode),
		ErrorText: proto.String(errorText),
	}

	var compressType = int32(1)
	var correlationId = int64(2)
	var attachmentSize = int32(3)

	t.Run("MarshalRequest", func(t *testing.T) {
		var rpcMeta = &RpcMeta{
			Request:        req,
			CompressType:   proto.Int32(compressType),
			CorrelationId:  proto.Int64(correlationId),
			AttachmentSize: proto.Int32(attachmentSize),
		}
		var data, err = proto.Marshal(rpcMeta)
		if err != nil {
			t.Errorf("RpcMeta.Request. err= %v", err)
		}

		var target = &RpcMeta{}
		if err := proto.Unmarshal(data, target); err != nil {
			t.Errorf("proto.Unmarshal(%v) = %v", data, err)
		}

	})
	t.Run("MarshalResponse", func(t *testing.T) {
		var rpcMeta = &RpcMeta{
			Response:       resp,
			CompressType:   proto.Int32(compressType),
			CorrelationId:  proto.Int64(correlationId),
			AttachmentSize: proto.Int32(attachmentSize),
		}
		var data, err = proto.Marshal(rpcMeta)
		if err != nil {
			t.Errorf("RpcMeta.Request. err= %v", err)
		}

		var target = &RpcMeta{}
		if err := proto.Unmarshal(data, target); err != nil {
			t.Errorf("proto.Unmarshal(%v) = %v", data, err)
		}
	})
}

func TestRpcRequestMetaV2(t *testing.T) {
	var src = &RpcRequestMeta{
		ServiceName: proto.String("LogicService"),
		MethodName:  proto.String("JsonCmd"),
		LogId:       proto.Int64(10240),
	}
	var data, err = proto.Marshal(src)
	if err != nil {
		t.Error(err)
	}

	var dst = &RpcRequestMeta{}
	if err := proto.Unmarshal(data, dst); err != nil {
		t.Error(err)
	}

	if src.GetServiceName() != dst.GetServiceName() ||
		src.GetMethodName() != dst.GetMethodName() ||
		src.GetLogId() != dst.GetLogId() {
		t.Errorf("[src:%+v][dst:%+v]", src, dst)
	}
}

/*
func TestBaiduStdServerCodec(t *testing.T) {
	t.Run("AddRequest/GetRequest", func(t *testing.T) {
		var req = &BaiduRpcStdProtocol{}
		var codec = &BaiduStdServerCodec{}
		var seq = codec.AddRequest(req)
		var targetReq = codec.GetRequest(seq)
		if targetReq != req {
			t.Errorf("codec.GetRequest(%d) = %v, expect = %v",
				seq, targetReq, req)
		}

		{
			seq++
			var targetReq = codec.GetRequest(seq)
			if targetReq != nil {
				t.Errorf("codec.GetRequest(%d) = %v, expect = %v",
					seq, targetReq, nil)
			}
		}
	})
	t.Run("ReadRequestHeader", func(t *testing.T) {
		var metaSize = 1024
		var bodySize = 2048
		var header = BaiduRpcStdProtocolHeader{
			MetaSize: uint32(metaSize),
			BodySize: uint32(bodySize),
		}
		buf, _ := header.Marshal()
		var rwc = &StubReadWriteCloser{
			buf: bytes.NewBuffer(buf),
		}
		var codec = &BaiduStdServerCodec{
			rwc: rwc,
		}
		var r = Request{}
		t.Errorf("[request:%v][codec:%v]", r, codec)
	})
}
*/

type StubReadWriteCloser struct {
	buf *bytes.Buffer
}

func (s *StubReadWriteCloser) Read(p []byte) (int, error) {
	return s.buf.Read(p)
}

func (s *StubReadWriteCloser) Write(p []byte) (n int, err error) {
	// skip
	return len(p), nil
}

func (s *StubReadWriteCloser) Close() error {
	return nil
}

func rpcRequestMetaEqual(lhs *RpcRequestMeta, rhs *RpcRequestMeta) bool {
	if lhs == rhs || (lhs.GetServiceName() == rhs.GetServiceName() &&
		lhs.GetMethodName() == rhs.GetMethodName() &&
		lhs.GetLogId() == rhs.GetLogId() &&
		lhs.GetTraceId() == rhs.GetTraceId() &&
		lhs.GetSpanId() == rhs.GetSpanId()) {
		return true
	} else {
		return false
	}
}

func rpcResponseMetaEqual(lhs *RpcResponseMeta, rhs *RpcResponseMeta) bool {
	if lhs == rhs ||
		(lhs.GetErrorCode() == rhs.GetErrorCode() &&
			lhs.GetErrorText() == rhs.GetErrorText()) {
		return true
	} else {
		return false
	}
}

func RpcMetaEqual(lhs *RpcMeta, rhs *RpcMeta) bool {
	if !rpcRequestMetaEqual(lhs.GetRequest(), rhs.GetRequest()) ||
		!rpcResponseMetaEqual(lhs.GetResponse(), rhs.GetResponse()) ||
		lhs.GetCompressType() != rhs.GetCompressType() ||
		lhs.GetCorrelationId() != rhs.GetCorrelationId() ||
		lhs.GetAttachmentSize() != rhs.GetAttachmentSize() {
		return false
	} else {
		return true
	}
}
