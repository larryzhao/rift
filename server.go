package rye

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/larryzhao/rye/xray"
)

type Server struct {
	Protocol   Protocl `yaml:"protocol"`
	Name       string  `yaml:"name"`
	Host       string  `yaml:"host"`
	Port       int     `yaml:"port"`
	User       string  `yaml:"user"`
	Flow       string  `yaml:"flow,omitempty"`
	Encryption string  `yaml:"encryption,omitempty"`

	// Network transport releated
	TransportProtocol TransportProtocol `yaml:"transport_protocol"`
	ServerName        string            `yaml:"server_name"`
	Path              string            `yaml:"path,omitempty"`
	AllowInsecure     bool              `yaml:"allow_insecure,omitempty"`
	Security          string            `yaml:"security,omitempty"`
	FingerPrint       string            `yaml:"fingerprint,omitempty"`
	PublicKey         string            `yaml:"public_key,omitempty"`
	ShortID           string            `yaml:"short_id,omitempty"`
}

func ParseServerFromURL(urlString string) (*Server, error) {
	var err error

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("parse url %s err: %w", urlString, err)
	}

	server := &Server{}
	server.Protocol, err = ParseProtocl(u.Scheme)
	if err != nil {
		return nil, err
	}

	server.Name = u.Hostname()
	if u.Fragment != "" {
		server.Name = u.Fragment
	}

	server.Port, err = strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("parse port err: %w", err)
	}
	server.User = u.User.String()
	server.Host = u.Hostname()

	server.TransportProtocol, err = ParseTransportProtocol(u.Query().Get("type"))
	if err != nil {
		return nil, err
	}
	server.Encryption = u.Query().Get("encryption")
	server.Flow = u.Query().Get("flow")

	server.ServerName = u.Query().Get("sni")
	if u.Query().Get("allowInsecure") == "0" {
		server.AllowInsecure = false
	} else {
		server.AllowInsecure = true
	}
	server.FingerPrint = u.Query().Get("fp")
	server.Security = u.Query().Get("security")
	server.PublicKey = u.Query().Get("publickey")
	server.PublicKey = u.Query().Get("pbk")
	server.Path = u.Query().Get("path")
	server.ShortID = u.Query().Get("sid")

	return server, nil
}

func (server *Server) ToOutbound() (*xray.Outbound, error) {
	switch server.Protocol {
	case ProtoclVLess:
		return server.toVlessOutbound()
	default:
		return nil, fmt.Errorf("unknown protocol %s", server.Protocol.String())
	}
}

func (server *Server) toVlessOutbound() (*xray.Outbound, error) {
	outbound := &xray.Outbound{
		Protocol:       server.Protocol.String(),
		Tag:            "proxy",
		Mux:            nil,
		StreamSettings: &xray.StreamSettings{},
	}

	// Settings
	message, _ := json.Marshal(map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": server.Host,
				"port":    server.Port,
				"users": []map[string]interface{}{
					{
						"id":         server.User,
						"flow":       server.Flow,
						"encryption": server.Encryption,
						"level":      0,
					},
				},
			},
		},
	})
	outbound.Settings = (*json.RawMessage)(&message)

	// StreamSettings
	outbound.StreamSettings.Network = server.TransportProtocol.String()
	outbound.StreamSettings.Security = server.Security

	switch server.TransportProtocol {
	case TransportProtocolWS:
		outbound.StreamSettings.WSSettings = &xray.WSSettings{
			Path: server.Path,
			Headers: map[string]string{
				"host": server.Host,
			},
		}
	default:
		// just do nothing currently
	}

	switch server.Security {
	case "tls":
		outbound.StreamSettings.TLSSettings = &xray.TLSSettings{
			Insecure:   server.AllowInsecure,
			ServerName: server.ServerName,
		}
	case "reality":
		outbound.StreamSettings.RealitySettings = &xray.RealitySettings{
			FingerPrint: server.FingerPrint,
			PublicKey:   server.PublicKey,
			ServerName:  server.ServerName,
			ShortID:     server.ShortID,
		}
	default:
		// just do nothing currently
	}

	return outbound, nil
}
