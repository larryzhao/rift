package commands

import (
	"fmt"

	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/pac"
	"github.com/spf13/cobra"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			// stop runners
			for _, proc := range repo.Status.RunningProcesses {
				err := rift.Stop(proc.PID)
				if err != nil {
					rift.PrintlnError("stop %s process %d err: %s", proc.Kind, proc.PID, err.Error())
				}
			}

			// remove PAC settings from network
			err := pac.RemoveSystemPAC()
			if err != nil {
				return fmt.Errorf("turn off system proxy err: %w", err)
			}

			rift.PrintlnInfo("stopped")
			return nil
		},
	}
}
