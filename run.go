package rye

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func Run(executable string, args []string) (int, error) {
	var err error

	// start runner
	command := exec.Command(executable, args...)
	err = command.Start()
	if err != nil {
		PrintlnVerbose("start runner err: %s", err.Error())
		return 0, fmt.Errorf("start runner err: %w", err)
	}

	pid := command.Process.Pid

	// detach
	err = command.Process.Release()
	if err != nil {
		PrintlnVerbose("detach runner err: %s", err.Error())
		return 0, fmt.Errorf("detach runner err: %w", err)
	}

	return pid, nil
}

func Stop(pid int) error {
	PrintlnVerbose("send SIGTERM to runner %d", pid)
	err := syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	PrintlnVerbose("wait 10 seconds for shutting down...")
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		default:
			killErr := syscall.Kill(pid, syscall.Signal(0))
			if killErr != nil {
				// process does not exist, so shutdown is successful
				PrintlnVerbose("runner shutdown ok")
				return nil
			}

			// otherwise the process is still running, so we go another round after 1 sec.
			PrintlnVerbose("runner process %d still exists, wait another sec...", pid)
			time.Sleep(1 * time.Second)
			continue
		case <-ticker.C:
			return fmt.Errorf("stop runner timeout")
		}
	}
}

func Running(pid int) (bool, error) {
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
