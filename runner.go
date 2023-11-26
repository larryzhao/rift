package rye

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

type Runner struct {
	xrayInstance *core.Instance
	startCh      chan interface{}
	logger       *slog.Logger
}

func NewRunner() *Runner {
	return &Runner{
		startCh: make(chan interface{}, 2),
	}
}

func (p *Runner) Run(ctx context.Context) error {
	defer close(p.startCh)

	repo, _ := ctx.Value(CtxKeyRepo).(*Repo)

	// initialize logger
	runnerLog, err := os.OpenFile(repo.RunnerLogFile(), os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer runnerLog.Close()
	p.logger = slog.New(slog.NewJSONHandler(runnerLog, nil))

	// start xray & PAC
	wg := &sync.WaitGroup{}

	ctxXray, cancelXray := context.WithCancel(context.Background())
	go p.startXray(ctxXray, wg, repo.XrayConfigFile())
	wg.Add(1)

	ctxPAC, cancelPAC := context.WithCancel(context.Background())
	go p.startPAC(ctxPAC, wg, repo.PACFile())
	wg.Add(1)

	var successCount int
	ticker := time.NewTicker(20 * time.Second)

startloop:
	for {
		select {
		case sig := <-p.startCh:
			switch sig := sig.(type) {
			case error:
				return fmt.Errorf("start runner err: %w", sig)
			default:
				successCount++
				p.logger.Info(fmt.Sprintf("start runner signal success * %d", successCount))
				if successCount == 2 {
					break startloop
				}
			}
		case <-ticker.C:
			return fmt.Errorf("start failed: timeout")
		}
	}

	p.logger.Info("runner started")

	sigCh := make(chan os.Signal, 1)
	defer close(sigCh)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	p.logger.Info("SIGTERM received")
	cancelXray()
	cancelPAC()
	wg.Wait()

	p.logger.Info("runner shutdown")

	return nil
}

func (p *Runner) startXray(ctx context.Context, wg *sync.WaitGroup, configFile string) {
	defer wg.Done()

	// TODO: change hardcoded filepath
	f, err := os.Open(configFile)
	if err != nil {
		p.startCh <- err
		return
	}

	xrayConfig, err := serial.LoadJSONConfig(f)
	if err != nil {
		p.startCh <- err
		return
	}

	p.xrayInstance, err = core.New(xrayConfig)
	if err != nil {
		p.startCh <- err
		return
	}

	if err := p.xrayInstance.Start(); err != nil {
		p.startCh <- err
		return
	}
	defer p.xrayInstance.Close()

	runtime.GC()
	debug.FreeOSMemory()

	// start ok
	p.startCh <- struct{}{}
	<-ctx.Done()

	p.logger.Info("xray shutdown")
}

func (p *Runner) startPAC(ctx context.Context, wg *sync.WaitGroup, pacFile string) {
	defer wg.Done()

	server := &http.Server{
		Addr: ":60061",
	}

	http.HandleFunc("/pac/proxy.js", func(w http.ResponseWriter, r *http.Request) {
		// TODO: change hardcoded filepath
		bb, err := os.ReadFile(pacFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("read pac file err: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bb)
	})

	go server.ListenAndServe()
	defer server.Shutdown(context.Background())
	p.logger.Info("PAC server started")

	err := setSystemPAC("http://127.0.0.1:60061/pac/proxy.js")
	if err != nil {
		p.startCh <- err
		return
	}
	p.logger.Info("system PAC set")

	// start ok
	p.startCh <- struct{}{}
	<-ctx.Done()

	server.Shutdown(context.Background())
	p.logger.Info("PAC server shutdown")
	unsetSystemPAC()
	p.logger.Info("system PAC unset")
}

func setSystemPAC(pacURL string) error {
	command := exec.Command("networksetup", "-setautoproxyurl", "Wi-Fi", pacURL)
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}

func unsetSystemPAC() error {
	// unset proxy
	command := exec.Command("networksetup", "-setautoproxystate", "Wi-Fi", "off")
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}

func StartRunner() (int, error) {
	var err error

	// start runner
	command := exec.Command("/usr/local/bin/rye", "run")
	err = command.Start()
	if err != nil {
		return 0, fmt.Errorf("start runner err: %w", err)
	}

	pid := command.Process.Pid

	// detach
	err = command.Process.Release()
	if err != nil {
		return 0, fmt.Errorf("detach runner err: %w", err)
	}

	return pid, nil
}

func StopRunner(pid int) error {
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
