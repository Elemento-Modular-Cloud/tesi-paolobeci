package metadata

// COMMENT: useful?

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const Endpoint = "http://..."  // TODO: set the correct endpoint

// Client is a client for the Elemento Cloud Server Metadata Endpoints.
type Client struct {
	endpoint string
	timeout  time.Duration

	httpClient              *http.Client
	// potential promehteus implementation
}

// A ClientOption is used to configure a [Client].
type ClientOption func(*Client)

// WithEndpoint configures a [Client] to use the specified Metadata API endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(client *Client) {
		client.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithHTTPClient configures a [Client] to perform HTTP requests with httpClient.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithTimeout specifies a time limit for requests made by this [Client]. Defaults to 5 seconds.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.timeout = timeout
	}
}

// NewClient creates a new [Client] with the options applied.
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		endpoint:   Endpoint,
		httpClient: &http.Client{},
		timeout:    5 * time.Second,
	}

	for _, option := range options {
		option(client)
	}

	client.httpClient.Timeout = client.timeout

	return client
}

// get executes an HTTP request against the API.
func (c *Client) get(path string) (string, error) {
	ctx := context.Background()
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+path, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body := string(bytes.TrimSpace(bodyBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("response status was %d", resp.StatusCode)
	}
	return body, nil
}

// Hostname returns the hostname of the server that did the request to the Metadata server.
func (c *Client) Hostname() (string, error) {
	return c.get("/hostname")
}

// InstanceID returns the ID of the server that did the request to the Metadata server.
func (c *Client) InstanceID() (int64, error) {
	resp, err := c.get("/instance-id")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(resp, 10, 64)
}

// PublicIPv4 returns the Public IPv4 of the server that did the request to the Metadata server.
func (c *Client) PublicIPv4() (net.IP, error) {
	resp, err := c.get("/public-ipv4")
	if err != nil {
		return nil, err
	}
	return net.ParseIP(resp), nil
}

// Region returns the Network Zone of the server that did the request to the Metadata server.
func (c *Client) Region() (string, error) {
	return c.get("/region")
}

// AvailabilityZone returns the datacenter of the server that did the request to the Metadata server.
func (c *Client) AvailabilityZone() (string, error) {
	return c.get("/availability-zone")
}

// PrivateNetworks returns details about the private networks the server is attached to.
// Returns YAML (unparsed).
func (c *Client) PrivateNetworks() (string, error) {
	return c.get("/private-networks")
}