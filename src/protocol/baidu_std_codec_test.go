package baidu_std

/*
import (
	"testing"
)

func TestBaiduStdServerCodec(t *testing.T) {
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
		t.Logf("[data.len:%d]", len(data))

		var rwc = &StubReadWriteCloser{
			buf: bytes.NewBuffer(data),
		}
		var codec = &BaiduStdServerCodec{
		//rwc: rwc,
		}
		var r = &Request{}
		if err := codec.ReadRequestHeader(r); err != nil {
			t.Errorf("(%v).ReadRequestHeader() = %v", codec, err)
		}
		t.Logf("[request:%v][codec:%v]", r, codec)
	})

	t.Run("ReadResponseHeader", func(t *testing.T) {
	})
}

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

*/
