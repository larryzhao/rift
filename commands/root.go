package commands

import (
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
			v, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			rye.PrintVerbosly = v
			return nil
		},
	}

	cmd.AddCommand(NewConnectCmd())
	cmd.AddCommand(NewStartCmd())
	cmd.AddCommand(NewStopCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewSubscriptionsCmd())

	cmd.PersistentFlags().BoolP("verbose", "v", false, "print verbosely")
	return &cmd
}
