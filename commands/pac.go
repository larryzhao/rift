package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
			res, err := pac.SyncGFWList(out, repo.Status.GFWList.ETag, repo.Status.GFWList.SHA256)
			if err != nil {
				return fmt.Errorf("update gfwlist err: %w", err)
			}

			now := time.Now()
			repo.Status.UpdateGFWListStatus(now, &now, res.ETag, res.SHA256, nil)
			if err := repo.SaveStatus(); err != nil {
				return fmt.Errorf("save status err: %w", err)
			}

			if res.Changed {
				fmt.Printf("updated %s (%d bytes)\n", out, res.Bytes)
			} else {
				fmt.Printf("%s already up to date\n", out)
			}
			return nil
		},
	}
}
