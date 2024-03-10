package commands

import (
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
			// repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			// url := args[0]

			// server, err := rye.ParseServerFromURL(url)
			// if err != nil {
			// 	rye.PrintlnError("parse server error: %s", err.Error())
			// 	return err
			// }

			// outbound, err := server.ToOutbound()
			// if err != nil {
			// 	rye.PrintlnError("build outbound from server err: %s", err.Error())
			// 	return err
			// }

			// repo.XrayConfig.SetOutbound("proxy", outbound)
			// err = repo.XrayConfig.Save()
			// if err != nil {
			// 	rye.PrintlnError("update xray/config.json err: %s", err.Error())
			// 	return err
			// }

			// rye.PrintlnInfo("xray config updated")
			return nil
		},
	}

	return cmd
}
