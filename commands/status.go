package commands

import "github.com/spf13/cobra"

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
}
