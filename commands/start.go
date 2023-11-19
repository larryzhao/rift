package commands

import (
	"os"

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

			proc, err := os.StartProcess("/usr/local/bin/rye", []string{}, &os.ProcAttr{Dir: r.Dir})
			if err != nil {
				return err
			}
			if err := r.WritePID(proc.Pid); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
