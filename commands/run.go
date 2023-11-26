package commands

import (
	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

// Comand Run
//
// `rye run`
//
// start rye runner
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := rye.NewRunner()
			return runner.Run(cmd.Context())
		},
	}

	return cmd
}
