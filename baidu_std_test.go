package brpc

import (
	proto "github.com/golang/protobuf/proto"
	"testing"
)

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
