package brpc

import (
	"sync"
)

type ResponseList struct {
	sync.Mutex
	reqs *Response
	len  int
}

func (rl *ResponseList) Get() *Response {
	rl.Lock()
	var req = rl.reqs

	if req == nil {
		req = new(Response)
	} else {
		rl.reqs = req.next
		rl.len--
		*req = Response{}
	}
	rl.Unlock()
	return req
}

func (rl *ResponseList) Put(req *Response) {
	rl.Lock()
	req.next = rl.reqs
	rl.reqs = req
	rl.len++
	rl.Unlock()
}

func (rl *ResponseList) Len() int {
	rl.Lock()
	var l = rl.len
	rl.Unlock()
	return l
}
