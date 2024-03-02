package main

import (
	"os"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/sites"
	"github.com/spf13/cobra"
)

const (
	AppGroupID string = "proxy-collector"
)

type App struct {
	rootCmd *cobra.Command
	runner  *SiteRunner
	cnf     *confs.CollectorConf
}

func NewApp() (a *App) {
	cnf := confs.NewCollectorConf()
	a = &App{
		rootCmd: &cobra.Command{},
		runner:  NewSiteRunner(cnf),
		cnf:     cnf,
	}
	a.rootCmd.AddGroup(&cobra.Group{ID: AppGroupID, Title: "Proxy Collector Commands: "})
	a.initiate()
	return
}

func (a *App) initiate() {
	a.rootCmd.AddCommand(&cobra.Command{
		Use:     "add-domain",
		Aliases: []string{"ad"},
		GroupID: AppGroupID,
		Short:   "Adds rawDomains to rawDomain list.",
		Run: func(cmd *cobra.Command, args []string) {
			a.cnf.AddRawDomains(args...)
		},
	})

	a.rootCmd.AddCommand(&cobra.Command{
		Use:     "add-subscribedUrls",
		Aliases: []string{"as"},
		GroupID: AppGroupID,
		Short:   "Adds urls to subscribedUrl list.",
		Run: func(cmd *cobra.Command, args []string) {
			a.cnf.AddSubs(args...)
		},
	})

	enableJsdelivr := "jsdelivr"
	enableProxy := "proxy"
	getProxiesCmd := &cobra.Command{
		Use:     "get-proxies",
		Aliases: []string{"gp"},
		GroupID: AppGroupID,
		Short:   "Collects proxies.",
		Run: func(cmd *cobra.Command, args []string) {
			if eJsdelivr, _ := cmd.Flags().GetBool(enableJsdelivr); eJsdelivr {
				os.Setenv(confs.ToEnableJsdelivrEnvName, "true")
			}
			if eProxy, _ := cmd.Flags().GetBool(enableProxy); eProxy {
				os.Setenv(confs.ToEnableProxyEnvName, "true")
			}
			if a.runner != nil {
				a.runner.AddSite(sites.NewSubVPN(a.cnf))
				a.runner.AddSite(sites.NewFreeFQVPN(a.cnf))
				a.runner.Run()
			}
		},
	}
	getProxiesCmd.Flags().BoolP(enableJsdelivr, "j", true, "Enables jsdelivr CDN.")
	getProxiesCmd.Flags().BoolP(enableProxy, "p", false, "Enables proxy.")
	a.rootCmd.AddCommand(getProxiesCmd)

	getEDomains := &cobra.Command{
		Use:     "test-domains",
		Aliases: []string{"td"},
		GroupID: AppGroupID,
		Short:   "Tests domains for edgetunnels.",
		Run: func(cmd *cobra.Command, args []string) {
			if eProxy, _ := cmd.Flags().GetBool(enableProxy); eProxy {
				os.Setenv(confs.ToEnableProxyEnvName, "true")
			}
			if a.runner != nil {
				a.runner.AddSite(sites.NewEDCollector(a.cnf))
				a.runner.AddSite(sites.NewEDomains(a.cnf))
				a.runner.Run()
			}
		},
	}
	getEDomains.Flags().BoolP(enableProxy, "p", false, "Enables proxy.")
	a.rootCmd.AddCommand(getEDomains)

	a.rootCmd.AddCommand(&cobra.Command{
		Use:     "show-rawdomains",
		Aliases: []string{"sr"},
		GroupID: AppGroupID,
		Short:   "Shows rawDomain list.",
		Run: func(cmd *cobra.Command, args []string) {
			a.cnf.ShowRawDomains()
		},
	})

	a.rootCmd.AddCommand(&cobra.Command{
		Use:     "show-subscribedurls",
		Aliases: []string{"ss"},
		GroupID: AppGroupID,
		Short:   "Shows subscribed urls.",
		Run: func(cmd *cobra.Command, args []string) {
			a.cnf.ShowSubs()
		},
	})

	a.rootCmd.AddCommand(&cobra.Command{
		Use:     "show-cryptokey",
		Aliases: []string{"sk"},
		GroupID: AppGroupID,
		Short:   "Shows cryptoKey.",
		Run: func(cmd *cobra.Command, args []string) {
			a.cnf.ShowCryptoKey()
		},
	})

	a.rootCmd.AddCommand(&cobra.Command{
		Use:     "reset-cryptokey",
		Aliases: []string{"rk"},
		GroupID: AppGroupID,
		Short:   "Resets cryptoKey.",
		Run: func(cmd *cobra.Command, args []string) {
			a.cnf.ResetCryptoKey()
		},
	})

	a.rootCmd.AddCommand(&cobra.Command{
		Use:     "set-localproxy",
		Aliases: []string{"sl"},
		GroupID: AppGroupID,
		Short:   "Sets local proxy for fetcher.",
		Long:    "Example: pxy sl <xxx>",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			a.cnf.SetLocalProxy(args[0])
		},
	})
}

func (a *App) Run() {
	if err := a.rootCmd.Execute(); err != nil {
		gprint.PrintError("%+v", err)
	}
}
