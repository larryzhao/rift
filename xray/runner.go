package xray

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/larryzhao/rye"
)

type Runner struct {
	Bin    string
	Config string
}

func NewRunner(bin string, config string) *Runner {
	return &Runner{Bin: bin, Config: config}
}

func (runner *Runner) Run() (int, error) {
	return rye.Run(runner.Bin, []string{"-c", runner.Config})
}

func (runner *Runner) ToConfig(server *rye.Server) ([]byte, error) {
	var conf Config
	bb, err := os.ReadFile(runner.Config)
	if err != nil {
		return nil, fmt.Errorf("read xray config %s err: %w", runner.Config, err)
	}

	err = json.Unmarshal(bb, &conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshal xray config %s err: %w", runner.Config, err)
	}

	outbound, err := toOutbound(server)
	if err != nil {
		return nil, fmt.Errorf("build outbound from server %s err: %w", server.ServerName, err)
	}

	err = conf.SetOutbound("proxy", outbound)
	if err != nil {
		return nil, fmt.Errorf("set outbound from server %s err: %w", server.ServerName, err)
	}

	data, err := json.Marshal(conf)
	if err != nil {
		return nil, fmt.Errorf("marshal xray config err: %w", err)
	}
	return data, nil
}
