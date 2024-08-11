package commands

import (
	"github.com/spf13/cobra"
)

// Comand Start
//
// `rye restart`
//
// stop rye and then start rye
func NewRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "restart",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			stopCmd := NewStopCmd()
			err = stopCmd.RunE(cmd, args)
			if err != nil {
				return err
			}

			startCmd := NewStartCmd()
			err = startCmd.RunE(cmd, args)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
