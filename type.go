package rye

import (
	"fmt"
)

type Protocl string

const (
	ProtoclUnknown   Protocl = "unknown"
	ProtoclVMess     Protocl = "vmess"
	ProtoclVLess     Protocl = "vless"
	ProtoclTrojan    Protocl = "trojan"
	ProtoclHysteria2 Protocl = "hysteria2"
)

func (p Protocl) String() string {
	return string(p)
}

func ParseProtocl(s string) (Protocl, error) {
	switch s {
	case "vmess":
		return ProtoclVMess, nil
	case "vless":
		return ProtoclVLess, nil
	case "trojan":
		return ProtoclTrojan, nil
	case "hysteria2":
		return ProtoclHysteria2, nil
	}

	return ProtoclUnknown, fmt.Errorf("unknown protocol: %s", s)
}

type TransportProtocol string

const (
	TransportProtocolUnknown TransportProtocol = "unknown"
	TransportProtocolTCP     TransportProtocol = "tcp"
	TransportProtocolWS      TransportProtocol = "ws"
)

func (p TransportProtocol) String() string {
	return string(p)
}

func ParseTransportProtocol(s string) (TransportProtocol, error) {
	switch s {
	case "tcp":
		return TransportProtocolTCP, nil
	case "ws":
		return TransportProtocolWS, nil
	default:
		return TransportProtocolUnknown, fmt.Errorf("unknown transport protocol: %s", s)
	}
}

type CtxKey int

const (
	CtxKeyRepo CtxKey = iota + 1
)

type Runnable interface {
	Run() (int, error)
}
