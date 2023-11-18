package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
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
			// repo, err := repo.LoadRepo()
			// if err != nil {
			// 	return err
			// }

			process := rye.NewProcess()
			pid, err := process.Start()
			if err != nil {
				return err
			}
			fmt.Println(pid)

			select {}
			return nil
		},
	}

	return cmd
}
