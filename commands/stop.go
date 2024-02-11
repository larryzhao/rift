package commands

import (
	"fmt"
	"os/exec"

	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			// stop runner
			err := rye.StopRunner(repo.Status.PID)
			if err != nil {
				return err
			}

			// unset proxy
			command := exec.Command("networksetup", "-setautoproxystate", "Wi-Fi", "off")
			err = command.Start()
			if err != nil {
				return fmt.Errorf("turn off system proxy err: %w", err)
			}

			rye.PrintlnInfo("stopped")
			return nil
		},
	}
}
