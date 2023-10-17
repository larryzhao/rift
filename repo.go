package rye

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/v2fly/v2ray-core/v5/infra/conf/serial"
	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
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
	Settings    *Settings
	V2RayConfig *v4.Config
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
	repo := &Repo{}

	f, err := os.Open(V2RayConfigJSONFile())
	if err != nil {
		return nil, fmt.Errorf("open file %s err: %w", V2RayConfigJSONFile(), err)
	}
	repo.V2RayConfig, err = serial.DecodeJSONConfig(f)
	if err != nil {
		return nil, fmt.Errorf("decode v2ray config err: %w", err)
	}

	return repo, nil
}
