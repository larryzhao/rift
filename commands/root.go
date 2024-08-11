package commands

import (
	"context"

	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

// rye start
// rye stop
//     stop connection
// rye subscriptions
// rye connect <url>

// NewRootCmd
func NewRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "rye",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set verbose by --verbose flag
			v, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			rye.PrintVerbosly = v

			// load repo and set context
			r, err := rye.LoadRepo()
			if err != nil {
				rye.PrintlnError("load repo err: %s", err.Error())
				return err
			}

			ctx := context.WithValue(cmd.Context(), rye.CtxKeyRepo, r)
			cmd.SetContext(ctx)

			return nil
		},
	}

	cmd.AddCommand(NewConnectCmd())
	cmd.AddCommand(NewStartCmd())
	cmd.AddCommand(NewStopCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewSubscriptionsCmd())
	cmd.AddCommand(NewServersCmd())
	cmd.AddCommand(NewPACCmd())
	cmd.AddCommand(NewVersionCmd())

	cmd.PersistentFlags().BoolP("verbose", "v", false, "print verbosely")
	return &cmd
}
