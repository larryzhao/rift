package server

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/url"
// 	"strconv"

// 	"github.com/larryzhao/rye"
// 	"github.com/larryzhao/rye/v2ray"
// )

// type VlessServer struct {
// 	protocol   rye.Protocl
// 	name       string
// 	Address    string
// 	Port       int
// 	UserID     string
// 	Flow       string
// 	Encryption string
// 	Network    *NetworkSettigs
// }

// func NewVlessServer(protocol rye.Protocl, name string) (*VlessServer, error) {
// 	return &VlessServer{
// 		protocol: protocol,
// 		name:     name,
// 	}, nil
// }

// // parseVlessServerFromURL
// // vless://31b98cae-da2d-4456-b351-f91838313f0a@108.181.22.55:443?type=tcp&encryption=none&security=reality&flow=xtls-rprx-vision&sni=www.apple.com&pbk=W01B4-7EqJFqY_unAM9SQFnXT7UKjna2yw-y16gVyRE&sid=a7fc4ec1af1cce8f&headerType=none&fp=chrome#%F0%9F%87%BA%F0%9F%87%B8%E7%BE%8E%E5%9B%BD%E9%AB%98%E9%80%9F01
// func (s *VlessServer) parseVlessFromURL(u url.URL) (*VlessServer, error) {
// 	protocol, err := rye.ParseProtocl(u.Scheme)
// 	if err != nil {
// 		return nil, err
// 	}

// 	name := u.Hostname()
// 	if u.Fragment != "" {
// 		name = u.Fragment
// 	}

// 	server, err := NewVlessServer(protocol, name)
// 	if err != nil {
// 		return nil, fmt.Errorf("create trojan server err: %w", err)
// 	}

// 	port, err := strconv.Atoi(u.Port())
// 	if err != nil {
// 		return nil, fmt.Errorf("parse port err: %w", err)
// 	}
// 	server.Port = port
// 	server.UserID = u.User.String()
// 	server.Address = u.Hostname()

// 	transportProtocol, err := rye.ParseTransportProtocol(u.Query().Get("type"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	server.Encryption = u.Query().Get("encryption")
// 	server.Flow = u.Query().Get("flow")

// 	server.Network.TransportProtocol = transportProtocol
// 	server.Network.ServerName = u.Query().Get("sni")
// 	if u.Query().Get("allowInsecure") == "0" {
// 		server.Network.AllowInsecure = false
// 	} else {
// 		server.Network.AllowInsecure = true
// 	}

// 	server.Network.FingerPrint = u.Query().Get("fp")
// 	server.Network.Security = u.Query().Get("security")
// 	server.Network.PublicKey = u.Query().Get("publickey")
// 	server.Network.PublicKey = u.Query().Get("pbk")
// 	server.Network.Path = u.Query().Get("path")
// 	server.Network.ShortID = u.Query().Get("shortid")

// 	return server, nil
// }

// func (server *VlessServer) Hostname() string {
// 	return server.Address
// }

// func (server *VlessServer) Name() string {
// 	return server.name
// }

// func (server *VlessServer) Protocl() rye.Protocl {
// 	return server.protocol
// }

// func (server *VlessServer) ToOutbound() *v2ray.Outbound {
// 	outbound := &v2ray.Outbound{
// 		Protocol:       server.Protocl().String(),
// 		Tag:            "proxy",
// 		Mux:            nil,
// 		StreamSettings: &v2ray.StreamSettings{},
// 	}

// 	// Settings
// 	message, _ := json.Marshal(map[string]interface{}{
// 		"vnext": []map[string]interface{}{
// 			{
// 				"address": server.Address,
// 				"port":    server.Port,
// 				"users": []map[string]interface{}{
// 					{
// 						"id":         server.UserID,
// 						"flow":       server.Flow,
// 						"encryption": server.Encryption,
// 						"level":      0,
// 					},
// 				},
// 			},
// 		},
// 	})
// 	outbound.Settings = (*json.RawMessage)(&message)

// 	// StreamSettings
// 	outbound.StreamSettings.Network = server.Network.TransportProtocol
// 	outbound.StreamSettings.Security = server.Network.Security

// 	switch server.Network.TransportProtocol {
// 	case rye.TransportProtocolWS:
// 		outbound.StreamSettings.WSSettings = &v2ray.WSSettings{
// 			Path: server.Network.Path,
// 			Headers: map[string]string{
// 				"host": server.Hostname(),
// 			},
// 		}
// 	default:
// 		// just do nothing currently
// 	}

// 	switch server.Network.Security {
// 	case "tls":
// 		outbound.StreamSettings.TLSSettings = &v2ray.TLSSettings{
// 			Insecure:   server.Network.AllowInsecure,
// 			ServerName: server.Network.ServerName,
// 		}
// 	case "reality":
// 		outbound.StreamSettings.RealitySettings = &v2ray.RealitySettings{
// 			FingerPrint: server.Network.FingerPrint,
// 			PublicKey:   server.Network.PublicKey,
// 			ServerName:  server.Network.ServerName,
// 			ShortID:     server.Network.ShortID,
// 		}
// 	default:
// 		// just do nothing currently
// 	}

// 	return outbound
// }
