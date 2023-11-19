package commands

import (
	"fmt"
	"os"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/repo"
	"github.com/spf13/cobra"
)

// Comand Run
//
// `rye run`
//
// rye main process
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			process := rye.NewProcess(repo.XrayConfigFile(), repo.PACFile())
			err = process.Start()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			return nil
		},
	}

	return cmd
}
