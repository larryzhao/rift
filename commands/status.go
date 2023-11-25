package commands

import (
	"fmt"

	"github.com/larryzhao/rye/repo"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			fmt.Println(r.Dir)
			return nil
		},
	}
}
