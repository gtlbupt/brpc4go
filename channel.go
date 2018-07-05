package brpc

import (
	"time"
)

type ChannelOptions struct {
	addr               string
	lb                 string
	connect_timeout_ms time.Duration
	timeout_ms         time.Duration
	max_retry          int
}

type Channel struct {
	options ChannelOptions
}

func NewChannel(options ChannelOptions) *Channel {
	return &Channel{
		options: options,
	}
}
