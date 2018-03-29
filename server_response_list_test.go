package brpc

import (
	"testing"
)

func TestServerResponse(t *testing.T) {
	var rl = ResponseList{}

	t.Run("Empty", func(t *testing.T) {
		if rl.Len() != 0 {
			t.Errorf("ResponseList.Len() = %d, expect = 0",
				rl.Len())
		}
	})
	t.Run("Get", func(t *testing.T) {
		var r = rl.Get()
		if r == nil {
			t.Errorf("ResponseList.Get() = nil, expect not nil")
		}
	})
	t.Run("Free", func(t *testing.T) {
		var r = rl.Get()
		rl.Put(r)
		var r1 = rl.Get()
		if r1 != r {
			t.Errorf("ResponseList.Put() should cache. r1 = %v, r = %v",
				r1, r)
		}
	})
}
