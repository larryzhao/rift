package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/singbox"
	"github.com/spf13/cobra"
)

// Comand Run
//
// `rye run proxy`
//
// foreground entrypoint that hosts the sing-box instance in the current
// process. `rye start` forks this command so the PID can be tracked.
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing run target")
			}
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)
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
