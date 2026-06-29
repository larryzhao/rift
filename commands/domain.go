package commands

import (
	"fmt"

	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/pac"
	"github.com/spf13/cobra"
)

func NewDomainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain",
		Short: "Manage per-domain proxy rules (domains.txt)",
	}

	cmd.AddCommand(newDomainProxyCmd())
	cmd.AddCommand(newDomainDirectCmd())
	cmd.AddCommand(newDomainStatusCmd())

	return cmd
}

func newDomainProxyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "proxy <domain>",
		Short: "Route a domain through the proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)
			d := pac.NormalizeDomain(args[0])
			if err := pac.SetDomainRule(repo.PACDomainsFile(), d, false); err != nil {
				return err
			}
			fmt.Printf("%s -> PROXY (added to domains.txt)\n", d)
			return nil
		},
	}
}

func newDomainDirectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "direct <domain>",
		Short: "Route a domain directly (no proxy)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)
			d := pac.NormalizeDomain(args[0])
			if err := pac.SetDomainRule(repo.PACDomainsFile(), d, true); err != nil {
				return err
			}
			fmt.Printf("%s -> DIRECT (added to domains.txt)\n", d)
			return nil
		},
	}
}

func newDomainStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <domain>",
		Short: "Show whether a domain is proxied or direct under current rules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)
			res, err := pac.Lookup(repo.PACDomainsFile(), repo.PACGFWListFile(), args[0])
			if err != nil {
				return err
			}

			decision := "DIRECT"
			if res.Proxied {
				decision = "PROXY"
			}

			var reason string
			switch res.Source {
			case "domains.txt":
				reason = fmt.Sprintf("matched domains.txt rule %q", res.Rule)
			case "gfwlist":
				reason = fmt.Sprintf("matched gfwlist rule %q", res.Rule)
			case "gfwlist-whitelist":
				reason = fmt.Sprintf("matched gfwlist whitelist rule %q", res.Rule)
			default:
				reason = "no rule matched (default)"
			}

			fmt.Printf("%s -> %s\n  %s\n", res.Host, decision, reason)
			return nil
		},
	}
}
