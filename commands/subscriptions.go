package commands

import (
	"fmt"

	"github.com/larryzhao/rye/repo"
	"github.com/spf13/cobra"
)

func NewSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			fmt.Println(r.Dir)
			return nil
		},
	}

	cmd.AddCommand(newSubscriptionsUpdateCmd())

	return cmd
}

func newSubscriptionsUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			err = r.UpdateSubscriptions()
			if err != nil {
				return err
			}
			return nil
		},
	}
}
