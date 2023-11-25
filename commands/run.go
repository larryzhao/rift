package commands

import (
	"os"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/repo"
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
			repo, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			logFile, err := os.OpenFile("/Users/larry/.rye/rye.log", os.O_RDWR|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			defer logFile.Close()

			runner := rye.NewRunner(repo.XrayConfigFile(), repo.PACFile(), logFile)
			err = runner.Run()
			if err != nil {
				rye.PrintError(err.Error())
				os.Exit(1)
			}

			return nil
		},
	}

	return cmd
}
