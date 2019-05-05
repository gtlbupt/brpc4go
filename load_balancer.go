package brpc

import (
	"fmt"
	"log"
	"sync"
)

type LoadBalancer interface {
	Init(options *ChannelOptions) error
	SelectServer() (*Socket, error)
}

type RandomLoadBalancer struct {
	sync.RWMutex
	ns    NameService
	codec ProtocolCodec

	total_nodes    []ServerNode
	health_nodes   []ServerNode
	unhealth_nodes []ServerNode
	sockets        map[string]*Socket
}

func (lb *RandomLoadBalancer) Init(options *ChannelOptions) error {
	var nodes, err = lb.ns.GetServerNodeList()
	if err != nil || len(nodes) == 0 {
		return fmt.Errorf("InitNameService Failed (%v)", err)
	}

	for _, node := range nodes {

		socket, err := CreateSocket(node, lb.codec, options)
		log.Printf("CreateSocket(%v) = %v", node, err)

		// Put
		lb.sockets[node.ToString()] = socket

		lb.total_nodes = append(lb.total_nodes, node)
		if err != nil {
			lb.unhealth_nodes = append(lb.unhealth_nodes, node)
		} else {
			lb.health_nodes = append(lb.health_nodes, node)
		}
	}

	log.Printf("[total:%d][health:%d][unhealth:%d]\n",
		len(lb.total_nodes),
		len(lb.health_nodes),
		len(lb.unhealth_nodes))

	return nil
}

func (lb *RandomLoadBalancer) SelectServer() (*Socket, error) {
	lb.Lock()
	defer lb.Unlock()
	return lb.selectServer()
}

func (lb *RandomLoadBalancer) selectServer() (*Socket, error) {
	if len(lb.health_nodes) > 0 {
		var node = lb.health_nodes[0]
		var s = lb.sockets[node.ToString()]
		return s, nil
	} else {
		return nil, fmt.Errorf("Not Available Server")
	}
}

func NewRandomLoadBalancer(ns NameService, codec ProtocolCodec) LoadBalancer {
	return &RandomLoadBalancer{
		ns:      ns,
		codec:   codec,
		sockets: make(map[string]*Socket),
	}
}

func NewLoadBalancer(addr string, lb string, options *ChannelOptions) (LoadBalancer, error) {
	var ns, err = NewNameService(addr)
	if err != nil {
		return nil, err
	}

	codec, err := NewProtocolCodec(options.protocol)
	if err != nil {
		return nil, err
	}

	switch lb {
	case LoadBalanceRandom:
		return NewRandomLoadBalancer(ns, codec), nil
	default:
		return nil, fmt.Errorf("lb(%s) does support now", lb)
	}
}
