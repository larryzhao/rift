package commands

import (
	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/repo"
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
			r, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			pid, err := rye.StartRunner()
			if err != nil {
				return err
			}

			if err := r.WritePID(pid); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
