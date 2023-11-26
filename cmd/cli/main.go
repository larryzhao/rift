package main

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/commands"
	"github.com/spf13/cobra"
)

var (
	version string
	build   string
)

func main() {
	root := commands.NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s %s\n", version, build)
			return nil
		},
	})

	err := root.Execute()
	if err != nil {
		rye.PrintlnError(err.Error())
	}
}
