package pac

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/larryzhao/rift"
)

type Runner struct {
}

func NewRunner() *Runner {
	return &Runner{}
}

func (runner *Runner) Run() (int, error) {
	bin, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("locate rift executable err: %w", err)
	}

	return rift.Run(bin, []string{"pac"})
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
