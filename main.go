package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"fmt"
	. "github.com/zrob/cfrevisions-plugin/models"
	. "github.com/zrob/cfrevisions-plugin/util"
	"errors"
)

type CFRevisionsPlugin struct{}

func main() {
	plugin.Start(new(CFRevisionsPlugin))
}

func (c *CFRevisionsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "revisions" {
		if len(args) != 2 {
			fmt.Println(c.GetMetadata().Commands[0].UsageDetails.Usage)
		} else {
			c.showRevisions(cliConnection, args)
		}
	}
	if args[0] == "revision" {
		if len(args) != 3 {
			fmt.Println(c.GetMetadata().Commands[1].UsageDetails.Usage)
		} else {
			c.showRevisionDetails(cliConnection, args)
		}
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
				Name:     "revisions",
				HelpText: "Display revisions for an app",
				UsageDetails: plugin.Usage{
					Usage: "cf revisions APPNAME",
				},
			},
			{
				Name:     "revision",
				HelpText: "Display a revision's details",
				UsageDetails: plugin.Usage{
					Usage: "cf revision APPNAME VERSION",
				},
			},
		},
	}
}

func (c *CFRevisionsPlugin) showRevisions(cliConnection plugin.CliConnection, args []string) {
	app := args[1]

	appGuid, err := getAppGuid(cliConnection, app)
	FreakOut(err)

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/apps/%s/revisions", appGuid))
	FreakOut(err)
	response := stringifyCurlResponse(output)
	revisions := RevisionsModel{}
	err = json.Unmarshal([]byte(response), &revisions)
	FreakOut(err)

	table := NewTable([]string{"version", "droplet"})
	fmt.Printf("Displaying revisions for app %s\r\n\r\n", app)
	for _, revision := range revisions.Resources {
		table.Add(fmt.Sprintf("%v", revision.Version), revision.Droplet.Guid)
	}
	table.Print()
}

func (c *CFRevisionsPlugin) showRevisionDetails(cliConnection plugin.CliConnection, args []string) {
	app := args[1]
	version := args[2]

	appGuid, err := getAppGuid(cliConnection, app)
	FreakOut(err)

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/apps/%s/revisions?versions=%s", appGuid, version))
	FreakOut(err)
	response := stringifyCurlResponse(output)
	revisions := RevisionsModel{}
	err = json.Unmarshal([]byte(response), &revisions)
	FreakOut(err)

	revision := revisions.Resources[0]

	fmt.Printf("Displaying revision details for revision %v of app %s\r\n\r\n", version, app)
	fmt.Printf("version: %v\r\n", revision.Version)
	fmt.Printf("droplet: %s\r\n", revision.Droplet.Guid)
}

func getAppGuid(cliConnection plugin.CliConnection, app string) (appGuid string, err error) {
	mySpace, _ := cliConnection.GetCurrentSpace()

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v2/apps?q=name:%s&q=space_guid:%s", app, mySpace.Guid))
	if err != nil {
		return
	}
	response := stringifyCurlResponse(output)

	apps := AppsModel{}
	err = json.Unmarshal([]byte(response), &apps)
	if (err != nil) {
		return
	}

	if len(apps.Resources) == 0 {
		err = errors.New(fmt.Sprintf("App %s not found", app))
		return
	}

	appGuid = apps.Resources[0].Metadata.Guid
	return
}

func stringifyCurlResponse(output []string) string {
	var responseString string
	for _, part := range output {
		responseString += part
	}
	return responseString
}
