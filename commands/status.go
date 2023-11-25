package commands

import (
	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {

			rye.PrintInfo("hello world")
			rye.PrintVerbose("verbose")

			return nil
		},
	}
}
