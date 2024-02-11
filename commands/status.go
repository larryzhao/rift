package commands

import (
	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)
			ok, err := repo.Status.IsRunnerRunning()
			if ok {
				rye.PrintlnInfo("xray running, connecting to %s %s", repo.Status.ServerGroup, repo.Status.ServerName)
			} else if err != nil {
				rye.PrintlnError("xray not running %s", err.Error())
			} else {
				rye.PrintlnError("xray not running")
			}

			ok, err = repo.Status.IsPACServerRunning()
			if ok {
				rye.PrintlnInfo("pac server running")
			} else if err != nil {
				rye.PrintlnError("pac server not running, %s", err.Error())
			} else {

			}

			ok, err = repo.Status.IsProxySet()
			if ok {
				rye.PrintlnInfo("autoproxy set for Wi-Fi")
			} else if err != nil {
				rye.PrintlnError("autoproxy not set, %s", err.Error())
			} else {
				rye.PrintlnError("autoproxy not set")
			}
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}
