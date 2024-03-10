package xray

import "github.com/larryzhao/rye"

type Runner struct {
	Bin    string
	Config string
}

func NewRunner(bin string, config string) *Runner {
	return &Runner{Bin: bin, Config: config}
}

func (runner *Runner) Run() (int, error) {
	return rye.Run(runner.Bin, []string{"--config", cli.Config})
}
