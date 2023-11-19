package repo

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"

	"github.com/larryzhao/rye/xray"
)

const (
	PIDFile      = "rye.pid"
	SettingsFile = "settings.yaml"
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

type Repo struct {
	Dir        string
	Settings   *Settings
	XrayConfig *xray.Config
	PID        int
}

func settingsFile(repoDir string) string {
	return path.Join(repoDir, "settings.yaml")
}

func xrayConfigFile(repoDir string) string {
	return path.Join(repoDir, "xray", "config.json")
}

func LoadRepo() (*Repo, error) {
	var err error

	repo := &Repo{}

	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("get user's home directory err: %s", err.Error()))
	}
	repo.Dir = path.Join(u.HomeDir, ".rye")

	// load v2ray config
	repo.XrayConfig, err = xray.ParseJSONConfig(xrayConfigFile(repo.Dir))
	if err != nil {
		return nil, fmt.Errorf("decode xray config err: %w", err)
	}

	repo.PID, err = readPID(repo.Dir)
	if err != nil {
		return nil, fmt.Errorf("read pid err: %w", err)
	}

	return repo, nil
}

func (repo *Repo) WritePID(pid int) error {
	return writePID(repo.Dir, pid)
}

func readPID(repoDir string) (int, error) {
	bb, err := os.ReadFile(path.Join(repoDir, PIDFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}

	pid, err := strconv.Atoi(string(bb))
	if err != nil {
		return 0, nil
	}

	return pid, nil
}

func writePID(repoDir string, pid int) error {
	return os.WriteFile(path.Join(repoDir, PIDFile), []byte(strconv.Itoa(pid)), 0644)
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
