package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(rye.Ver())
			return nil
		},
	}
}
