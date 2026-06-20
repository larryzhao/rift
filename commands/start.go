package commands

import (
	"fmt"

	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/pac"
	"github.com/larryzhao/rift/singbox"
	"github.com/spf13/cobra"
)

// Comand Start
//
// `rift start`
//
// start rift
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			ok, err := repo.Status.IsProxyRunning()
			if err != nil {
				return fmt.Errorf("check if proxy is running err: %w", err)
			}
			if !ok {
				runner := singbox.NewRunner(repo.SingboxConfigFile(), repo.RunnerLogFile())
				pid, err := runner.Run()
				if err != nil {
					return fmt.Errorf("start sing-box err: %w", err)
				}

				repo.Status.UpdateRunningProcess("proxy", pid)
				err = repo.SaveStatus()
				if err != nil {
					return fmt.Errorf("save status err: %w", err)
				}
			}

			ok, err = repo.Status.IsPACServerRunning()
			if err != nil {
				return err
			}
			if !ok {
				runner := pac.NewRunner()
				pid, err := runner.Run()
				if err != nil {
					rift.PrintlnError("start pac err: %s", err.Error())
					return err
				}
				repo.Status.UpdateRunningProcess("pac", pid)
				err = repo.SaveStatus()
				if err != nil {
					rift.PrintlnError("update status err: %s", err.Error())
					return err
				}
			}

			rift.PrintlnInfo("started")
			return nil
		},
	}

	return cmd
}
