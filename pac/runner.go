package pac

import (
	"os/exec"

	"github.com/larryzhao/rift"
)

type Runner struct {
}

func NewRunner() *Runner {
	return &Runner{}
}

func (runner *Runner) Run() (int, error) {
	return rift.Run("/usr/local/bin/rift", []string{"pac"})
}

func SetSystemPAC(pacURL string) error {
	command := exec.Command("networksetup", "-setautoproxyurl", "Wi-Fi", pacURL)
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}

func RemoveSystemPAC() error {
	// unset proxy
	command := exec.Command("networksetup", "-setautoproxystate", "Wi-Fi", "off")
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}
