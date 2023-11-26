package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

// Comand Start
//
// `rye start`
//
// start rye
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := rye.StartRunner()
			if err != nil {
				return err
			}

			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)
			if err := repo.WriteRunnerPID(pid); err != nil {
				return fmt.Errorf("update runner pid err: %w", err)
			}

			rye.PrintlnInfo("started")
			return nil
		},
	}

	return cmd
}
