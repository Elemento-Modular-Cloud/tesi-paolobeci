package metadata

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const Endpoint = "http://127.0.0.1/ecloud/v1/metadata"

// Client is a client for the Hetzner Cloud Server Metadata Endpoints.
type Client struct {
	endpoint string
	timeout  time.Duration

	httpClient              *http.Client
	instrumentationRegistry string
}

// A ClientOption is used to configure a [Client].
type ClientOption func(*Client)

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

	if client.instrumentationRegistry != "" {
		// i := instrumentation.New("metadata", client.instrumentationRegistry)
		// not implemented
		client.httpClient.Transport = http.DefaultTransport
	}
	return client
}

// InstanceID returns the ID of the server that did the request to the Metadata server.
func (c *Client) InstanceID() (int, error) {
	resp, err := c.get("/instance-id")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(resp)
}

// get executes an HTTP request against the API.
func (c *Client) get(path string) (string, error) {
	url := c.endpoint + path
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("response status was %d", resp.StatusCode)
	}
	return body, nil
}
