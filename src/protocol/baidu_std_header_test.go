package baidu_std

import (
	"testing"
)

func TestBaiduRpcStdProtocolHeader(t *testing.T) {

	t.Run("Constants", func(t *testing.T) {
		var expect = 12
		if BAIDU_STD_RPC_MSG_HEADER_LEN != 12 {
			t.Errorf("BAIDU_STD_RPC_MSG_HEADER_LEN = %d, expect = %d",
				BAIDU_STD_RPC_MSG_HEADER_LEN, expect)
		}
		var s = "PRPC"
		if BAIDU_STD_MAGIC_STRING != s {
			t.Errorf("BAIDU_STD_MAGIC_STRING = %s, expect = %s",
				BAIDU_STD_MAGIC_STRING, s)
		}
	})

	t.Run("GetHeaderLen", func(t *testing.T) {
		var header = BaiduRpcStdProtocolHeader{}
		if header.GetHeaderLen() != BAIDU_STD_RPC_MSG_HEADER_LEN {
			t.Errorf("h.GetHeaderLen() = %d, expect = %d",
				header.GetHeaderLen(), BAIDU_STD_RPC_MSG_HEADER_LEN)
		}
	})

	t.Run("SetAndGetMetaSize", func(t *testing.T) {
		var metaSize = 1024
		var header = BaiduRpcStdProtocolHeader{}
		header.SetMetaSize(metaSize)
		if header.GetMetaSize() != metaSize {
			t.Errorf("h.GetMetaSize() = %d, expect = %d",
				header.GetMetaSize(), metaSize)
		}
	})

	t.Run("SetAndGetBodySize", func(t *testing.T) {
		var bodySize = 2048
		var header = BaiduRpcStdProtocolHeader{}
		header.SetBodySize(bodySize)
		if header.GetBodySize() != bodySize {
			t.Errorf("h.GetBodySize() = %d, expect = %d",
				header.GetBodySize(), bodySize)
		}
	})

	t.Run("MarshalUnmarshal", func(t *testing.T) {
		var metaSize = 1024
		var bodySize = 2048
		var header = BaiduRpcStdProtocolHeader{}
		header.SetMetaSize(metaSize)
		header.SetBodySize(bodySize)

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
