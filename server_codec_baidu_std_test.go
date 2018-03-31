package brpc

import (
	baidu_std "./src/protocol"
	"bytes"
	"testing"
)

func TestNewBaiduStdServerCodec(t *testing.T) {

	t.Run("AddGetRequest", func(t *testing.T) {
		var codec = &BaiduStdServerCodec{}
		{
			var req = &baidu_std.BaiduRpcStdProtocol{}
			var seq = codec.AddRequest(req)
			var expect_seq uint64 = 1
			if seq != expect_seq {
				t.Errorf("Codec.AddRequest(%T) = %d, expect = %d",
					req, seq, expect_seq)
			}
		}
		{
			var req = &baidu_std.BaiduRpcStdProtocol{}
			var seq = codec.AddRequest(req)
			var expect_seq uint64 = 2
			if seq != expect_seq {
				t.Errorf("Codec.AddRequest(%T) = %d, expect = %d",
					req, seq, expect_seq)
			}
		}
	})
	t.Run("GetRequest", func(t *testing.T) {

		t.Run("NotExpist", func(t *testing.T) {
			var codec = &BaiduStdServerCodec{}
			var seq uint64 = 0
			var expect_req = codec.GetRequest(seq)
			if expect_req != nil {
				t.Errorf("codec.GetRequest(%d) = %p, expect = nil",
					seq, expect_req)
			}
		})

		t.Run("Exist", func(t *testing.T) {
			var codec = &BaiduStdServerCodec{}
			var req = &baidu_std.BaiduRpcStdProtocol{}
			var seq = codec.AddRequest(req)
			var expect_req = codec.GetRequest(seq)
			if expect_req != req {
				t.Errorf("codec.GetRequest(%d) = %p, expect = %p",
					seq, req, expect_req)
			}
		})
	})

	t.Run("ReadRequestHeader", func(t *testing.T) {
		var data = baidu_std.MakeTestRequest()
		var buffer = bytes.NewBuffer(data)
		var codec = &BaiduStdServerCodec{
			r: buffer,
		}

		var r = Request{}
		var err = codec.ReadRequestHeader(&r)
		if err != nil {
			t.Errorf("codec.ReadRequestHeader(%T) = %v, expect = nil",
				r, err)
		}
		if r.Seq != 1 {
			t.Errorf("r.Seq = %d, expect = 1",
				r.Seq)
		}
		if r.ServiceMethod != "LogicService.JsonCmd" {
			t.Errorf("r.SM = %v", r.ServiceMethod)
		}
	})
}
