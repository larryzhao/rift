package commands

import (
	"context"

	"github.com/larryzhao/rift"
	"github.com/spf13/cobra"
)

// rift start
// rift stop
//     stop connection
// rift subscriptions
// rift connect <url>

// NewRootCmd
func NewRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "rift",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set verbose by --verbose flag
			v, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			rift.PrintVerbosly = v

			// load repo and set context
			r, err := rift.LoadRepo()
			if err != nil {
				rift.PrintlnError("load repo err: %s", err.Error())
				return err
			}

			ctx := context.WithValue(cmd.Context(), rift.CtxKeyRepo, r)
			cmd.SetContext(ctx)

			return nil
		},
	}

	cmd.AddCommand(NewConnectCmd())
	cmd.AddCommand(NewStartCmd())
	cmd.AddCommand(NewRestartCmd())
	cmd.AddCommand(NewStopCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewSubscriptionsCmd())
	cmd.AddCommand(NewServersCmd())
	cmd.AddCommand(NewPACCmd())
	cmd.AddCommand(NewDomainCmd())
	cmd.AddCommand(NewVersionCmd())

	cmd.PersistentFlags().BoolP("verbose", "v", false, "print verbosely")
	return &cmd
}
