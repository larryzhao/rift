package server

import (
	"fmt"
	"net/url"
	"strconv"

	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
)

type TrojanServer struct {
	protocol Protocl
	hostname string
	name     string
	Port     int
	Password string
	Network  *NetworkSettigs
}

func NewTrojanServer(protocol Protocl, hostname string, name string) (*TrojanServer, error) {
	return &TrojanServer{
		protocol: protocol,
		hostname: hostname,
		name:     name,
		Network:  &NetworkSettigs{},
	}, nil
}

func parseTrojanFromURL(u *url.URL) (*TrojanServer, error) {
	protocol, err := ParseProtocl(u.Scheme)
	if err != nil {
		return nil, err
	}

	name := u.Hostname()
	if u.Fragment != "" {
		name = u.Fragment
	}

	server, err := NewTrojanServer(protocol, u.Hostname(), name)
	if err != nil {
		return nil, fmt.Errorf("create trojan server err: %w", err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("parse port err: %w", err)
	}

	server.Port = port
	server.Password = u.User.String()

	networkType, err := ParseNetworkType(u.Query().Get("type"))
	if err != nil {
		return nil, err
	}

	server.Network.Type = networkType
	server.Network.ServerName = u.Query().Get("sni")
	if u.Query().Get("allowInsecure") == "0" {
		server.Network.AllowInsecure = false
	} else {
		server.Network.AllowInsecure = true
	}

	server.Network.Path = u.Query().Get("path")
	server.Network.Security = "tls"

	return server, nil
}

func (server *TrojanServer) Hostname() string {
	return server.hostname
}

func (server *TrojanServer) Name() string {
	return server.name
}

func (server *TrojanServer) Protocl() Protocl {
	return server.protocol
}

func (server *TrojanServer) OutboundConfig() *v4.OutboundDetourConfig {
	return nil
}
