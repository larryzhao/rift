package commands

import (
	"os/exec"

	"github.com/larryzhao/rye"
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

			// stop runner
			err = rye.StopRunner(r.PID)
			if err != nil {
				return err
			}

			// unset proxy
			command := exec.Command("networksetup", "-setautoproxystate", "Wi-Fi", "off")
			err = command.Start()
			if err != nil {
				return err
			}

			return nil
		},
	}
}
