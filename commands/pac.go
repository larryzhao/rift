package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/pac"
	"github.com/spf13/cobra"
)

func NewPACCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "pac",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			server := pac.NewServer(60061, repo.PACDomainsFile(), repo.PACGFWListFile())
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

	cmd.AddCommand(newPACUpdateCmd())

	return cmd
}

func newPACUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Download the latest gfwlist (refreshed ~every 6h upstream)",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			out := repo.PACGFWListFile()
			n, err := pac.SyncGFWList(out)
			if err != nil {
				return fmt.Errorf("update gfwlist err: %w", err)
			}
			fmt.Printf("updated %s (%d bytes)\n", out, n)
			return nil
		},
	}
}
