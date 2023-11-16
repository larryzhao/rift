package repo

import (
	"fmt"
	"os/user"
	"path"
	"time"

	"github.com/larryzhao/rye/v2ray"
)

type Subscription struct {
	Name          string    `yaml:"name"`
	URL           string    `yaml:"url"`
	AddedAt       time.Time `yaml:"added_at"`
	LastUpdatedAt time.Time `yaml:"last_updated_at"`
	SkipUpdate    bool      `yaml:"skip_update"`
}

type Settings struct {
	Log           string         `yaml:"log"`
	Subscriptions []Subscription `yaml:"subscriptions"`
}

// type V2RayConfig struct {
// 	ConfigJSONFile string
// 	config         *v4.Config
// }

// func NewV2RayConfig(configFile string) (*V2RayConfig, error) {
// 	f, err := os.Open(configFile)
// 	if err != nil {
// 		return nil, fmt.Errorf("open file %s err: %w", configFile, err)
// 	}

// 	conf := &V2RayConfig{
// 		ConfigJSONFile: configFile,
// 	}
// 	conf.config, err = serial.DecodeJSONConfig(f)
// 	if err != nil {
// 		return nil, fmt.Errorf("decode v2ray config err: %w", err)
// 	}

// 	return conf, nil
// }

// func (conf *V2RayConfig) SetOutbound(proxy string, outbound v4.OutboundDetourConfig) error {
// 	found := -1
// 	for idx, config := range conf.config.OutboundConfigs {
// 		if config.Tag == "proxy" {
// 			found = idx
// 			break
// 		}
// 	}

// 	// not found, let's append
// 	if found == -1 {
// 		conf.config.OutboundConfigs = append(conf.config.OutboundConfigs, outbound)
// 		return nil
// 	}

// 	conf.config.OutboundConfigs[found] = outbound
// 	return nil
// }

// func (conf *V2RayConfig) Save() error {
// 	bb, err := json.Marshal(conf.config)
// 	if err != nil {
// 		return fmt.Errorf("marshal v2ray config err: %w", err)
// 	}

// 	err = os.WriteFile(conf.ConfigJSONFile, bb, os.FileMode(0644))
// 	if err != nil {
// 		return fmt.Errorf("write v2ray config err: %w", err)
// 	}

// 	return nil
// }

type Repo struct {
	Settings    *Settings
	V2RayConfig *v2ray.Config
}

func RepoDir() string {
	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("get user's home directory err: %s", err.Error()))
	}
	return path.Join(u.HomeDir, ".rye")
}

func SettingsFile() string {
	return path.Join(RepoDir(), "settings.yaml")
}

func V2RayConfigJSONFile() string {
	return path.Join(RepoDir(), "v2ray.json")
}

func LoadRepo() (*Repo, error) {
	var err error

	repo := &Repo{}

	// load v2ray config
	repo.V2RayConfig, err = v2ray.ParseJSONConfig(V2RayConfigJSONFile())
	if err != nil {
		return nil, fmt.Errorf("decode v2ray config err: %w", err)
	}

	return repo, nil
}
