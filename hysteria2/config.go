package hysteria2

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
