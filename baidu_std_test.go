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

func TestMarshalUnmarshal(t *testing.T) {
	var body_size uint32 = 1
	var meta_size uint32 = 2
	var src = BaiduStdRpcHeader{
		BodySize: body_size,
		MetaSize: meta_size,
	}

	data, err := src.Marshal()
	if err != nil {
		t.Error(err)
	}

	var dst = BaiduStdRpcHeader{}
	if err := dst.Unmarshal(data); err != nil {
		t.Errorf("[data:%v][err:%v]", data, err)
	}

	if src.BodySize != dst.BodySize ||
		src.MetaSize != dst.MetaSize {
		t.Error(src)
	}
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

func TestRpcResponseMeta(t *testing.T) {
}

func TestRpcMeta(t *testing.T) {
}
