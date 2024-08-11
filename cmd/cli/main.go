package main

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/commands"
	"github.com/spf13/cobra"
)

var (
	Version string
	Build   string
)

func main() {
	root := commands.NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s %s\n", Version, Build)
			return nil
		},
	})

	err := root.Execute()
	if err != nil {
		rye.PrintlnError(err.Error())
	}
}
