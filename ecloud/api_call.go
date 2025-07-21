package ecloud

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

// ------------------------------ API CALLS FUNCTIONS -------------------------

// Login to the API
func (c *Client) Login(reqBody interface{}) (*schema.LoginResponse, error) {
	var res schema.LoginResponse
	err := c.CallAPI("POST", "47777", "/api/v1/authenticate/login", reqBody, &res, false)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Status login
func (c *Client) StatusLogin() (*schema.StatusLoginResponse, error) {
	var res schema.StatusLoginResponse
	err := c.CallAPI("GET", "47777", "/api/v1/authenticate/status", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Logout from the API
func (c *Client) Logout() (*schema.LogoutResponse, error) {
	var res schema.LogoutResponse
	err := c.CallAPI("POST", "47777", "/api/v1/authenticate/logout", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Health check Compute
func (c *Client) HealthCheckCompute() (*schema.HealthCheckComputeResponse, error) {
	var res schema.HealthCheckComputeResponse
	err := c.CallAPI("GET", "17777", "/", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Can allocate a new compute instance
func (c *Client) CanAllocateCompute(reqBody schema.CanAllocateComputeRequest) (*schema.CanAllocateComputeResponse, error) {
	// Original API call code
	// var res schema.CanAllocateComputeResponse
	// err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/canallocate", reqBody, &res, true)
	// if err != nil {
	// 	return nil, err
	// }
	// return &res, nil

	// Mock response that always returns true since the API is not implemented yet
	res := &schema.CanAllocateComputeResponse{
		Mesos: []schema.ProviderInfo{
			{
				Price: schema.Price{
					Hour:  0.0,
					Month: 0.0,
					Unit:  "EUR",
				},
				Provider: "mock-provider",
				Region:   "eu-south",
			},
		},
	}
	return res, nil
}

// Create a new compute instance
func (c *Client) CreateCompute(reqBody schema.CreateComputeRequest) (*schema.CreateComputeResponse, error) {
	var res schema.CreateComputeResponse
	err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/register", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Compute instances status
func (c *Client) GetCompute() (*schema.GetComputeResponse, error) {
	var res schema.GetComputeResponse
	err := c.CallAPI("GET", "17777", "/api/v1.0/client/vm/status", nil, &res, true)
	if err != nil {
		return nil, err
	}

	// Convert RAM size from GB to MB for each server
	for i := range res {
		if res[i].ReqJSON.RamSize > 0 {
			res[i].ReqJSON.RamSize = res[i].ReqJSON.RamSize * 1024
		}
	}

	return &res, nil
}

// Compute templates
func (c *Client) ComputeTemplates() (*schema.ComputeTemplatesResponse, error) {
	var res schema.ComputeTemplatesResponse
	err := c.CallAPI("GET", "17777", "/api/v1.0/client/vm/templates", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Compute instance delete
func (c *Client) DeleteCompute(reqBody schema.DeleteComputeRequest) (*schema.DeleteComputeResponse, error) {
	var res schema.DeleteComputeResponse
	err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Health check Storage
func (c *Client) HealthCheckStorage() (*schema.HealthCheckStorageResponse, error) {
	var res schema.HealthCheckStorageResponse
	err := c.CallAPI("GET", "27777", "/", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Can create a new storage volume
func (c *Client) CanCreateStorage(reqBody schema.CanCreateStorageRequest) (*schema.CanCreateStorageResponse, error) {
	var res schema.CanCreateStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/cancreate", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create a new storage volume
func (c *Client) CreateStorage(reqBody schema.CreateStorageRequest) (*schema.CreateStorageResponse, error) {
	var res schema.CreateStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/create", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create new storage volume with specified image
func (c *Client) CreateStorageImage(reqBody schema.CreateStorageImageRequest) (*schema.CreateStorageImageResponse, error) {
	var res schema.CreateStorageImageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/cloudinit/create", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create new cloudinit volume
func (c *Client) CreateStorageCloudInit(reqBody schema.CreateStorageCloudInitRequest) (*schema.CreateStorageCloudInitResponse, error) {
	var res schema.CreateStorageCloudInitResponse

	// Marshal reqBody to JSON and encode as base64
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reqBody: %w", err)
	}
	encodedPayload := base64.StdEncoding.EncodeToString(jsonBytes)

	// Fixed file path
	filepath := "@/Users/paolob/Downloads/cloud-config.yml" // Change as needed
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Build URL with base64-encoded payload as last path segment
	urlPath := "/api/v1.0/client/volume/cloudinit/metadata/" + encodedPayload
	url := c.endpoint + ":27777" + urlPath

	// Create HTTP request with file bytes as body
	req, err := http.NewRequest("POST", url, bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &res, nil
}

func (c *Client) FeedFileIntoCloudInitStorage(reqBody schema.FeedFileIntoCloudInitStorageRequest) (string, error) {
	// Marshal reqBody to JSON and encode as base64
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal reqBody: %w", err)
	}
	encodedPayload := base64.StdEncoding.EncodeToString(jsonBytes)

	// Fixed file path
	filepath := "@/Users/paolob/Downloads/cloud-config.yml" // Change as needed
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Build URL with base64-encoded payload as last path segment
	urlPath := "/api/v1.0/client/volume/cloudinit/metadata/" + encodedPayload
	url := c.endpoint + ":27777" + urlPath

	// Create HTTP request with file bytes as body
	req, err := http.NewRequest("POST", url, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 206:
		return "CONTINUE", nil
	case 200:
		return "OK", nil
	default:
		return "ERROR", nil
	}
}

// Get storages
func (c *Client) GetStorage() (*schema.GetStorageResponse, error) {
	var res schema.GetStorageResponse
	err := c.CallAPI("GET", "27777", "/api/v1.0/client/volume/accessible", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Get storage by ID
func (c *Client) GetStorageByID(reqBody schema.GetStorageByIDRequest) (*schema.GetStorageByIDResponse, error) {
	var res schema.GetStorageByIDResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/info", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete a storage volume
func (c *Client) DeleteStorage(reqBody interface{}) (*schema.DeleteStorageResponse, error) {
	var res schema.DeleteStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ------------------------------ MOCKED ENDPOINTS -----------------------------

// Get a Network by Id
func (c *Client) GetNetworkByID(reqBody interface{}) (*schema.NetworkGetResponse, error) {
	var res schema.NetworkGetResponse

	//! Mock data
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	res.Network = schema.Network{
		ID:      1,
		Name:    "test-network",
		Created: time.Now(),
		IPRange: ipnet,
		Subnets: []schema.NetworkSubnet{
			{
				Type:        "cloud",
				IPRange:     ipnet,
				NetworkZone: "eu-south",
				Gateway:     net.ParseIP("10.0.0.1"),
			},
		},
		Routes:     "",
		Servers:    []*schema.Server{},
		Protection: false,
		Labels: map[string]string{
			"environment": "test",
		},
	}

	return &res, nil
}

// List all networks
func (c *Client) ListNetwork(reqBody interface{}) (*schema.NetworkListResponse, error) {
	var res schema.NetworkListResponse

	//! Mock data
	_, ipnet1, _ := net.ParseCIDR("10.0.0.0/24")
	_, ipnet2, _ := net.ParseCIDR("192.168.0.0/24")
	res.Networks = []schema.Network{
		{
			ID:      1,
			Name:    "test.k8s",
			Created: time.Now(),
			IPRange: ipnet1,
			Subnets: []schema.NetworkSubnet{
				{
					Type:        "cloud",
					IPRange:     ipnet1,
					NetworkZone: "eu-south-1",
					Gateway:     net.ParseIP("10.0.0.1"),
				},
			},
			Routes:     "",
			Servers:    []*schema.Server{},
			Protection: false,
			Labels: map[string]string{
				"environment": "test",
			},
		},
		{
			ID:      2,
			Name:    "test-network-2",
			Created: time.Now(),
			IPRange: ipnet2,
			Subnets: []schema.NetworkSubnet{
				{
					Type:        "cloud",
					IPRange:     ipnet2,
					NetworkZone: "eu-south",
					Gateway:     net.ParseIP("192.168.0.1"),
				},
			},
			Routes:     "",
			Servers:    []*schema.Server{},
			Protection: false,
			Labels: map[string]string{
				"environment": "prod",
			},
		},
	}

	return &res, nil
}

// Delete a network
func (c *Client) DeleteNetwork(network interface{}) (*Response, error) {
	var res Response

	//! Mock data - just return an empty response with 200 status
	res.Response = &http.Response{
		StatusCode: http.StatusOK,
	}

	return &res, nil
}

// Create a network
func (c *Client) CreateNetwork(network interface{}) (*schema.NetworkCreateResponse, error) {
	var res schema.NetworkCreateResponse

	//! Mock data
	req, ok := network.(schema.NetworkCreateRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	_, ipnet, _ := net.ParseCIDR(req.IPRange)
	res.Network = schema.Network{
		ID:         3, // Increment from the last network ID
		Name:       req.Name,
		Created:    time.Now(),
		IPRange:    ipnet,
		Subnets:    req.Subnets,
		Routes:     "",
		Servers:    []*schema.Server{},
		Protection: false,
		Labels:     *req.Labels,
	}

	return &res, nil
}

// SSH KEYS...

// ------------------------------ UTILS FUNCTIONS -----------------------------

// Base function to perform API calls
func (c *Client) CallAPI(method, port, path string, reqBody, resType interface{}, needAuth bool) error {
	req, err := c.NewRequest(method, port, path, reqBody, needAuth)
	if err != nil {
		return err
	}
	response, err := c.Do(req)
	if err != nil {
		return err
	}
	return c.UnmarshalResponse(response, resType)
}

// NewRequest returns a new HTTP request
func (c *Client) NewRequest(method, port, path string, reqBody interface{}, needAuth bool) (*http.Request, error) {
	var body []byte
	var err error

	if reqBody != nil {
		body, err = json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Request body for %s %s:\n%s\n", method, path, string(body)) // TEST
	}

	target := c.endpoint + ":" + port + path
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Add("Accept", "application/json")

	if needAuth {
		// TODO: add auth
	}

	c.httpClient.Timeout = c.timeout
	return req, nil
}

// Do sends an HTTP request and returns an HTTP response
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.logger != nil {
		c.logger.LogRequest(req)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if c.logger != nil {
		c.logger.LogResponse(resp)
	}
	return resp, nil
}

// UnmarshalResponse checks the response and unmarshals it into the response type if needed
func (c *Client) UnmarshalResponse(response *http.Response, resType interface{}) error {
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		apiErr := &APIError{
			StatusCode: response.StatusCode,
			Message:    string(body),
		}
		return apiErr
	}

	if len(body) == 0 || resType == nil {
		return nil
	}

	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	return d.Decode(resType)
}

// ------------------------------ ERROR HANDLING -----------------------------

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %d - %s", e.StatusCode, e.Message)
}
