package rift

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type RunningProcess struct {
	PID       int
	Kind      string
	StartedAt int64 `json:"started_at"`
}

type SubscriptionUpdateStatus struct {
	LastAttemptAt int64  `json:"last_attempt_at,omitempty"`
	LastSuccessAt int64  `json:"last_success_at,omitempty"`
	LastError     string `json:"last_error,omitempty"`
}

// GFWListUpdateStatus tracks the state of the gfwlist auto-update. ETag is the
// value returned by the server, replayed as If-None-Match to skip downloads;
// SHA256 is the local content hash used to detect no-op updates.
type GFWListUpdateStatus struct {
	LastAttemptAt int64  `json:"last_attempt_at,omitempty"`
	LastSuccessAt int64  `json:"last_success_at,omitempty"`
	ETag          string `json:"etag,omitempty"`
	SHA256        string `json:"sha256,omitempty"`
	LastError     string `json:"last_error,omitempty"`
}

type Status struct {
	ServerGroup          string                              `json:"server_group"`
	ServerName           string                              `json:"server_name"`
	Protocl              Protocl                             `json:"protocol"`
	PACPort              int                                 `json:"pac_port"`
	RunningProcesses     []RunningProcess                    `json:"running_processes"`
	SubscriptionStatuses map[string]SubscriptionUpdateStatus `json:"subscription_statuses,omitempty"`
	GFWList              GFWListUpdateStatus                 `json:"gfwlist,omitzero"`
}

func (status *Status) PIDByKind(kind string) int {
	for _, proc := range status.RunningProcesses {
		if proc.Kind != kind {
			continue
		}
		return proc.PID
	}
	return 0
}

func (status *Status) UpdateRunningProcess(kind string, pid int) {
	var idx = -1
	for i, proc := range status.RunningProcesses {
		if proc.Kind == kind {
			idx = i
			break
		}
	}
	if idx > -1 {
		status.RunningProcesses[idx].PID = pid
		status.RunningProcesses[idx].StartedAt = time.Now().Unix()
		return
	}
	status.RunningProcesses = append(status.RunningProcesses, RunningProcess{
		Kind:      kind,
		PID:       pid,
		StartedAt: time.Now().Unix(),
	})
}

func (status *Status) ClearRunningProcess(kind string) {
	var processes []RunningProcess
	for _, proc := range status.RunningProcesses {
		if proc.Kind == kind {
			continue
		}
		processes = append(processes, proc)
	}
	status.RunningProcesses = processes
}

func (status *Status) IsRunningByKind(kind string) (bool, error) {
	pid := status.PIDByKind(kind)
	return Running(pid)
}

func (status *Status) IsAutoUpdateRunning() (bool, error) {
	return status.IsRunningByKind("autoupdate")
}

func (status *Status) UpdateSubscriptionStatus(name string, attemptAt time.Time, successAt *time.Time, err error) {
	if status.SubscriptionStatuses == nil {
		status.SubscriptionStatuses = map[string]SubscriptionUpdateStatus{}
	}

	subStatus := status.SubscriptionStatuses[name]
	subStatus.LastAttemptAt = attemptAt.Unix()
	if successAt != nil {
		subStatus.LastSuccessAt = successAt.Unix()
		subStatus.LastError = ""
	} else if err != nil {
		subStatus.LastError = err.Error()
	}
	status.SubscriptionStatuses[name] = subStatus
}

func (status *Status) UpdateGFWListStatus(attemptAt time.Time, successAt *time.Time, etag, sha256 string, err error) {
	status.GFWList.LastAttemptAt = attemptAt.Unix()
	status.GFWList.ETag = etag
	status.GFWList.SHA256 = sha256
	if successAt != nil {
		status.GFWList.LastSuccessAt = successAt.Unix()
		status.GFWList.LastError = ""
	} else if err != nil {
		status.GFWList.LastError = err.Error()
	}
}

func (status *Status) IsProxySet() (bool, error) {
	// unset proxy
	command := exec.Command("networksetup", "-getautoproxyurl", "Wi-Fi")
	bb, err := command.Output()
	if err != nil {
		return false, err
	}

	// check if Enabled.
	reg := regexp.MustCompile(`Enabled: (.*)`)
	matches := reg.FindStringSubmatch(string(bb))
	if len(matches) > 0 {
		enablement := matches[1]
		if strings.ToLower(strings.Trim(enablement, " ")) == "no" {
			return false, nil
		}
	}

	// check if proxy is from rift.
	reg = regexp.MustCompile(`URL: (.*)`)
	matches = reg.FindStringSubmatch(string(bb))
	if len(matches) > 0 && matches[1] == "http://127.0.0.1:60061/pac/proxy.js" {
		return true, nil
	}
	return false, nil
}

func (status *Status) IsPACServerRunning() (bool, error) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:60061", 3*time.Second)
	if err != nil {
		// if err is ECONNREFUSED means the server is not running
		if errors.Is(err, syscall.ECONNREFUSED) {
			return false, nil
		}
		return false, err
	}
	if conn == nil {
		return false, nil
	}
	conn.Close()
	return true, nil
}

func (status *Status) IsProxyRunning() (bool, error) {
	return status.IsRunningByKind("proxy")
}

func (status *Status) Load(filepath string) error {
	bb, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("read runner file err: %w", err)
	}
	err = json.Unmarshal(bb, status)
	if err != nil {
		return fmt.Errorf("unmarshal runner data err: %w", err)
	}

	return nil
}

func (status *Status) Save(path string) error {
	bb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".status-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(bb); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, path)
}
