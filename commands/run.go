package commands

import (
	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

// Comand Run
//
// `rye run`
//
// rye main process
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			process := rye.NewProcess()
			err := process.Start()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
