package commands

import (
	"strings"

	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

func NewSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			updatedSubs, err := repo.UpdateSubscriptions()
			if err != nil {
				return err
			}

			var updatedSubNames []string
			for _, sub := range updatedSubs {
				updatedSubNames = append(updatedSubNames, sub.Name)
			}
			rye.PrintlnInfo("%s updated", strings.Join(updatedSubNames, ","))

			return nil
		},
	}
}
