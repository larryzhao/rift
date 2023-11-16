package commands

import (
	"fmt"

	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/repo"
	"github.com/spf13/cobra"
)

// Comand Connect
//
// `rye connect <url>`
//
// changes the current outbound proxy to the server url points to.
func NewConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "connect",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			repo, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			// ctx := repo.NewContext(repo)
			// fmt.Println(ctx)
			// server, err := server.ParseURL(url)
			// if err != nil {
			// 	return err
			// }

			server, err := rye.ParseServerFromURL(url)
			if err != nil {
				return err
			}

			outbound, err := server.ToOutbound()
			if err != nil {
				return err
			}

			repo.V2RayConfig.SetOutbound("proxy", outbound)
			err = repo.V2RayConfig.Save()
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
