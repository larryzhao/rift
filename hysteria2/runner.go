package hysteria2

import (
	"github.com/larryzhao/rye"
)

type Runner struct {
	Bin    string
	Config string
}

func NewRunner(bin string, config string) *Runner {
	return &Runner{
		Bin:    bin,
		Config: config,
	}
}

func (cli *Runner) Run() (int, error) {
	return rye.Run(cli.Bin, []string{"--config", cli.Config})
}
