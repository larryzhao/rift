package hysteria2

import (
	"fmt"

	"github.com/larryzhao/rye"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    string `yaml:"server"`
	Auth      string `yaml:"auth"`
	Bandwidth struct {
		Up   string `yaml:"up"`
		Down string `yaml:"down"`
	} `yaml:"bandwidth"`
	Socks5 struct {
		Listen string `yaml:"listen"`
	} `yaml:"socks5"`
	HTTP struct {
		Listen string `yaml:"listen"`
	} `yaml:"http"`
	TLS struct {
		Insecure bool `yaml:"insecure"`
	} `yaml:"tls"`
}

func ToConfig(server *rye.Server) ([]byte, error) {
	conf := &Config{
		Server: fmt.Sprintf("%s:%d", server.Host, server.Port),
		Auth:   server.User,
		Bandwidth: struct {
			Up   string `yaml:"up"`
			Down string `yaml:"down"`
		}{
			Up:   "20 mbps",
			Down: "100 mbps",
		},
		Socks5: struct {
			Listen string `yaml:"listen"`
		}{
			Listen: "127.0.0.1:6153",
		},
		HTTP: struct {
			Listen string `yaml:"listen"`
		}{
			Listen: "127.0.0.1:6152",
		},
		TLS: struct {
			Insecure bool `yaml:"insecure"`
		}{
			Insecure: server.AllowInsecure,
		},
	}

	return yaml.Marshal(conf)
}
