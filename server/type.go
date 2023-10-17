package server

import "fmt"

type Protocl string

const (
	ProtoclUnknown Protocl = "unknown"
	ProtoclVMess   Protocl = "vmess"
	ProtoclVLess   Protocl = "vless"
	ProtoclTrojan  Protocl = "trojan"
)

func ParseProtocl(s string) (Protocl, error) {
	switch s {
	case "vmess":
		return ProtoclVMess, nil
	case "vless":
		return ProtoclVLess, nil
	case "trojan":
		return ProtoclTrojan, nil
	}

	return ProtoclUnknown, fmt.Errorf("unknown protocol: %s", s)
}

type NetworkType string

const (
	NetworkTypeUnknown NetworkType = "unknown"
	NetworkTypeTCP     NetworkType = "tcp"
	NetworkTypeWS      NetworkType = "ws"
)

func ParseNetworkType(s string) (NetworkType, error) {
	switch s {
	case "tcp":
		return NetworkTypeTCP, nil
	case "ws":
		return NetworkTypeWS, nil
	default:
		return NetworkTypeUnknown, fmt.Errorf("unknown network type: %s", s)
	}
}

type Server interface {
	Protocl() Protocl
	Name() string
	Hostname() string
}

type NetworkSettigs struct {
	Type          NetworkType
	ServerName    string
	Path          string
	AllowInsecure bool
	Security      string
}
