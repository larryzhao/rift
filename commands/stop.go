package commands

import (
	"os/exec"
	"syscall"

	"github.com/larryzhao/rye/repo"
	"github.com/spf13/cobra"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			err = syscall.Kill(r.PID, syscall.SIGTERM)
			if err != nil {
				return err
			}

			command := exec.Command("networksetup", "-setautoproxystate", "Wi-Fi", "off")
			err = command.Start()
			if err != nil {
				return err
			}

			// TODO: check if correctly stopped
			return nil
		},
	}
}
