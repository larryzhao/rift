package rye

import (
	"context"
	"fmt"
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

// TODO: refactor this to runner
type Runner struct {
	xrayConfigFile string
	pacFile        string
	xrayInstance   *core.Instance
	startCh        chan interface{}
	stopCh         chan string
	wg             *sync.WaitGroup
}

func NewRunner(xrayConfigFile string, pacFile string) *Runner {
	return &Runner{
		xrayConfigFile: xrayConfigFile,
		pacFile:        pacFile,
		startCh:        make(chan interface{}, 2),
		stopCh:         make(chan string),
		wg:             &sync.WaitGroup{},
	}
}

func StopRunner(pid int) error {
	err := syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		default:
			killErr := syscall.Kill(pid, syscall.Signal(0))
			if killErr != nil {
				// process does not exist, so shutdown is successful
				return nil
			}

			// otherwise the process is still running, so we go another round after 1 sec.
			time.Sleep(1 * time.Second)
			continue
		case <-ticker.C:
			return fmt.Errorf("stop runner timeout")
		}
	}
}

func (p *Runner) Run() error {
	defer close(p.startCh)

	go p.startXray()
	p.wg.Add(1)
	go p.startPAC()
	p.wg.Add(1)

	var successes int
	ticker := time.NewTicker(20 * time.Second)

startloop:
	for {
		select {
		case sig := <-p.startCh:
			switch sig := sig.(type) {
			case error:
				return fmt.Errorf("start failed: %w", sig)
			default:
				fmt.Println("received signal OK")
				successes++
				if successes == 2 {
					// start successfully
					err := p.enableSystemPACSettings("http://127.0.0.1:60061/pac/proxy.js")
					if err != nil {
						return err
					}
					// TODO: update system PAC settings
					break startloop
				}
			}
		case <-ticker.C:
			return fmt.Errorf("start failed: timeout")
		}
	}

	sigCh := make(chan os.Signal, 1)
	defer close(sigCh)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	p.shutdown()
	p.wg.Wait()
	return nil
}

func (p *Runner) shutdown() {
	p.stopCh <- "xray"
	p.stopCh <- "pac"
}

func (p *Runner) startXray() {
	defer p.wg.Done()

	// TODO: change hardcoded filepath
	f, err := os.Open(p.xrayConfigFile)
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
	fmt.Println("send signal OK - xray")
	p.startCh <- struct{}{}

	for {
		sig := <-p.stopCh
		if sig == "xray" {
			return
		}
	}
}

func (p *Runner) startPAC() {
	defer p.wg.Done()

	server := &http.Server{
		Addr: ":60061",
	}

	http.HandleFunc("/pac/proxy.js", func(w http.ResponseWriter, r *http.Request) {
		// TODO: change hardcoded filepath
		bb, err := os.ReadFile(p.pacFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("read pac file err: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bb)
	})

	go server.ListenAndServe()

	// start ok
	fmt.Println("send signal OK - PAC")
	p.startCh <- struct{}{}

	for {
		sig := <-p.stopCh
		if sig == "pac" {
			server.Shutdown(context.Background())
			return
		}
	}
}

func (p *Runner) enableSystemPACSettings(pacURL string) error {
	command := exec.Command("networksetup", "-setautoproxyurl", "Wi-Fi", pacURL)
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}

func (p *Runner) disableSystemPACSettings() error {
	command := exec.Command("networksetup", "-setautoproxyurl", "Wi-Fi", "off")
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}
