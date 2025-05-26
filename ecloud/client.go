package ecloud

// TODO: fallo simile ad hetzner

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Response struct {
	*http.Response
	Meta Meta

	// body holds a copy of the http.Response body that must be used within the handler
	// chain. The http.Response.Body is reserved for external users.
	body []byte
}

// Meson endpoints of supportd Cloud Providers
const (
	ArubaCloudEU   = "http://192.168.1.103" // = "https://eu.arubacloud.public.elemento.cloud/api/v1.0"
	OvhEU          = "http://192.168.1.103" // = "https://eu.ovh.public.elemento.cloud/api/v1.0"
	GigasEU        = "http://192.168.1.103" // = "https://eu.gigas.public.elemento.cloud/api/v1.0"
	IonosEU        = "http://192.168.1.103" // = "https://eu.ionos.public.elemento.cloud/api/v1.0"
)

var Endpoints = map[string]string{
	"arubacloud-eu":   ArubaCloudEU,
	"ovh-eu":          OvhEU,
	"gigas-eu":        GigasEU,
	"ionos-eu":        IonosEU,
}


// Local endpoints for deamons connection
const (
	AuthenticateRoute = "http://localhost:47777/api/v1/authenticate/"
	ComputeRoute	  = "http://localhost:17777/api/v1.0/client/vm/"
	StorageRoute	  = "http://localhost:27777/api/v1.0/client/volume/"
	NetworkRoute	  = "http://localhost:37777/"
)

var Routes = map[string]string{
	"authenticate": AuthenticateRoute,
	"compute":      ComputeRoute,
	"storage":      StorageRoute,
	"network":      NetworkRoute,
}

// Client represents a client to call the Elemento Cloud API
type Client struct {
	endpoint                string
	retryMaxRetries         int
	timeout 			    time.Duration
	httpClient              *http.Client
	applicationName         string
	applicationVersion      string
	userAgent               string
	logger 					Logger

	Server           		ServerClient
	Network		 			NetworkClient
	SSHKey          		SSHKeyClient

	// TODO
}

func NewClient(applicationName string, applicationVersion string) (*Client, error) {
	if applicationName == "" {
		return nil, fmt.Errorf("application name cannot be empty")
	}
	if applicationVersion == "" {
		return nil, fmt.Errorf("application version cannot be empty")
	}

	client := &Client{
		endpoint:           "http://127.0.0.1",
		retryMaxRetries:    3,
		timeout:            30 * time.Second,
		httpClient:         &http.Client{},
		applicationName:    applicationName,
		applicationVersion: applicationVersion,
		userAgent:          fmt.Sprintf("%s/%s", applicationName, applicationVersion),
	}

	client.Server = ServerClient{client: client}
	client.Network = NetworkClient{client: client}

	// TODO: research real data needed for the client

	return client, nil
}

// New Client creation from env variables
func NewClientFromEnv(path string) (*Client, error) {
	var client Client

	// Get and check the configuration
	if err := client.loadConfig(path); err != nil {
		return nil, err
	}
	return &client, nil
}

// ListOpts specifies options for listing resources.
type ListOpts struct {
	Page          int    // Page (starting at 1)
	PerPage       int    // Items per page (0 means default)
	LabelSelector string // Label selector for filtering by labels
}

func (c *Client) all(f func(int) (*Response, error)) error {
	var (
		page = 1
	)
	for {
		resp, err := f(page)
		if err != nil {
			return err
		}
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.NextPage == 0 {
			return nil
		}
		page = resp.Meta.Pagination.NextPage
	}
}

// Meta represents meta information included in an API response.
type Meta struct {
	Pagination *Pagination
	Ratelimit  Ratelimit
}

// Pagination represents pagination meta information.
type Pagination struct {
	Page         int
	PerPage      int
	PreviousPage int
	NextPage     int
	LastPage     int
	TotalEntries int
}

// Ratelimit represents ratelimit information.
type Ratelimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// Values returns the ListOpts as URL values.
func (l ListOpts) Values() url.Values {
	vals := url.Values{}
	if l.Page > 0 {
		vals.Add("page", strconv.Itoa(l.Page))
	}
	if l.PerPage > 0 {
		vals.Add("per_page", strconv.Itoa(l.PerPage))
	}
	if len(l.LabelSelector) > 0 {
		vals.Add("label_selector", l.LabelSelector)
	}
	return vals
}
