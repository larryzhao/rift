package rye

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Protocl string

const (
	ProtoclUnknown   Protocl = "unknown"
	ProtoclVMess     Protocl = "vmess"
	ProtoclVLess     Protocl = "vless"
	ProtoclTrojan    Protocl = "trojan"
	ProtoclHysteria2 Protocl = "hysteria2"
	ProtoclSS        Protocl = "ss"
)

func (p Protocl) String() string {
	return string(p)
}

func (p Protocl) ShortName() string {
	switch p {
	case ProtoclVLess:
		return "VL"
	case ProtoclTrojan:
		return "TR"
	case ProtoclHysteria2:
		return "HY"
	case ProtoclVMess:
		return "VM"
	case ProtoclSS:
		return "SS"
	default:
		panic(fmt.Sprintf("unknown protocol: %s", p.String()))
	}
}

func (p Protocl) Style() lipgloss.Style {
	switch p {
	case ProtoclVLess:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#e95c55")).Bold(true)
	case ProtoclVMess:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6ea55a")).Bold(true)
	case ProtoclTrojan:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#237eb3")).Bold(true)
	case ProtoclHysteria2:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3a687f")).Bold(true)
	case ProtoclSS:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a66bbe")).Bold(true)
	default:
		panic(fmt.Sprintf("unknown protocol: %s", p.String()))
	}
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
	case "ss", "shadowsocks":
		return ProtoclSS, nil
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
	ToConfig(server *Server) ([]byte, error)
}
