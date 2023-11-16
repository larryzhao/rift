package server

import (
	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/v2ray"
)

type Server interface {
	Protocl() rye.Protocl
	Name() string
	Hostname() string
	ToOutbound() *v2ray.Outbound
}

type NetworkSettigs struct {
	TransportProtocol rye.TransportProtocol
	ServerName        string
	Path              string
	AllowInsecure     bool
	Security          string
	FingerPrint       string
	PublicKey         string
	ShortID           string
}
