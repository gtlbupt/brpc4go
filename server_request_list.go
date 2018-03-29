package brpc

import (
	"sync"
)

type RequestList struct {
	sync.Mutex
	reqs *Request
	len  int
}

func (rl *RequestList) Get() *Request {
	rl.Lock()
	var req = rl.reqs

	if req == nil {
		req = new(Request)
	} else {
		rl.reqs = req.next
		rl.len--
		*req = Request{}
	}
	rl.Unlock()
	return req
}

func (rl *RequestList) Put(req *Request) {
	rl.Lock()
	req.next = rl.reqs
	rl.reqs = req
	rl.len++
	rl.Unlock()
}

func (rl *RequestList) Len() int {
	rl.Lock()
	var l = rl.len
	rl.Unlock()
	return l
}
