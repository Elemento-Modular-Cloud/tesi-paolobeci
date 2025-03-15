package ecloud

import (
	"fmt"
	"net/http"
	"time"
)

type Response struct {
	*http.Response
	// Meta Meta TODO

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

type Client struct {
	endpoint                string
	token                   string
	tokenValid              bool
	retryMaxRetries         int
	timeout 			    time.Duration
	httpClient              *http.Client
	applicationName         string
	applicationVersion      string
	userAgent               string

	// TODO

	// Firewall         FirewallClient
	// Location         LocationClient
	// Network          NetworkClient
	// Pricing          PricingClient
	// SSHKey           SSHKeyClient
	// Logger 			 Logger
}

func NewClient(endpoint string, token string, applicationName string, applicationVersion string) *Client {
	return &Client{
		endpoint:           endpoint,
		token:              token,
		tokenValid:         false,
		retryMaxRetries:    3,
		timeout:            30 * time.Second,
		httpClient:         &http.Client{},
		applicationName:    applicationName,
		applicationVersion: applicationVersion,
		userAgent:          fmt.Sprintf("%s/%s", applicationName, applicationVersion),
	}
	// TODO: research real data needed for the client
}
