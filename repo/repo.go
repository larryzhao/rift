package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/xray"
	"gopkg.in/yaml.v3"
)

type Log struct {
	Location string `yaml:"location"`
}

type Settings struct {
	Log           *Log                `yaml:"log"`
	Subscriptions []*rye.Subscription `yaml:"subscriptions"`
}

type Server struct {
	Group  string      `yaml:"group"`
	Server *rye.Server `yaml:"server"`
}

type Repo struct {
	Dir        string
	XrayConfig *xray.Config
	Settings   *Settings
	Servers    []*Server
	PID        int
}

func (repo *Repo) settingsFile() string {
	return path.Join(repo.Dir, "settings.yaml")
}

func (repo *Repo) XrayConfigFile() string {
	return path.Join(repo.Dir, "xray", "config.json")
}

func (repo *Repo) PACFile() string {
	return path.Join(repo.Dir, "pac.js")
}

func (repo *Repo) PIDFile() string {
	return path.Join(repo.Dir, "rye.pid")
}

func (repo *Repo) UpdateSubscriptions() error {
	for _, sub := range repo.Settings.Subscriptions {
		ctx := context.Background()
		servers, err := sub.Fetch(ctx)
		if err != nil {
			rye.PrintError("fetch subscription %s err: %s", sub.Name, err.Error())
			continue
		}

		// replace servers
		err = repo.UpdateServersByGroup(sub.Name, servers)
		if err != nil {
			rye.PrintError("update servers from %s err: %s", sub.Name, err.Error())
			continue
		}

		sub.LastUpdatedAt = time.Now()
		rye.PrintInfo("subscription: %s updated", sub.Name)
	}

	// save settings
	bb, err := yaml.Marshal(repo.Settings)
	if err != nil {
		return err
	}

	err = os.WriteFile(repo.settingsFile(), bb, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repo) UpdateServersByGroup(group string, servers []*rye.Server) error {
	var newServers []*Server

	for _, srv := range repo.Servers {
		if srv.Group == group {
			continue
		}
		newServers = append(newServers, srv)
	}

	for _, srv := range servers {
		newServers = append(newServers, &Server{
			Group:  group,
			Server: srv,
		})
	}

	repo.Servers = newServers

	// save servers.yaml
	bb, err := yaml.Marshal(repo.Servers)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(repo.Dir, "servers.yaml"), bb, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadRepo() (*Repo, error) {
	var err error

	repo := &Repo{}

	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("get user's home directory err: %s", err.Error()))
	}
	repo.Dir = path.Join(u.HomeDir, ".rye")

	// load xray config
	repo.XrayConfig, err = xray.ParseJSONConfig(repo.XrayConfigFile())
	if err != nil {
		return nil, fmt.Errorf("decode xray config err: %w", err)
	}

	// read current pid
	repo.PID, err = readPID(repo.PIDFile())
	if err != nil {
		return nil, fmt.Errorf("read pid err: %w", err)
	}

	// load settings
	bb, err := os.ReadFile(repo.settingsFile())
	if err != nil {
		return nil, fmt.Errorf("read settings.yaml err: %w", err)
	}
	repo.Settings = &Settings{}
	err = yaml.Unmarshal(bb, repo.Settings)
	if err != nil {
		return nil, fmt.Errorf("unmarshal settings.yaml err: %w", err)
	}

	// load servers
	bb, err = os.ReadFile(path.Join(repo.Dir, "servers.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read servers.yaml err: %w", err)
	}
	err = yaml.Unmarshal(bb, &repo.Servers)
	if err != nil {
		return nil, fmt.Errorf("unmarshal servers.yaml err: %w", err)
	}

	return repo, nil
}

func (repo *Repo) WritePID(pid int) error {
	return writePID(repo.PIDFile(), pid)
}

func readPID(pidFile string) (int, error) {
	bb, err := os.ReadFile(pidFile)
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

func writePID(pidFile string, pid int) error {
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}
