package xray

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Top level Configurations
// DNS
// TODO: do DNS
type DNS struct{}

// Inbounds
type Inbound struct {
	Listen   string `json:"listen"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Settings struct {
		Auth    string `json:"auth"`
		Timeout int    `json:"timeout"`
	} `json:"settings"`
}

// Log
type Log struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}

// Outbound
type Mux struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency"`
}

type TLSSettings struct {
	Insecure   bool   `json:"allowInsecure"`
	ServerName string `json:"serverName"`
	// Certs                            []*TLSCertConfig      `json:"certificates"`
	// ALPN                             *cfgcommon.StringList `json:"alpn"`
	// EnableSessionResumption          bool                  `json:"enableSessionResumption"`
	// DisableSystemRoot                bool                  `json:"disableSystemRoot"`
	// PinnedPeerCertificateChainSha256 *[]string             `json:"pinnedPeerCertificateChainSha256"`
	// VerifyClientCertificate          bool                  `json:"verifyClientCertificate"`
}

type WSSettings struct {
	Path                 string            `json:"path"`
	Headers              map[string]string `json:"headers,omitempty"`
	AcceptProxyProtocol  bool              `json:"acceptProxyProtocol,omitempty"`
	MaxEarlyData         int32             `json:"maxEarlyData,omitempty"`
	UseBrowserForwarding bool              `json:"useBrowserForwarding,omitempty"`
	EarlyDataHeaderName  string            `json:"earlyDataHeaderName,omitempty"`
}

type RealitySettings struct {
	FingerPrint string `json:"fingerprint,omitempty"`
	PublicKey   string `json:"publicKey,omitempty"`
	ShortID     string `json:"shortId,omitempty"`
	ServerName  string `json:"serverName,omitempty"`
}

type StreamSettings struct {
	Network         string           `json:"network"`
	Security        string           `json:"security"`
	TLSSettings     *TLSSettings     `json:"tlsSettings,omitempty"`
	WSSettings      *WSSettings      `json:"wsSettings,omitempty"`
	RealitySettings *RealitySettings `json:"realitySettings,omitempty"`
}

type Outbound struct {
	Tag            string           `json:"tag"`
	Mux            *Mux             `json:"mux"`
	Protocol       string           `json:"protocol"`
	Settings       *json.RawMessage `json:"settings"`
	StreamSettings *StreamSettings  `json:"streamSettings"`
}

// Routing
type RoutingSettings struct {
	DomainStrategy string `json:"domainStrategy"`
	Rules          []any  `json:"rules"`
}

// Config v2ray configuration
type Config struct {
	Filepath  string      `json:"-"`
	DNS       *DNS        `json:"dns,omitempty"`
	Inbounds  []*Inbound  `json:"inbounds"`
	Log       *Log        `json:"log"`
	Outbounds []*Outbound `json:"outbounds"`
	Routing   struct {
		Settings *RoutingSettings `json:"settings"`
	} `json:"routing"`
	Transport struct{} `json:"transport"`
}

func ParseJSONConfig(configFile string) (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("open v2ray config file err: %w", err)
	}
	defer f.Close()

	bb, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read v2ray config file err: %w", err)
	}

	var config Config
	err = json.Unmarshal(bb, &config)
	if err != nil {
		return nil, fmt.Errorf("decode v2ray config err: %w", err)
	}

	config.Filepath = configFile

	return &config, nil
}

func (conf *Config) SetOutbound(proxy string, outbound *Outbound) error {
	found := -1
	for idx, config := range conf.Outbounds {
		if config.Tag == "proxy" {
			found = idx
			break
		}
	}

	// not found, let's append
	if found == -1 {
		conf.Outbounds = append(conf.Outbounds, outbound)
		return nil
	}

	conf.Outbounds[found] = outbound
	return nil
}

func (conf *Config) Save() error {
	bb, err := json.Marshal(conf)
	if err != nil {
		return fmt.Errorf("encode v2ray config file err: %w", err)
	}

	return os.WriteFile(conf.Filepath, bb, 0644)
}
