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
	"syscall"
	"time"

	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

type Process struct {
	xrayConfigFile string
	pacFile        string
	xrayInstance   *core.Instance
	startCh        chan interface{}
	stopCh         chan string
}

func NewProcess(xrayConfigFile string, pacFile string) *Process {
	return &Process{
		xrayConfigFile: xrayConfigFile,
		pacFile:        pacFile,
		startCh:        make(chan interface{}, 2),
		stopCh:         make(chan string),
	}
}

func (p *Process) Start() error {
	defer close(p.startCh)

	go p.startXray()
	go p.startPAC()

	var successes int
	ticker := time.NewTicker(20 * time.Second)

startloop:
	for {
		select {
		case sig := <-p.startCh:
			switch sig.(type) {
			case error:
				return fmt.Errorf("start failed: %w", sig.(error))
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
	p.Stop()
	return nil
}

func (p *Process) Stop() {
	p.stopCh <- "xray"
	p.stopCh <- "pac"
}

func (p *Process) startXray() {
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

func (p *Process) startPAC() {
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

func (p *Process) enableSystemPACSettings(pacURL string) error {
	command := exec.Command("networksetup", "-setautoproxyurl", "Wi-Fi", pacURL)
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}

func (p *Process) disableSystemPACSettings() error {
	command := exec.Command("networksetup", "-setautoproxyurl", "Wi-Fi", "off")
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}
