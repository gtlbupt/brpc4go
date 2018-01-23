package brpc

import (
	proto "github.com/golang/protobuf/proto"
	"testing"
)

func TestBaiduStdRpcHeader(t *testing.T) {
	var metaSize = 1024
	var bodySize = 2048
	var header = BaiduStdRpcHeader{
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
		var header = BaiduStdRpcHeader{
			MetaSize: uint32(metaSize),
			BodySize: uint32(bodySize),
		}
		buf, err := header.Marshal()
		if err != nil || len(buf) != header.GetHeaderLen() {
			t.Errorf("(%v).Marshal() = %v",
				header, err)
		}

		var target = &BaiduStdRpcHeader{}
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

func TestBaiduStdServerCodec(t *testing.T) {
	t.Run("AddRequest", func(t *testing.T) {
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
}
