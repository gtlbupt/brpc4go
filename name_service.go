package brpc

import (
	"fmt"
	"strconv"
	"strings"
)

type NameService interface {
	GetServerNodeList() ([]ServerNode, error)
}

type ListNameService struct {
	addr string
}

func NewListNameService(addr string) NameService {
	return &ListNameService{
		addr: addr,
	}
}

func (ns *ListNameService) GetServerNodeList() ([]ServerNode, error) {
	var r []ServerNode
	var l = strings.TrimLeft(ns.addr, "list://")
	var ip_ports = strings.Split(l, ",")
	for _, v := range ip_ports {
		ip_port := strings.Split(v, ":")
		if len(ip_port) != 2 {
			return nil, fmt.Errorf("Bad Format(%s)", v)
		}
		var ip = ip_port[0]
		var port, err = strconv.Atoi(ip_port[1])
		if err != nil {
			return nil, err
		}
		var node = ServerNode{Ip: ip, Port: port}
		r = append(r, node)
	}
	return r, nil
}

func NewNameService(addr string) (NameService, error) {
	if strings.HasPrefix(addr, NameServiceList) {
		return NewListNameService(addr), nil
	} else {
		return nil, fmt.Errorf("(%s) does have NameService", addr)
	}
}
