package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/server"
	"github.com/spf13/cobra"
)

// connect command implementation
// `rye connect url`
// changes the current outbound proxy to the server url points to.
func NewConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "connect",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			repo, err := rye.LoadRepo()
			if err != nil {
				return err
			}

			ctx := rye.NewContext(repo)
			fmt.Println(ctx)

			server, err := server.ParseURL(url)
			if err != nil {
				return err
			}

			fmt.Printf("server: %v", server)

			// server, err := rye.NewServer(url)
			// if err != nil {
			// 	return err
			// }
			return nil
		},
	}

	return cmd
}
