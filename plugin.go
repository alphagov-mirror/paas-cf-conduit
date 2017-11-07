package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"code.cloudfoundry.org/cli/plugin"
)

var (
	conn plugin.CliConnection
)

type Plugin struct {
	cmd *cobra.Command
}

func (p *Plugin) Run(c plugin.CliConnection, args []string) {
	conn = c // FIXME: this isn't great, can we pass into p.cmd somehow?
	// set defaults
	org, err := conn.GetCurrentOrg()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	p.cmd.PersistentFlags().Lookup("org").Value.Set(org.Name)
	space, err := conn.GetCurrentSpace()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	p.cmd.PersistentFlags().Lookup("space").Value.Set(space.Name)
	// parse
	p.cmd.SetArgs(args)
	if err := p.cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	meta := plugin.PluginMetadata{
		Name: "conduit",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 26,
			Build: 0,
		},
		Commands: []plugin.Command{},
	}
	for _, cmd := range p.cmd.Commands() {
		if cmd.Hidden {
			continue
		}
		opts := map[string]string{}
		meta.Commands = append(meta.Commands, plugin.Command{
			Name:     cmd.Name(),
			HelpText: cmd.Long,
			UsageDetails: plugin.Usage{
				Usage:   cmd.UsageString(),
				Options: opts,
			},
		})
	}
	return meta
}
