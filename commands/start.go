package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/hysteria2"
	"github.com/larryzhao/rye/pac"
	"github.com/larryzhao/rye/xray"
	"github.com/spf13/cobra"
)

// Comand Start
//
// `rye start`
//
// start rye
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			ok, err := repo.Status.IsProxyRunning()
			if err != nil {
				return fmt.Errorf("check if proxy is running err: %w", err)
			}
			if !ok {
				switch repo.Status.Protocl {
				case rye.ProtoclHysteria2:
					runner := hysteria2.NewRunner("/opt/homebrew/bin/hysteria", repo.HysteriaConfigFile())
					pid, err := runner.Run()
					if err != nil {
						return fmt.Errorf("start hysteria2 err: %w", err)
					}

					repo.Status.UpdateRunningProcess("proxy", pid)
					err = repo.SaveStatus()
					if err != nil {
						return fmt.Errorf("save status err: %w", err)
					}
				case rye.ProtoclVLess, rye.ProtoclVMess:
					runner := xray.NewRunner("/opt/homebrew/bin/xray", repo.XrayConfigFile())
					pid, err := runner.Run()
					if err != nil {
						return fmt.Errorf("start xray err: %w", err)
					}

					repo.Status.UpdateRunningProcess("proxy", pid)
					err = repo.SaveStatus()
					if err != nil {
						return fmt.Errorf("save status err: %w", err)
					}
				default:
					return fmt.Errorf("don't know how to start a %s server", repo.Status.Protocl.ShortName())
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
					rye.PrintlnError("start pac err: %s", err.Error())
					return err
				}
				repo.Status.UpdateRunningProcess("pac", pid)
				err = repo.SaveStatus()
				if err != nil {
					rye.PrintlnError("update status err: %s", err.Error())
					return err
				}
			}

			rye.PrintlnInfo("started")
			return nil
		},
	}

	return cmd
}
