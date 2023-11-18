package rye

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

type Process struct {
	xrayInstance *core.Instance
	startCh      chan interface{}
	stopCh       chan string
}

func NewProcess() *Process {
	return &Process{
		startCh: make(chan interface{}, 2),
		stopCh:  make(chan string),
	}
}

func (p *Process) Start() (int, error) {
	defer close(p.startCh)

	go p.startXray()
	go p.startPAC()

	var successes int
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case sig := <-p.startCh:
			switch sig.(type) {
			case error:
				return 0, fmt.Errorf("start failed: %w", sig.(error))
			default:
				fmt.Println("received signal OK")
				successes++
				if successes == 2 {
					// start successfully
					// TODO: update system PAC settings
					return os.Getpid(), nil
				}
			}
		case <-ticker.C:
			return 0, fmt.Errorf("timeout starting process")
		}
	}
}

func (p *Process) Stop() {
	p.stopCh <- "xray"
	p.stopCh <- "pac"
}

func (p *Process) startXray() {
	// TODO: change hardcoded filepath
	// configArg := cmdarg.Arg{"/Users/larry/.rye/v2ray.json"}
	f, err := os.Open("/Users/larry/.rye/v2ray.json")
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
		Addr: ":6061",
	}

	http.HandleFunc("/pac/proxy.js", func(w http.ResponseWriter, r *http.Request) {
		// TODO: change hardcoded filepath
		bb, err := os.ReadFile("/Users/larry/.v2up/pac.js")
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
