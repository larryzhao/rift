package rye

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
	urlParts := strings.Split(serverStr, "://")
	if len(urlParts) != 2 {
		return nil, fmt.Errorf("invalid server string")
	}

	// see if the second part of the url is a base64 encoded one
	if strings.Contains(urlParts[1], ":") {
		return parseServerFromURL(serverStr)
	}

	return parseServerFromBase64EncodedJSON(serverStr)
}

func parseServerFromBase64EncodedJSON(serverStr string) (*Server, error) {
	urlParts := strings.Split(serverStr, "://")
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

func parseServerFromURL(urlString string) (*Server, error) {
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

	if server.Protocol != ProtoclHysteria2 {
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
