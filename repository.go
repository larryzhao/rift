package rye

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"gopkg.in/yaml.v3"
)

type Log struct {
	Location string `yaml:"location"`
}

type Settings struct {
	Log           *Log            `yaml:"log"`
	Subscriptions []*Subscription `yaml:"subscriptions"`
}

type RepoServer struct {
	Group  string  `yaml:"group"`
	Server *Server `yaml:"server"`
}

type Repo struct {
	Dir string
	// XrayConfig *xray.Config
	Settings *Settings
	Servers  []*RepoServer
	Status   *Status
}

func (repo *Repo) settingsFile() string {
	return path.Join(repo.Dir, "settings.yaml")
}

func (repo *Repo) RunnerLogFile() string {
	return path.Join(repo.Dir, "runner.log")
}

func (repo *Repo) HysteriaConfigFile() string {
	return path.Join(repo.Dir, "hysteria2", "config.yaml")
}

func (repo *Repo) XrayConfigFile() string {
	return path.Join(repo.Dir, "xray", "config.json")
}

func (repo *Repo) PACFile() string {
	return path.Join(repo.Dir, "pac", "pac.js")
}

// func (repo *Repo) RunnerPIDFile() string {
// 	return path.Join(repo.Dir, "runner.pid")
// }

func (repo *Repo) StatusFile() string {
	return path.Join(repo.Dir, "status.json")
}

func (repo *Repo) UpdateSubscriptions() ([]*Subscription, error) {
	var updatedSubs []*Subscription
	for _, sub := range repo.Settings.Subscriptions {
		ctx := context.Background()
		servers, err := sub.Fetch(ctx)
		if err != nil {
			PrintlnError("fetch subscription %s err: %s", sub.Name, err.Error())
			continue
		}

		// replace servers
		err = repo.UpdateServersByGroup(sub.Name, servers)
		if err != nil {
			PrintlnError("update servers from %s err: %s", sub.Name, err.Error())
			continue
		}

		sub.LastUpdatedAt = time.Now()

		updatedSubs = append(updatedSubs, sub)
		PrintlnVerbose("subscription: %s updated", sub.Name)
	}

	// save settings
	bb, err := yaml.Marshal(repo.Settings)
	if err != nil {
		return nil, fmt.Errorf("marshal repo settings err: %w", err)
	}

	err = os.WriteFile(repo.settingsFile(), bb, 0644)
	if err != nil {
		return nil, fmt.Errorf("write repo settings file err: %w", err)
	}

	return updatedSubs, nil
}

func (repo *Repo) UpdateServersByGroup(group string, servers []*Server) error {
	var newServers []*RepoServer

	for _, srv := range repo.Servers {
		if srv.Group == group {
			continue
		}
		newServers = append(newServers, srv)
	}

	for _, srv := range servers {
		newServers = append(newServers, &RepoServer{
			Group:  group,
			Server: srv,
		})
	}

	repo.Servers = newServers

	// save servers.yaml
	bb, err := yaml.Marshal(repo.Servers)
	if err != nil {
		return fmt.Errorf("marshal repo servers err: %w", err)
	}

	err = os.WriteFile(path.Join(repo.Dir, "servers.yaml"), bb, 0644)
	if err != nil {
		return fmt.Errorf("write repo servers file err: %w", err)
	}

	return nil
}

func LoadRepo() (*Repo, error) {
	var err error

	repo := &Repo{}

	repo.Dir = os.Getenv("REPO_DIR")

	if repo.Dir == "" {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("get user's home directory err: %s", err.Error())
		}
		repo.Dir = path.Join(u.HomeDir, ".rye")
	}

	// load xray config
	// repo.XrayConfig, err = xray.ParseJSONConfig(repo.XrayConfigFile())
	// if err != nil {
	// 	return nil, fmt.Errorf("decode xray config err: %w", err)
	// }

	// load runner status
	repo.Status = &Status{}
	err = repo.Status.Load(repo.StatusFile())
	if err != nil {
		return nil, err
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

func (repo *Repo) SaveStatus() error {
	return repo.Status.Save(repo.StatusFile())
}
