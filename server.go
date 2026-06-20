package rift

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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
	TransportProtocol TransportProtocol `yaml:"transport_protocol,omitempty"`
	ServerName        string            `yaml:"server_name"`
	Path              string            `yaml:"path,omitempty"`
	AllowInsecure     bool              `yaml:"allow_insecure,omitempty"`
	Security          string            `yaml:"security,omitempty"`
	FingerPrint       string            `yaml:"fingerprint,omitempty"`
	PublicKey         string            `yaml:"public_key,omitempty"`
	ShortID           string            `yaml:"short_id,omitempty"`
	AlterID           string            `yaml:"alterid,omitempty"`
}

func ParseServer(serverStr string) (*Server, error) {
	urlParts := strings.SplitN(serverStr, "://", 2)
	if len(urlParts) != 2 {
		return nil, fmt.Errorf("invalid server string")
	}

	protocol, err := ParseProtocl(urlParts[0])
	if err != nil {
		return nil, fmt.Errorf("parse protocol from %s err: %w", urlParts[0], err)
	}

	switch protocol {
	case ProtoclSS:
		return parseSSServerFromURL(serverStr)
	case ProtoclHysteria2:
		return parseHysteria2ServerFromURL(serverStr)
	case ProtoclVLess:
		return parseVLessServerFromURL(serverStr)
	case ProtoclVMess:
		return parseVMessServerFromURL(serverStr)
	case ProtoclTrojan:
		return parseTrojanServerFromURL(serverStr)
	default:
		return nil, fmt.Errorf("unknown protocol: %s", protocol.String())
	}
}

func parseServerFromBase64EncodedJSON(serverStr string) (*Server, error) {
	urlParts := strings.SplitN(serverStr, "://", 2)
	if len(urlParts) != 2 {
		return nil, fmt.Errorf("invalid server string")
	}

	jsonData, err := base64.StdEncoding.DecodeString(urlParts[1])
	if err != nil {
		return nil, err
	}

	type jsonServerStruct struct {
		V    string `json:"v"`
		PS   string `json:"ps"`
		Add  string `json:"add"`
		Port string `json:"port"`
		ID   string `json:"id"`
		AID  string `json:"aid"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
	}

	var jsonServer jsonServerStruct
	err = json.Unmarshal(jsonData, &jsonServer)
	if err != nil {
		return nil, fmt.Errorf("parse server attrs from json err: %w", err)
	}

	var server Server

	server.Protocol, err = ParseProtocl(urlParts[0])
	if err != nil {
		return nil, fmt.Errorf("parse protocol from %s err: %w", urlParts[0], err)
	}
	server.Host = jsonServer.Add
	server.Port, err = strconv.Atoi(jsonServer.Port)
	if err != nil {
		return nil, fmt.Errorf("parse port to int err: %w", err)
	}
	server.User = jsonServer.ID
	server.Name = jsonServer.PS
	server.Encryption = "auto"
	server.TransportProtocol, err = ParseTransportProtocol(jsonServer.Net)
	if err != nil {
		return nil, fmt.Errorf("parse transport protocol from %s err: %w", jsonServer.Net, err)
	}
	server.Security = jsonServer.Type
	server.AlterID = jsonServer.AID

	return &server, nil
}

func parseVLessServerFromURL(rawURL string) (*Server, error) {
	return parseCommonServerFromURL(rawURL, true)
}

func parseVMessServerFromURL(rawURL string) (*Server, error) {
	urlParts := strings.SplitN(rawURL, "://", 2)
	if len(urlParts) == 2 && !strings.Contains(urlParts[1], ":") {
		return parseServerFromBase64EncodedJSON(rawURL)
	}

	return parseCommonServerFromURL(rawURL, true)
}

func parseTrojanServerFromURL(rawURL string) (*Server, error) {
	return parseCommonServerFromURL(rawURL, true)
}

func parseHysteria2ServerFromURL(rawURL string) (*Server, error) {
	return parseCommonServerFromURL(rawURL, false)
}

func parseCommonServerFromURL(rawURL string, parseTransport bool) (*Server, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url %s err: %w", rawURL, err)
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

	if parseTransport {
		server.TransportProtocol, err = ParseTransportProtocol(u.Query().Get("type"))
		if err != nil {
			return nil, err
		}
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

func parseSSServerFromURL(rawURL string) (*Server, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url %s err: %w", rawURL, err)
	}

	server := &Server{Protocol: ProtoclSS}
	if u.User != nil && u.Hostname() != "" && u.Port() != "" {
		server.Host = u.Hostname()
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("parse port err: %w", err)
		}
		server.Port = port

		userInfo := u.User.String()
		if decodedUserInfo, unescapeErr := url.PathUnescape(userInfo); unescapeErr == nil {
			userInfo = decodedUserInfo
		}
		if decoded, decodeErr := decodeBase64String(userInfo); decodeErr == nil {
			userInfo = decoded
		}
		if err := applyShadowsocksUserInfo(server, userInfo); err != nil {
			return nil, err
		}
	} else {
		encoded := strings.TrimPrefix(rawURL, u.Scheme+"://")
		if idx := strings.IndexAny(encoded, "?#"); idx >= 0 {
			encoded = encoded[:idx]
		}
		decoded, err := decodeBase64String(encoded)
		if err != nil {
			return nil, fmt.Errorf("decode shadowsocks url err: %w", err)
		}
		if err := applyLegacyShadowsocksServer(server, decoded); err != nil {
			return nil, err
		}
	}

	server.Name = server.Host
	if u.Fragment != "" {
		server.Name = u.Fragment
	}

	return server, nil
}

func applyLegacyShadowsocksServer(server *Server, decoded string) error {
	at := strings.LastIndex(decoded, "@")
	if at < 0 {
		return fmt.Errorf("invalid shadowsocks url")
	}

	if err := applyShadowsocksUserInfo(server, decoded[:at]); err != nil {
		return err
	}

	hostPort := decoded[at+1:]
	host, port, ok := strings.Cut(hostPort, ":")
	if !ok {
		return fmt.Errorf("invalid shadowsocks host and port")
	}
	server.Host = host
	var err error
	server.Port, err = strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse port err: %w", err)
	}

	return nil
}

func applyShadowsocksUserInfo(server *Server, userInfo string) error {
	method, password, ok := strings.Cut(userInfo, ":")
	if !ok || method == "" || password == "" {
		return fmt.Errorf("invalid shadowsocks user info")
	}
	server.Encryption = method
	server.User = password
	return nil
}

func decodeBase64String(encoded string) (string, error) {
	encoded = strings.TrimSpace(encoded)
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(encoded)
		if err == nil {
			return string(decoded), nil
		}
	}
	return "", fmt.Errorf("invalid base64 data")
}
