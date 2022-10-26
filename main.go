package main

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

type item struct {
	Key   string
	Value string
	Id    string
}

type lookupRoute struct{}

func main() {
	plugin.Start(new(lookupRoute))
}

func (l lookupRoute) Run(cliConnection plugin.CliConnection, args []string) {
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			os.Exit(1)
		}
	}()

	// region preface

	hasApiEndpoint, err := cliConnection.HasAPIEndpoint()
	if err != nil {
		return
	}
	if !hasApiEndpoint {
		err = fmt.Errorf("no API endpoint set")
		return
	}

	apiEndpoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		err = fmt.Errorf("get cf api endpoint: %w", err)
		return
	}

	loggedIn, err := cliConnection.IsLoggedIn()
	if err != nil {
		return
	}
	if !loggedIn {
		err = fmt.Errorf("error: not logged in")
		return
	}

	apiToken, err := cliConnection.AccessToken()
	if err != nil {
		err = fmt.Errorf("get access token: %w", err)
		return
	}

	// endregion

	cc := NewApi(apiEndpoint, apiToken)

	if len(args) < 2 {
		fmt.Printf("nothing to do")
		return
	}

	query := args[1]

	// region route lookup

	parts := strings.SplitN(query, ".", 2)
	if len(parts) < 2 {
		err = fmt.Errorf("'%s' is not a domain", query)
		return
	}

	hostName := parts[0]
	domainName := parts[1]

	domains, err := cc.GetDomains([]string{domainName})
	if err != nil {
		return
	}
	if len(domains) == 0 {
		err = fmt.Errorf("found no matching domains")
		return
	}
	if len(domains) > 1 {
		err = fmt.Errorf("found multiple matching domains")
		return
	}

	domain := domains[0]

	routes, err := cc.GetRoutes([]string{hostName}, []string{domain.Guid})
	if err != nil {
		return
	}
	if len(routes) == 0 {
		err = fmt.Errorf("found no matching routes")
		return
	}
	if len(routes) > 1 {
		err = fmt.Errorf("found multiple matching routes")
		return
	}

	route := routes[0]

	if len(route.Destinations) == 0 {
		err = fmt.Errorf("route has no destination")
		return
	}
	if len(route.Destinations) > 1 {
		err = fmt.Errorf("route has multiple destination")
		return
	}

	app, err := cc.GetApp(route.Destinations[0].App.Guid)
	if err != nil {
		return
	}

	space, err := cc.GetSpace(app.Relationships.Space.Data.Guid)
	if err != nil {
		return
	}

	org, err := cc.GetOrganization(space.Relationships.Organization.Data.Guid)
	if err != nil {
		return
	}

	fmt.Printf("Organization: %s (%s)\n", org.Name, org.Guid)
	fmt.Printf("Space       : %s (%s)\n", space.Name, space.Guid)
	fmt.Printf("App         : %s (%s)\n", app.Name, app.Guid)

	// endregion
}

func (l lookupRoute) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "lookup-route",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "lookup-route",
				HelpText: "Lookup routes in the cloudfoundry API",
			},
		},
	}
}
