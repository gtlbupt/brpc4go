package brpc

import (
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"time"
)

type ChannelOptions struct {
	addr                  string
	lb                    string
	connection_timeout_ms time.Duration
	timeout_ms            time.Duration
	max_retry             int
	protocol              string
	connection_type       string
}

func (options *ChannelOptions) IsTLS() bool {
	return false
}

func NewChannelOptions() *ChannelOptions {
	return &ChannelOptions{
		connection_timeout_ms: 200,
		timeout_ms:            500,
		max_retry:             1,
		protocol:              "baidu_std",
		connection_type:       "single",
	}
}

type Channel struct {
	options ChannelOptions
	lb      LoadBalancer
}

func NewChannel() *Channel {
	return &Channel{}
}

func (c *Channel) Init(
	addr string, lb string, options *ChannelOptions) error {

	c.options = *options

	if c.options.protocol != ProtocolBaiduStd ||
		c.options.connection_type != ConnTypeSingle {
		return fmt.Errorf("NotSupport Protocol(%s) or ConnType(%s)",
			c.options.protocol, c.options.connection_type)
	}

	var load_balancer, err = NewLoadBalancer(addr, lb, options)
	if err != nil {
		return err
	}
	if err := load_balancer.Init(options); err != nil {
		return err
	} else {
		c.lb = load_balancer
		return nil
	}
}

func (c *Channel) CallMethod(
	md MethodDescriptor,
	cntl Controller,
	request *proto.Message,
	response *proto.Message,
	done chan<- struct{}) {

	if done == nil {
	}
}
