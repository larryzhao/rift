package main

import (
	"fmt"

	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/commands"
	"github.com/spf13/cobra"
)

func main() {
	root := commands.NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(rift.Ver())
			return nil
		},
	})

	err := root.Execute()
	if err != nil {
		rift.PrintlnError("%s", err.Error())
	}
}
