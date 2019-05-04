package brpc

import (
	"testing"
)

func TestChannelOptions(t *testing.T) {
	t.Run("NewChannelOptions", func(t *testing.T) {
		var opt = NewChannelOptions()
		if opt.connection_timeout_ms != 200 {
			t.Errorf("opt.connection_timeout_ms = %d, expect = %dms",
				opt.connection_timeout_ms, 200)
		}

		if opt.timeout_ms != 500 {
			t.Errorf("opt.timeout_ms = %d, expect = %d",
				opt.timeout_ms, 500)
		}

		if opt.max_retry != 1 {
			t.Errorf("opt.max_retry = %d, expect = %d",
				opt.max_retry, 1)
		}

		if opt.protocol != "baidu_std" {
			t.Errorf("opt.protocol = %s, expect = %s",
				opt.protocol, "baidu_std")
		}
	})
}

func TestChannelInit(t *testing.T) {
	var addr = "list://127.0.0.1:8080"
	var lb = "random"
	var options = NewChannelOptions()
	t.Run("Default", func(t *testing.T) {
		var channel = NewChannel()
		err := channel.Init(addr, lb, options)
		if err != nil {
			t.Errorf("channel.Init(%s, %s, %v) = %v",
				addr, lb, options, err)
		}
	})
}
