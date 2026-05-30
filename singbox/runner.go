package singbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/larryzhao/rye"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

// Runner spawns the current rye binary as a detached `run proxy` child so the
// sing-box library runs in its own process and can still be tracked via PID.
type Runner struct {
	Config  string
	LogFile string
}

func NewRunner(config string, logFile string) *Runner {
	return &Runner{Config: config, LogFile: logFile}
}

func (r *Runner) Run() (int, error) {
	bin, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("locate rye executable err: %w", err)
	}

	cmd := exec.Command(bin, "run", "proxy")

	if r.LogFile != "" {
		f, err := os.OpenFile(r.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			cmd.Stdout = f
			cmd.Stderr = f
		}
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start sing-box runner err: %w", err)
	}

	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return 0, fmt.Errorf("detach sing-box runner err: %w", err)
	}
	return pid, nil
}

func (r *Runner) ToConfig(server *rye.Server) ([]byte, error) {
	return BuildConfig(server)
}

// RunForeground loads the sing-box config and runs the Box instance in the
// current process until SIGTERM/SIGINT is received.
func RunForeground(configPath string) error {
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read sing-box config err: %w", err)
	}

	ctx := include.Context(context.Background())

	options, err := json.UnmarshalExtendedContext[option.Options](ctx, configContent)
	if err != nil {
		return fmt.Errorf("parse sing-box config err: %w", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	instance, err := box.New(box.Options{
		Context: runCtx,
		Options: options,
	})
	if err != nil {
		return fmt.Errorf("create sing-box instance err: %w", err)
	}

	if err := instance.Start(); err != nil {
		instance.Close()
		return fmt.Errorf("start sing-box instance err: %w", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-signals

	return instance.Close()
}
