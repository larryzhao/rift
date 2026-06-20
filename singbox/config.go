package singbox

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/larryzhao/rift"
)

const (
	MixedListen    = "127.0.0.1"
	MixedPort      = 6152
	SocksPort      = 6153
	ProxyOutbound  = "proxy"
	DirectOutbound = "direct"
)

func BuildConfig(server *rift.Server) ([]byte, error) {
	outbound, err := buildOutbound(server)
	if err != nil {
		return nil, err
	}

	conf := map[string]any{
		"log": map[string]any{
			"level":     "info",
			"timestamp": true,
		},
		"inbounds": []any{
			map[string]any{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      MixedListen,
				"listen_port": MixedPort,
			},
			map[string]any{
				"type":        "socks",
				"tag":         "socks-in",
				"listen":      MixedListen,
				"listen_port": SocksPort,
			},
		},
		"outbounds": []any{
			outbound,
			map[string]any{"type": "direct", "tag": DirectOutbound},
		},
		"route": map[string]any{
			"final": ProxyOutbound,
		},
	}

	return json.MarshalIndent(conf, "", "  ")
}

func buildOutbound(server *rift.Server) (map[string]any, error) {
	switch server.Protocol {
	case rift.ProtoclVLess:
		return buildVlessOutbound(server), nil
	case rift.ProtoclVMess:
		return buildVMessOutbound(server)
	case rift.ProtoclSS:
		return buildShadowsocksOutbound(server), nil
	case rift.ProtoclHysteria2:
		return buildHysteria2Outbound(server), nil
	case rift.ProtoclTrojan:
		return buildTrojanOutbound(server), nil
	default:
		return nil, fmt.Errorf("unknown protocol %s", server.Protocol.String())
	}
}

func baseOutbound(typ string, server *rift.Server) map[string]any {
	return map[string]any{
		"type":        typ,
		"tag":         ProxyOutbound,
		"server":      server.Host,
		"server_port": server.Port,
	}
}

func buildVlessOutbound(server *rift.Server) map[string]any {
	out := baseOutbound("vless", server)
	out["uuid"] = server.User
	if server.Flow != "" {
		out["flow"] = server.Flow
	}
	if tls := buildTLS(server); tls != nil {
		out["tls"] = tls
	}
	if tr := buildTransport(server); tr != nil {
		out["transport"] = tr
	}
	return out
}

func buildVMessOutbound(server *rift.Server) (map[string]any, error) {
	out := baseOutbound("vmess", server)
	out["uuid"] = server.User

	alterID := 0
	if server.AlterID != "" {
		v, err := strconv.Atoi(server.AlterID)
		if err != nil {
			return nil, fmt.Errorf("parse vmess alter_id %s err: %w", server.AlterID, err)
		}
		alterID = v
	}
	out["alter_id"] = alterID

	security := server.Encryption
	if security == "" {
		security = "auto"
	}
	out["security"] = security

	if tls := buildTLS(server); tls != nil {
		out["tls"] = tls
	}
	if tr := buildTransport(server); tr != nil {
		out["transport"] = tr
	}
	return out, nil
}

func buildShadowsocksOutbound(server *rift.Server) map[string]any {
	out := baseOutbound("shadowsocks", server)
	out["method"] = server.Encryption
	out["password"] = server.User
	return out
}

func buildHysteria2Outbound(server *rift.Server) map[string]any {
	out := baseOutbound("hysteria2", server)
	out["password"] = server.User
	tls := map[string]any{
		"enabled":  true,
		"insecure": server.AllowInsecure,
	}
	if server.ServerName != "" {
		tls["server_name"] = server.ServerName
	}
	out["tls"] = tls
	return out
}

func buildTrojanOutbound(server *rift.Server) map[string]any {
	out := baseOutbound("trojan", server)
	out["password"] = server.User
	if tls := buildTLS(server); tls != nil {
		out["tls"] = tls
	}
	if tr := buildTransport(server); tr != nil {
		out["transport"] = tr
	}
	return out
}

func buildTLS(server *rift.Server) map[string]any {
	switch server.Security {
	case "tls":
		tls := map[string]any{
			"enabled":  true,
			"insecure": server.AllowInsecure,
		}
		if server.ServerName != "" {
			tls["server_name"] = server.ServerName
		}
		if server.FingerPrint != "" {
			tls["utls"] = map[string]any{
				"enabled":     true,
				"fingerprint": server.FingerPrint,
			}
		}
		return tls
	case "reality":
		tls := map[string]any{
			"enabled":  true,
			"insecure": server.AllowInsecure,
		}
		if server.ServerName != "" {
			tls["server_name"] = server.ServerName
		}
		if server.FingerPrint != "" {
			tls["utls"] = map[string]any{
				"enabled":     true,
				"fingerprint": server.FingerPrint,
			}
		}
		reality := map[string]any{
			"enabled": true,
		}
		if server.PublicKey != "" {
			reality["public_key"] = server.PublicKey
		}
		if server.ShortID != "" {
			reality["short_id"] = server.ShortID
		}
		tls["reality"] = reality
		return tls
	default:
		return nil
	}
}

func buildTransport(server *rift.Server) map[string]any {
	switch server.TransportProtocol {
	case rift.TransportProtocolWS:
		tr := map[string]any{
			"type": "ws",
		}
		if server.Path != "" {
			tr["path"] = server.Path
		}
		if server.Host != "" {
			tr["headers"] = map[string]any{
				"Host": server.Host,
			}
		}
		return tr
	default:
		return nil
	}
}
