package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*repo.Repo)
			return fmt.Errorf("just a test error from status")
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}
