package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/larryzhao/rift"
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
	cmd.AddCommand(newSubscriptionsAddCmd())

	return cmd
}

func newSubscriptionsUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			updatedSubs, err := repo.UpdateSubscriptions()
			if err != nil {
				return err
			}

			var updatedSubNames []string
			for _, sub := range updatedSubs {
				updatedSubNames = append(updatedSubNames, sub.Name)
			}
			rift.PrintlnInfo("%s updated", strings.Join(updatedSubNames, ","))

			return nil
		},
	}
}

func newSubscriptionsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use: "add",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			name := args[0]
			url := args[1]

			if name == "" {
				return fmt.Errorf("you need to provide a name for the subscription")
			}

			if url == "" {
				return fmt.Errorf("you need to provide a url for the subscription")
			}

			sub := &rift.Subscription{
				Name:       name,
				URL:        url,
				AddedAt:    time.Now(),
				SkipUpdate: false,
			}

			err := repo.AddSubscription(sub)
			if err != nil {
				return err
			}

			rift.PrintlnInfo("Subscription %s added.", name)
			return nil
		},
	}
}
