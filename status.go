package rye

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
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

type Status struct {
	ServerGroup      string           `json:"server_group"`
	ServerName       string           `json:"server_name"`
	Protocl          Protocl          `json:"protocol"`
	PACPort          int              `json:"pac_port"`
	RunningProcesses []RunningProcess `json:"running_processes"`
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

	// check if proxy is from rye.
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
	var pid int
	for _, p := range status.RunningProcesses {
		if p.Kind != "proxy" {
			continue
		}
		pid = p.PID
		break
	}

	if pid <= 0 {
		return false, nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
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

func (status *Status) Save(filepath string) error {
	bb, err := json.Marshal(status)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, bb, 0644)
	if err != nil {
		return nil
	}

	return nil
}
