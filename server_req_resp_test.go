package brpc

import (
	"testing"
)

func TestServerRequest(t *testing.T) {

	t.Run("Empty", func(t *testing.T) {
		var rl = RequestList{}
		if rl.Len() != 0 {
			t.Errorf("RequestList.Len() = %d, expect = 0",
				rl.Len())
		}
	})

	t.Run("EmptyGet", func(t *testing.T) {
		var rl = RequestList{}
		var r = rl.Get()
		if r == nil {
			t.Errorf("RequestList.Get() = nil, expect not nil")
		}
		if rl.Len() != 0 {
			t.Errorf("RequestList.Len() = %d, expect = 0",
				rl.Len())
		}
	})

	var newRequest = func() *Request {
		return &Request{}
	}

	t.Run("Put", func(t *testing.T) {
		var rl = RequestList{}
		var r = newRequest()
		rl.Put(r)
		var l = 1
		if rl.Len() != l {
			t.Errorf("RequestList.Len() = %d, expect = %d",
				rl.Len(), l)
		}
		var r1 = newRequest()
		rl.Put(r1)
		l++
		if rl.Len() != l {
			t.Errorf("RequestList.Len() = %d, expect = %d",
				rl.Len(), l)
		}

		var r2 = rl.Get()
		l--
		if r2 != r1 {
			t.Errorf("RequestList.Put() should cache. r2 = %p, r1 = %p",
				r2, r1)
		}
		if rl.Len() != l {
			t.Errorf("RequestList.Len() = %d, expect = %d",
				rl.Len(), l)
		}
	})
}

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
