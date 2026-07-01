package rift

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Log struct {
	Location string `yaml:"location"`
}

const (
	defaultSubscriptionsInterval = 30 * time.Minute
	defaultGFWListInterval       = 6 * time.Hour
)

// Duration is a time.Duration that (un)marshals to/from a human string such as
// "30m" or "6h" in YAML (yaml.v3 does not handle time.Duration natively).
type Duration time.Duration

func (d Duration) Duration() time.Duration { return time.Duration(d) }

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	if s == "" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("parse interval %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// AutoUpdate holds the per-feature auto-update configuration. When Skip is true
// the feature is bypassed by the autoupdate daemon.
type AutoUpdate struct {
	Skip     bool     `yaml:"skip"`
	Interval Duration `yaml:"interval"`
}

type SubscriptionsConfig struct {
	AutoUpdate AutoUpdate      `yaml:"autoupdate"`
	Items      []*Subscription `yaml:"items"`
}

type GFWListConfig struct {
	AutoUpdate AutoUpdate `yaml:"autoupdate"`
}

type Settings struct {
	Log           *Log                 `yaml:"log"`
	Subscriptions *SubscriptionsConfig `yaml:"subscriptions"`
	GFWList       *GFWListConfig       `yaml:"gfwlist"`
}

// SubscriptionItems returns the configured subscriptions, or nil if none.
func (s *Settings) SubscriptionItems() []*Subscription {
	if s.Subscriptions == nil {
		return nil
	}
	return s.Subscriptions.Items
}

// SubscriptionsAutoUpdate returns the subscriptions auto-update config with the
// default interval applied when unset.
func (s *Settings) SubscriptionsAutoUpdate() AutoUpdate {
	au := AutoUpdate{}
	if s.Subscriptions != nil {
		au = s.Subscriptions.AutoUpdate
	}
	if au.Interval <= 0 {
		au.Interval = Duration(defaultSubscriptionsInterval)
	}
	return au
}

// GFWListAutoUpdate returns the gfwlist auto-update config with the default
// interval applied when unset.
func (s *Settings) GFWListAutoUpdate() AutoUpdate {
	au := AutoUpdate{}
	if s.GFWList != nil {
		au = s.GFWList.AutoUpdate
	}
	if au.Interval <= 0 {
		au.Interval = Duration(defaultGFWListInterval)
	}
	return au
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

type SubscriptionUpdateStat struct {
	Subscription    *Subscription
	Name            string
	Updated         bool
	ChangedServers  int
	PreviousServers int
	CurrentServers  int
	Err             error
}

func (repo *Repo) settingsFile() string {
	return path.Join(repo.Dir, "settings.yaml")
}

func (repo *Repo) RunnerLogFile() string {
	return path.Join(repo.Dir, "runner.log")
}

func (repo *Repo) SingboxConfigFile() string {
	return path.Join(repo.Dir, "singbox", "config.json")
}

func (repo *Repo) PACFile() string {
	return path.Join(repo.Dir, "pac", "pac.js")
}

func (repo *Repo) PACDomainsFile() string {
	return path.Join(repo.Dir, "pac", "domains.txt")
}

func (repo *Repo) PACGFWListFile() string {
	return path.Join(repo.Dir, "pac", "gfwlist.txt")
}

func (repo *Repo) AutoUpdateLogFile() string {
	return path.Join(repo.Dir, "autoupdate.log")
}

// AppendAutoUpdateLog appends a single timestamped line to the autoupdate log.
func (repo *Repo) AppendAutoUpdateLog(format string, args ...interface{}) error {
	logFile := repo.AutoUpdateLogFile()
	if err := os.MkdirAll(path.Dir(logFile), 0755); err != nil {
		return fmt.Errorf("create autoupdate log directory err: %w", err)
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open autoupdate log file err: %w", err)
	}
	defer file.Close()

	line := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("write autoupdate log file err: %w", err)
	}
	return nil
}

// func (repo *Repo) RunnerPIDFile() string {
// 	return path.Join(repo.Dir, "runner.pid")
// }

func (repo *Repo) StatusFile() string {
	return path.Join(repo.Dir, "status.json")
}

func (repo *Repo) AddSubscription(subscription *Subscription) error {
	if repo.Settings.Subscriptions == nil {
		repo.Settings.Subscriptions = &SubscriptionsConfig{}
	}
	for _, sub := range repo.Settings.Subscriptions.Items {
		if sub.Name == subscription.Name {
			return fmt.Errorf("subscription should have a distinct name")
		}
	}

	repo.Settings.Subscriptions.Items = append(repo.Settings.Subscriptions.Items, subscription)

	bb, err := yaml.Marshal(repo.Settings)
	if err != nil {
		return fmt.Errorf("marshal repo settings err: %w", err)
	}
	err = os.WriteFile(repo.settingsFile(), bb, 0644)
	if err != nil {
		return fmt.Errorf("write repo settings file err: %w", err)
	}

	return nil
}

func (repo *Repo) UpdateSubscriptions() ([]*Subscription, error) {
	stats, err := repo.UpdateSubscriptionsWithStats()
	if err != nil {
		return nil, err
	}

	var updatedSubs []*Subscription
	for _, stat := range stats {
		if stat.Updated {
			updatedSubs = append(updatedSubs, stat.Subscription)
		}
	}
	return updatedSubs, nil
}

func (repo *Repo) UpdateSubscriptionsWithStats() ([]SubscriptionUpdateStat, error) {
	var stats []SubscriptionUpdateStat
	for _, sub := range repo.Settings.SubscriptionItems() {
		oldServers := repo.ServersByGroup(sub.Name)
		ctx := context.Background()
		servers, err := sub.Fetch(ctx)
		if err != nil {
			PrintlnError("fetch subscription %s err: %s", sub.Name, err.Error())
			stats = append(stats, SubscriptionUpdateStat{
				Subscription:    sub,
				Name:            sub.Name,
				Updated:         false,
				ChangedServers:  0,
				PreviousServers: len(oldServers),
				CurrentServers:  len(oldServers),
				Err:             err,
			})
			continue
		}

		changedServers := diffServerCount(oldServers, servers)

		// replace servers
		err = repo.UpdateServersByGroup(sub.Name, servers)
		if err != nil {
			PrintlnError("update servers from %s err: %s", sub.Name, err.Error())
			stats = append(stats, SubscriptionUpdateStat{
				Subscription:    sub,
				Name:            sub.Name,
				Updated:         false,
				ChangedServers:  0,
				PreviousServers: len(oldServers),
				CurrentServers:  len(oldServers),
				Err:             err,
			})
			continue
		}

		sub.LastUpdatedAt = time.Now()

		stats = append(stats, SubscriptionUpdateStat{
			Subscription:    sub,
			Name:            sub.Name,
			Updated:         true,
			ChangedServers:  changedServers,
			PreviousServers: len(oldServers),
			CurrentServers:  len(servers),
		})
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

	return stats, nil
}

func (repo *Repo) UpdateServersByGroup(group string, servers []*Server) error {
	var newServers []*RepoServer
	var oldServers []*RepoServer

	// preserve the servers that NOT belongs to the group to be updated
	// they should be kept unchanged after the update
	for _, srv := range repo.Servers {
		if srv.Group == group {
			// keep old servers just for verbose output
			if PrintVerbosly {
				oldServers = append(oldServers, srv)
			}
			continue
		}
		newServers = append(newServers, srv)
	}

	// add the new servers to the group to be updated
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
		repo.Dir = path.Join(u.HomeDir, ".rift")
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

func (repo *Repo) ServersByGroup(group string) []*Server {
	var servers []*Server
	for _, repoServer := range repo.Servers {
		if repoServer.Group != group {
			continue
		}
		servers = append(servers, repoServer.Server)
	}
	return servers
}

func diffServerCount(oldServers []*Server, newServers []*Server) int {
	oldMap := map[string]int{}
	newMap := map[string]int{}

	for _, server := range oldServers {
		key := serverIdentity(server)
		oldMap[key]++
	}
	for _, server := range newServers {
		key := serverIdentity(server)
		newMap[key]++
	}

	changed := 0
	for key, oldCount := range oldMap {
		newCount := newMap[key]
		if oldCount > newCount {
			changed += oldCount - newCount
		}
	}
	for key, newCount := range newMap {
		oldCount := oldMap[key]
		if newCount > oldCount {
			changed += newCount - oldCount
		}
	}

	return changed
}

func serverIdentity(server *Server) string {
	if server == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(server.Protocol.String())
	b.WriteString("|")
	b.WriteString(server.Name)
	b.WriteString("|")
	b.WriteString(server.Host)
	b.WriteString("|")
	b.WriteString(strconv.Itoa(server.Port))
	b.WriteString("|")
	b.WriteString(server.User)
	b.WriteString("|")
	b.WriteString(server.Flow)
	b.WriteString("|")
	b.WriteString(server.Encryption)
	b.WriteString("|")
	b.WriteString(server.TransportProtocol.String())
	b.WriteString("|")
	b.WriteString(server.ServerName)
	b.WriteString("|")
	b.WriteString(server.Path)
	b.WriteString("|")
	if server.AllowInsecure {
		b.WriteString("1")
	} else {
		b.WriteString("0")
	}
	b.WriteString("|")
	b.WriteString(server.Security)
	b.WriteString("|")
	b.WriteString(server.FingerPrint)
	b.WriteString("|")
	b.WriteString(server.PublicKey)
	b.WriteString("|")
	b.WriteString(server.ShortID)
	b.WriteString("|")
	b.WriteString(server.AlterID)
	return b.String()
}
