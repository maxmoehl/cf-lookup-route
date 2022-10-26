package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// region Api

type Api struct {
	endpoint string
	token    string
	client   http.Client
}

func NewApi(endpoint string, token string) *Api {
	return &Api{
		endpoint: endpoint,
		token:    token,
		client:   http.Client{},
	}
}

// do is an internal helper to add the authorization token to every request
// made.
func (api *Api) do(req *http.Request, parameters map[string][]string) (*http.Response, error) {
	req.Header.Set("Authorization", api.token)
	req.URL.RawQuery = url.Values(parameters).Encode()

	// for debugging:
	// fmt.Println(req.URL.String())

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		errorMsg, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err.Error())
		}
		return nil, fmt.Errorf("received non 200 status code: %d (%s)", res.StatusCode, string(errorMsg))
	}

	return res, nil
}

// get is an internal helper function which only needs the path of the request,
// the base URL will be added by this function. It then calls do to perform the
// request.
func (api *Api) get(path string, parameters map[string][]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, api.endpoint+path, nil)
	if err != nil {
		return nil, err
	}
	return api.do(req, parameters)
}

// GetDomains returns the domains visible by the current token.
//
// TODO: implement pagination
func (api *Api) GetDomains(names []string) (domains []Domain, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("get domains: %w", err)
		}
	}()

	parameters := map[string][]string{
		"names": names,
	}

	res, err := api.get("/v3/domains", parameters)
	if err != nil {
		return nil, err
	}

	var domainsRes DomainsResponse
	err = json.NewDecoder(res.Body).Decode(&domainsRes)
	if err != nil {
		return nil, err
	}

	return domainsRes.Resources, err
}

// GetRoutes returns the routes visible by the current token.
//
// TODO: implement pagination
func (api *Api) GetRoutes(hosts []string, domainGuids []string) (routes []Route, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("get routes: %w", err)
		}
	}()

	parameters := map[string][]string{
		"hosts":        hosts,
		"domain_guids": domainGuids,
	}

	res, err := api.get("/v3/routes", parameters)
	if err != nil {
		return nil, err
	}

	var routesRes RoutesResponse
	err = json.NewDecoder(res.Body).Decode(&routesRes)
	if err != nil {
		return nil, err
	}

	return routesRes.Resources, err
}

// GetApp returns a single app by its guid.
func (api *Api) GetApp(guid string) (app App, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("get app: %w", err)
		}
	}()

	res, err := api.get("/v3/apps/"+guid, nil)
	if err != nil {
		return App{}, err
	}

	err = json.NewDecoder(res.Body).Decode(&app)
	if err != nil {
		return App{}, err
	}

	return app, nil
}

// GetSpace returns a single space by its guid.
func (api *Api) GetSpace(guid string) (space Space, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("get app: %w", err)
		}
	}()

	res, err := api.get("/v3/spaces/"+guid, nil)
	if err != nil {
		return Space{}, err
	}

	err = json.NewDecoder(res.Body).Decode(&space)
	if err != nil {
		return Space{}, err
	}

	return space, nil
}

// GetOrganization returns a single organization by its guid.
func (api *Api) GetOrganization(guid string) (organization Organization, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("get organization: %w", err)
		}
	}()

	res, err := api.get("/v3/organizations/"+guid, nil)
	if err != nil {
		return Organization{}, err
	}

	err = json.NewDecoder(res.Body).Decode(&organization)
	if err != nil {
		return Organization{}, err
	}

	return organization, nil
}

// endregion

// region Helpers

type link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type pagination struct {
	TotalResults int  `json:"total_results"`
	TotalPages   int  `json:"total_pages"`
	Next         link `json:"next"`
	Previous     link `json:"previous"`
}

// endregion

// region Domains

type Domain struct {
	Guid               string      `json:"guid"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	Name               string      `json:"name"`
	Internal           bool        `json:"internal"`
	RouterGroup        interface{} `json:"router_group"`
	SupportedProtocols []string    `json:"supported_protocols"`
	Relationships      struct {
		Organization struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"organization"`
	} `json:"relationships"`
	Links struct {
		Self                link `json:"self"`
		RouteReservations   link `json:"route_reservations"`
		Organization        link `json:"organization,omitempty"`
		SharedOrganizations link `json:"shared_organizations,omitempty"`
	} `json:"links"`
}

type DomainsResponse struct {
	Pagination pagination `json:"pagination"`
	Resources  []Domain   `json:"resources"`
}

// endregion

// region Routes

type Route struct {
	Guid         string    `json:"guid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Protocol     string    `json:"protocol"`
	Host         string    `json:"host"`
	Path         string    `json:"path"`
	Port         *int      `json:"port"`
	Url          string    `json:"url"`
	Destinations []struct {
		Guid string `json:"guid"`
		App  struct {
			Guid    string `json:"guid"`
			Process struct {
				Type string `json:"type"`
			} `json:"process"`
		} `json:"app"`
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	} `json:"destinations"`
	Relationships struct {
		Space struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"space"`
		Domain struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"domain"`
	} `json:"relationships"`
	Links struct {
		Self         link `json:"self"`
		Space        link `json:"space"`
		Destinations link `json:"destinations"`
		Domain       link `json:"domain"`
	} `json:"links"`
}

type RoutesResponse struct {
	Pagination pagination `json:"pagination"`
	Resources  []Route    `json:"resources"`
}

// endregion

// region Apps

type App struct {
	Guid      string    `json:"guid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	State     string    `json:"state"`
	Lifecycle struct {
		Type string `json:"type"`
		Data struct {
			Buildpacks []string `json:"buildpacks"`
			Stack      string   `json:"stack"`
		} `json:"data"`
	} `json:"lifecycle"`
	Relationships struct {
		Space struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"space"`
	} `json:"relationships"`
	Links struct {
		Self                 link `json:"self"`
		EnvironmentVariables link `json:"environment_variables"`
		Space                link `json:"space"`
		Processes            link `json:"processes"`
		Packages             link `json:"packages"`
		CurrentDroplet       link `json:"current_droplet"`
		Droplets             link `json:"droplets"`
		Tasks                link `json:"tasks"`
		Start                link `json:"start"`
		Stop                 link `json:"stop"`
		Revisions            link `json:"revisions"`
		DeployedRevisions    link `json:"deployed_revisions"`
		Features             link `json:"features"`
	} `json:"links"`
}

type AppsResponse struct {
	Pagination pagination `json:"pagination"`
	Resources  []App      `json:"resources"`
}

// endregion
// region Spaces

type Space struct {
	Guid          string    `json:"guid"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Relationships struct {
		Organization struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"organization"`
		Quota struct {
			Data *struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"quota"`
	} `json:"relationships"`
	Links struct {
		Self          link `json:"self"`
		Organization  link `json:"organization"`
		Features      link `json:"features"`
		ApplyManifest link `json:"apply_manifest"`
		Quota         link `json:"quota,omitempty"`
	} `json:"links"`
}

type SpacesResponse struct {
	Pagination pagination `json:"pagination"`
	Resources  []Space    `json:"resources"`
}

// endregion
// region Organization

type Organization struct {
	Guid          string    `json:"guid"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Suspended     bool      `json:"suspended"`
	Relationships struct {
		Quota struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"quota"`
	} `json:"relationships"`
	Links struct {
		Self          link `json:"self"`
		Domains       link `json:"domains"`
		DefaultDomain link `json:"default_domain"`
		Quota         link `json:"quota"`
	} `json:"links"`
}

type OrganizationsResponse struct {
	Pagination pagination     `json:"pagination"`
	Resources  []Organization `json:"resources"`
}

// endregion
