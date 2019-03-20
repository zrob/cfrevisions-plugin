package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"fmt"
	. "github.com/zrob/cfrevisions-plugin/models"
	. "github.com/zrob/cfrevisions-plugin/util"
	"errors"
	"sort"
	"strings"
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
	if args[0] == "rollback" {
		if len(args) != 3 {
			fmt.Println(c.GetMetadata().Commands[2].UsageDetails.Usage)
		} else {
			c.rollback(cliConnection, args)
		}
	}
	if args[0] == "enable-revisions" {
		if len(args) != 2 {
			fmt.Println(c.GetMetadata().Commands[3].UsageDetails.Usage)
		} else {
			c.enableRevisions(cliConnection, args)
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
			{
				Name:     "rollback",
				HelpText: "Rollback to a previous revision",
				UsageDetails: plugin.Usage{
					Usage: "cf rollback APPNAME VERSION",
				},
			},
			{
				Name:     "enable-revisions",
				HelpText: "Enable an app to create and use revisions",
				UsageDetails: plugin.Usage{
					Usage: "cf enable-revisions APPNAME",
				},
			},
		},
	}
}

func (c *CFRevisionsPlugin) showRevisions(cliConnection plugin.CliConnection, args []string) {
	app := args[1]

	appGuid, err := getAppGuid(cliConnection, app)
	FreakOut(err)

	var revisions []RevisionModel
	morePages := true
	url := fmt.Sprintf("v3/apps/%s/revisions", appGuid)

	for morePages {
		output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", url)
		FreakOut(err)
		response := stringifyCurlResponse(output)
		revisionsPage := RevisionsModel{}
		err = json.Unmarshal([]byte(response), &revisionsPage)
		FreakOut(err)

		if revisionsPage.Next.Href != "" {
			p := strings.Split(revisionsPage.Next.Href, "v3")
			url = fmt.Sprintf("v3%s", p[1])
		} else {
			morePages = false
		}

		for _, r := range revisionsPage.Resources {
			revisions = append(revisions, r)
		}
	}

	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i].Version > revisions[j].Version
	})

	table := NewTable([]string{"version", "droplet"})
	fmt.Printf("Displaying revisions for app %s\r\n\r\n", app)
	for _, revision := range revisions {
		table.Add(fmt.Sprintf("%v", revision.Version), revision.Droplet.Guid)
	}
	table.Print()
}

func (c *CFRevisionsPlugin) showRevisionDetails(cliConnection plugin.CliConnection, args []string) {
	app := args[1]
	version := args[2]

	appGuid, err := getAppGuid(cliConnection, app)
	FreakOut(err)

	// get revision
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/apps/%s/revisions?versions=%s", appGuid, version))
	FreakOut(err)
	response := stringifyCurlResponse(output)
	revisions := RevisionsModel{}
	err = json.Unmarshal([]byte(response), &revisions)
	FreakOut(err)

	revision := revisions.Resources[0]

	// get revision env vars
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/revisions/%s/environment_variables", revision.Guid))
	FreakOut(err)
	response = stringifyCurlResponse(output)
	env := EnvVars{}
	err = json.Unmarshal([]byte(response), &env)
	FreakOut(err)

	fmt.Printf("Displaying revision details for revision %v of app %s\r\n\r\n", version, app)
	fmt.Printf("version: %v\r\n", revision.Version)
	fmt.Printf("droplet: %s\r\n", revision.Droplet.Guid)
	fmt.Println("")
	fmt.Println("environment variables")
	if len(env.Var) == 0 {
		fmt.Println("none")
	} else {
		for k, v := range env.Var {
			fmt.Printf("%s: %s\r\n", k, v)
		}
	}
}

func (c *CFRevisionsPlugin) rollback(cliConnection plugin.CliConnection, args []string) {
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

	revisionGuid := revisions.Resources[0].Guid

	fmt.Printf("Rolling back app %s to version %v...\r\n\r\n", app, version)

	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/deployments", "-X", "POST", "-d",
		fmt.Sprintf(`{"revision":{"guid":"%s"},"relationships":{"app":{"data":{"guid":"%s"}}}`, revisionGuid, appGuid))
	FreakOut(err)
	response = stringifyCurlResponse(output)
	deployment := DeploymentModel{}
	err = json.Unmarshal([]byte(response), &deployment)
	FreakOut(err)

	if deployment.Guid == "" {
		errors := ErrorsModel{}
		err = json.Unmarshal([]byte(response), &errors)
		FreakOut(err)

		fmt.Printf("Failed to initiate rollback: %s\r\n", errors.Errors[0].Detail)
		return
	}

	fmt.Println("Succeeded. Deployment in progress.")
}

func (c *CFRevisionsPlugin) enableRevisions(cliConnection plugin.CliConnection, args []string) {
	app := args[1]

	appGuid, err := getAppGuid(cliConnection, app)
	FreakOut(err)

	fmt.Printf("Enabling revisions for app %s\r\n", app)

	_, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/apps/%s/features/revisions", appGuid),
		"-X", "PATCH", "-d", `{"enabled":true}`)
	FreakOut(err)

	fmt.Println("OK")
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
