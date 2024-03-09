package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/hysteria2"
	"github.com/spf13/cobra"
)

// Comand Run
//
// `rye run`
//
// start rye runner
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)
			switch args[0] {
			case "pac":

			case "hysteria2":
				runner := hysteria2.Runner{
					Bin:    "/opt/homebrew/bin/hysteria",
					Config: repo.HysteriaConfigFile(),
				}

				pid, err := runner.Run()
				if err != nil {
					return err
				}

				repo.Status.Protocl = rye.ProtoclHysteria2
				repo.Status.UpdateRunningProcess("proxy", pid)
				err = repo.SaveStatus()
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf("invalid arg: %s", args[0])
			}
			return nil
		},
	}

	return cmd
}
