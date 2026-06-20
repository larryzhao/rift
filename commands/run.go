package commands

import (
	"fmt"

	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/singbox"
	"github.com/spf13/cobra"
)

// Comand Run
//
// `rift run proxy`
//
// foreground entrypoint that hosts the sing-box instance in the current
// process. `rift start` forks this command so the PID can be tracked.
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing run target")
			}
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)
			switch args[0] {
			case "pac":
				return nil
			case "proxy":
				return singbox.RunForeground(repo.SingboxConfigFile())
			default:
				return fmt.Errorf("invalid arg: %s", args[0])
			}
		},
	}

	return cmd
}
