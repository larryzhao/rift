package commands

import (
	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/pac"
	"github.com/spf13/cobra"
)

func NewPACCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "pac",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			server := pac.NewServer(60061, repo.PACFile())
			// TODO: run in goroutine and handle sig
			return server.Run()
		},
	}

	return cmd
}
