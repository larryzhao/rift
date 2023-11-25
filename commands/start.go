package commands

import (
	"os"
	"os/exec"
	"path"

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

			errFile, err := os.OpenFile(path.Join(r.Dir, "error.log"), os.O_RDWR|os.O_APPEND, os.ModeAppend)
			if err != nil {
				return err
			}
			defer errFile.Close()

			outFile, err := os.OpenFile(path.Join(r.Dir, "out.log"), os.O_RDWR|os.O_APPEND, os.ModeAppend)
			if err != nil {
				return err
			}
			defer outFile.Close()

			command := exec.Command("/usr/local/bin/rye", "run")
			command.Stdout = outFile
			command.Stderr = errFile

			err = command.Start()
			if err != nil {
				panic(err)
			}

			pid := command.Process.Pid
			rye.PrintInfo("started")

			err = command.Process.Release()
			if err != nil {
				panic(err)
			}

			if err := r.WritePID(pid); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
