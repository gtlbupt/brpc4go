package baidu_std

import (
	proto "github.com/golang/protobuf/proto"
	"testing"
)

func makeTestBaiduRpcProtocolMeta() *RpcRequestMeta {
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
	return req
}

func makeTestBaiduRpcProtocolData(s string) []byte {
	return []byte(s)
}

func TestBaiduRpcStdProtocolMarshal(t *testing.T) {
}

func TestBaiduRpcStdProtocolUnmarshal(t *testing.T) {
}

func TestRpcMetaRequest(t *testing.T) {
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

func TestRpcMetaResponse(t *testing.T) {
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

	t.Run("Marshal", func(t *testing.T) {
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

		if !rpcMetaEqual(rpcMeta, target) {
			t.Errorf("rpcMetaEqual(%v, %v)", rpcMeta, target)
		}
		// t.Logf("[src:%v][dst:%v]", rpcMeta, target)
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

		if !rpcMetaEqual(rpcMeta, target) {
			t.Errorf("rpcMetaEqual(%v, %v)", rpcMeta, target)
		}
		// t.Logf("[src:%v][dst:%v]", rpcMeta, target)
	})
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

func rpcMetaEqual(lhs *RpcMeta, rhs *RpcMeta) bool {
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
