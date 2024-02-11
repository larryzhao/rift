package rye

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type RunnerStatus struct {
	ServerGroup string `json:"server_group"`
	ServerName  string `json:"server_name"`
	PID         int    `json:"pid"`
	StartedAt   int64  `json:"started_at"`
	PACPort     int    `json:"pac_port"`
}

func (status *RunnerStatus) IsProxySet() (bool, error) {
	// unset proxy
	command := exec.Command("networksetup", "-getautoproxyurl", "Wi-Fi")
	bb, err := command.Output()
	if err != nil {
		return false, err
	}

	if string(bb) == "http://127.0.0.1:60061/pac/proxy.js" {
		return true, nil
	}

	return false, nil
}

func (status *RunnerStatus) IsPACServerRunning() (bool, error) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:60061", 3*time.Second)
	if err != nil {
		return false, err
	}
	if conn == nil {
		return false, nil
	}
	conn.Close()
	return true, nil
}

func (status *RunnerStatus) IsRunnerRunning() (bool, error) {
	if status.PID <= 0 {
		return false, nil
	}

	proc, err := os.FindProcess(int(status.PID))
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

func (status *RunnerStatus) Load(filepath string) error {
	bb, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("read runner file err: %w", err)
	}

	status = &RunnerStatus{}
	err = json.Unmarshal(bb, status)
	if err != nil {
		return fmt.Errorf("unmarshal runner data err: %w", err)
	}

	return nil
}

func (status *RunnerStatus) Save(filepath string) error {
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
