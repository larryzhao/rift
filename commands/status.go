package commands

import (
	"fmt"
	"time"

	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)
			ok, err := repo.Status.IsProxyRunning()
			if ok {
				rye.PrintlnInfo("proxy running, connecting to %s %s", repo.Status.ServerGroup, repo.Status.ServerName)
			} else if err != nil {
				rye.PrintlnError("proxy not running %s", err.Error())
			} else {
				rye.PrintlnError("proxy not running")
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

			autoUpdateRunning, err := repo.Status.IsAutoUpdateRunning()
			if err != nil {
				rye.PrintlnError("autoupdate: no (check err: %s)", err.Error())
			} else if autoUpdateRunning {
				rye.PrintlnInfo("autoupdate: running")
			} else {
				rye.PrintlnError("autoupdate: no")
			}

			for _, sub := range repo.Settings.Subscriptions {
				subStatus, ok := repo.Status.SubscriptionStatuses[sub.Name]
				line := rye.SprintfError("unknown")
				if ok && subStatus.LastSuccessAt > 0 {
					successAt := time.Unix(subStatus.LastSuccessAt, 0).Format("2006-01-02 15:04")
					line = rye.SprintfInfo("updated at: %s", successAt)
				}
				if ok && subStatus.LastError != "" {
					if subStatus.LastAttemptAt > 0 {
						lastErrAt := time.Unix(subStatus.LastAttemptAt, 0).Format("2006-01-02 15:04")
						line = fmt.Sprintf("%s, %s", line, rye.SprintfError("last error at: %s", lastErrAt))
					} else {
						line = fmt.Sprintf("%s, %s", line, rye.SprintfError("last error at: unknown"))
					}
				}
				fmt.Printf("  %s: %s\n", sub.Name, line)
			}
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}
