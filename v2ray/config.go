package v2ray

// import (
// 	"fmt"
// 	"io"
// 	"os"

// 	"github.com/larryzhao/rye"
// 	"github.com/spyzhov/ajson"
// )

// type Config struct {
// 	node *ajson.Node
// }

// // LoadConfig loads v2ray JSON configuration from file
// func LoadConfig(path string) (*Config, error) {
// 	f, err := os.Open(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("open config file err: %w", err)
// 	}

// 	data, err := io.ReadAll(f)
// 	if err != nil {
// 		return nil, fmt.Errorf("read config file err: %w", err)
// 	}

// 	n, err := ajson.Unmarshal(data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Config{
// 		node: n,
// 	}, nil
// }

// // NewConfig creates a new v2ray Config from template
// func NewConfig() (*Config, error) {
// 	return nil, nil
// }

// func (conf *Config) UseTrojan(server *rye.TrojanServer) error {
// 	var err error

// 	// 1. find proxy outbound, and add trojan outbound
// 	nodes, err := conf.node.JSONPath("")
// 	if err != nil {
// 		return fmt.Errorf("get outbounds error: %w", err)
// 	}

// 	if len(nodes) > 1 {
// 		return fmt.Errorf("")
// 	}
// 	outboundsNode := nodes[0]

// 	outbounds, err := outboundsNode.GetArray()
// 	if err != nil {
// 		return fmt.Errorf("wrong outbounds node: %w", err)
// 	}

// 	for _, outbound := range outbounds {
// 		tagNode, err := outbound.GetKey("tag")
// 		if err != nil {
// 			return fmt.Errorf("")
// 		}
// 		if tagNode.MustString() != "proxy" {
// 			continue
// 		}

// 		err = outbound.Delete()
// 		if err != nil {
// 			return fmt.Errorf("remove proxy outbound err: %w", err)
// 		}
// 	}

// 	// 2. add proxy outbound
// 	newProxyOutbound := ajson.ObjectNode("", map[string]*ajson.Node{
// 		"mux": ajson.NullNode(""),
// 		"tag": ajson.StringNode("", "proxy"),
// 	})

// 	outboundsNode.AppendArray(newProxyOutbound)

// 	return nil
// }

// // type Log struct {
// // 	Access   string `json:"access"`
// // 	Error    string `json:"error"`
// // 	Loglevel string `json:"loglevel"`
// // }

// // type Inbound struct {
// // 	Listen   string           `json:"listen"`
// // 	Port     string           `json:"port"`
// // 	Protocol string           `json:"protocol"`
// // 	Settings *InboundSettings `json:"settings"`
// // }

// // type InboundSettings struct {
// // 	Auth    string `json:"auth"`
// // 	Timeout any    `json:"timeout"`
// // 	UDP     bool   `json:"udp"`
// // }

// // type Outbound struct {
// // 	Tag            string            `json:"tag"`
// // 	Protocl        string            `json:"protocol"`
// // 	StreamSettings *StreamSettings   `json:"streamSettings"`
// // 	Settings       *OutboundSettings `json:"settings"`
// // }

// // type OutboundSettings struct {
// // 	Vnext   []*Vnext  `json:"vnext,omitempty"`
// // 	Servers []*Server `json:"servers, omitempty"`
// // }

// // type Server struct {
// // 	Address  string `json:"address"`
// // 	Port     int64  `json:"port"`
// // 	Password string `json:"password"`
// // }

// // type Vnext struct {
// // 	Address string `json:"address"`
// // 	Port    int64  `json:"port"`
// // 	Users   []struct {
// // 		AlterID  string `json:"alterId"`
// // 		ID       string `json:"id"`
// // 		Security string `json:"security"`
// // 	} `json:"users"`
// // }

// // type TLSSettings struct {
// // 	ServerName    string `json:"serverName"`
// // 	AllowInsecure bool   `json:"allowInsecure"`
// // }

// // type StreamSettings struct {
// // 	Network     string       `json:"network"`
// // 	Security    string       `json:"security"`
// // 	TLSSettings *TLSSettings `json:"tlsSettings,omitempty"`
// // }

// // type Config struct {
// // 	Log       *Log        `json:"log"`
// // 	Inbounds  []*Inbound  `json:"inbounds"`
// // 	Outbounds []*Outbound `json:"outbound"`
// // }
