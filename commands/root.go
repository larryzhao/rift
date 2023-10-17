package commands

import "github.com/spf13/cobra"

// rye start
// rye stop
//     stop connection
// rye subscriptions
// rye connect <url>

// NewRootCmd
func NewRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "rye",
	}

	cmd.AddCommand(NewConnectCmd())
	return &cmd
}
