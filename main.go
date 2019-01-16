package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
)

type CFRevisionsPlugin struct{}

func main() {
	plugin.Start(new(CFRevisionsPlugin))
}

func (c *CFRevisionsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "hello" {
		fmt.Println("hello world")
	}
}

func (c *CFRevisionsPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cfrevisions",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "hello",
				HelpText: "sample help text",
				UsageDetails: plugin.Usage{
					Usage: "cf hello",
				},
			},
		},
	}
}
