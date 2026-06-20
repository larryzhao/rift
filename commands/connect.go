package commands

import (
	"github.com/spf13/cobra"
)

// Comand Connect
//
// `rift connect <url>`
//
// changes the current outbound proxy to the server url points to.
func NewConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "connect",
		RunE: func(cmd *cobra.Command, args []string) error {
			// repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			// url := args[0]

			// server, err := rift.ParseServerFromURL(url)
			// if err != nil {
			// 	rift.PrintlnError("parse server error: %s", err.Error())
			// 	return err
			// }

			// outbound, err := server.ToOutbound()
			// if err != nil {
			// 	rift.PrintlnError("build outbound from server err: %s", err.Error())
			// 	return err
			// }

			// repo.XrayConfig.SetOutbound("proxy", outbound)
			// err = repo.XrayConfig.Save()
			// if err != nil {
			// 	rift.PrintlnError("update xray/config.json err: %s", err.Error())
			// 	return err
			// }

			// rift.PrintlnInfo("xray config updated")
			return nil
		},
	}

	return cmd
}
