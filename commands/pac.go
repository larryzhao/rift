package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
			go server.Run()
			// TODO: remove hard code
			err := pac.SetSystemPAC("http://127.0.0.1:60061/pac/proxy.js")
			if err != nil {
				return fmt.Errorf("set system PAC err: %w", err)
			}
			defer pac.RemoveSystemPAC()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGQUIT)
			<-sigChan

			return nil
		},
	}

	return cmd
}
