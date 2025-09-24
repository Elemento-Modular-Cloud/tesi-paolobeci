package ecloud

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
	"golang.org/x/crypto/ssh"
)

//go:embed version.txt
var version string

// ------------------------------ API CALLS FUNCTIONS -------------------------

// Login to the API
func (c *Client) Login(reqBody *schema.LoginRequest) (*schema.LoginResponse, error) {
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
func (c *Client) CreateStorageCloudInit(reqBody schema.CreateStorageCloudInitRequest, userData string) (*schema.CreateStorageCloudInitResponse, error) {
	var res schema.CreateStorageCloudInitResponse

	fmt.Printf("Marshalling request body: %+v\n", reqBody)
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reqBody: %w", err)
	}
	encodedPayload := base64.StdEncoding.EncodeToString(jsonBytes)
	fmt.Printf("Encoded payload (base64): %s\n", encodedPayload)

	// -------------- CLOUD-INIT USER-DATA MODIFICATION --------------
	// Use the embedded user-data template from cloudconfig package
	modifiedContent := CloudinitTemplate
	if reqBody.Name != "" {
		hostname := strings.TrimSuffix(reqBody.Name, "-cloudinit")
		modifiedContent = strings.Replace(modifiedContent, "hostname: myhost", "hostname: "+hostname, 1)
		fmt.Printf("Replaced hostname with: %s\n", hostname)
	}

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field - using the modified content
	fileWriter, err := writer.CreateFormFile("file", "user-data")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Write the modified content instead of copying from file
	_, err = fileWriter.Write([]byte(modifiedContent))
	if err != nil {
		return nil, fmt.Errorf("failed to write modified content: %w", err)
	}

	writer.Close()
	// ------------- END cloud-init user-data modification -------------

	// Build URL for the request
	url := c.endpoint + ":27777/api/v1.0/client/volume/cloudinit/metadata/" + encodedPayload

	// Create request with multipart body
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the correct Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("Sending HTTP request to: %s\n", url)
	fmt.Printf("Content-Type: %s\n", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		fmt.Println("Closing response body.")
		resp.Body.Close()
	}()

	fmt.Printf("Received response with status code: %d\n", resp.StatusCode)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("API error response body: %s\n", string(body))
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	fmt.Println("Decoding response body into struct...")
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Decoded response: %+v\n", res)
	return &res, nil
}

func (c *Client) FeedFileIntoCloudInitStorage(reqBody schema.FeedFileIntoCloudInitStorageRequest) (string, error) {
	// Marshal reqBody to JSON and encode as base64
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal reqBody: %w", err)
	}
	encodedPayload := base64.StdEncoding.EncodeToString(jsonBytes)
	fmt.Printf("Encoded payload (base64): %s\n", encodedPayload)

	// Use the embedded meta-data template from cloudconfig package
	metaDataContent := MetaDataTemplate

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field - using just the filename, not the full path
	fileWriter, err := writer.CreateFormFile("file", "meta-data")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Write the meta-data content from string
	_, err = fileWriter.Write([]byte(metaDataContent))
	if err != nil {
		return "", fmt.Errorf("failed to write meta-data content: %w", err)
	}

	writer.Close()

	// Build URL with base64-encoded payload as last path segment
	url := c.endpoint + ":27777/api/v1.0/client/volume/cloudinit/metadata/" + encodedPayload

	// Create request with multipart body
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set the correct Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("Sending HTTP request to: %s\n", url)
	fmt.Printf("Content-Type: %s\n", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		fmt.Println("Closing response body.")
		resp.Body.Close()
	}()

	fmt.Printf("Received response with status code: %d\n", resp.StatusCode)
	switch resp.StatusCode {
	case 206:
		return "CONTINUE", nil
	case 200:
		return "OK", nil
	default:
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("API error response body: %s\n", string(body))
		return "", fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
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
func (c *Client) DeleteStorage(reqBody schema.DeleteStorageRequest) (*schema.DeleteStorageResponse, error) {
	var res schema.DeleteStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Get a Network by Id
func (c *Client) GetNetworkByID(reqBody schema.GetNetworkByIDRequest) (*schema.GetNetworkByIDResponse, error) {
	var res schema.GetNetworkByIDResponse

	err := c.CallAPI("POST", "37777", "/api/v1.0/client/network/info", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List all networks
func (c *Client) ListNetwork() (*schema.ListNetworkResponse, error) {
	var res schema.ListNetworkResponse

	err := c.CallAPI("GET", "37777", "/api/v1.0/client/network/list", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete a network
func (c *Client) DeleteNetwork(reqBody schema.DeleteNetworkRequest) (*schema.DeleteNetworkResponse, error) {
	var res schema.DeleteNetworkResponse

	err := c.CallAPI("DELETE", "37777", "/api/v1.0/client/network/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create a network
func (c *Client) CreateNetwork(network schema.CreateNetworkRequest) (*schema.CreateNetworkResponse, error) {
	var res schema.CreateNetworkResponse

	err := c.CallAPI("POST", "37777", "/api/v1.0/client/network/create", network, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ------------------------------ DEVELOPMENT ENDPOINTS -----------------------------

type VM struct {
	Role string `json:"role"`
	UUID string `json:"uuid"`
}

type ExecuteClusterStartupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *Client) ExecuteClusterStartup() (*ExecuteClusterStartupResponse, error) {
	// Get compute instances
	computeResponse, err := c.GetCompute()
	if err != nil {
		return nil, fmt.Errorf("failed to get compute instances: %w", err)
	}

	// Extract VMs with roles and UUIDs
	var vms []VM
	for _, server := range *computeResponse {
		var role string
		vmName := strings.ToLower(server.ReqJSON.VMName)

		if strings.Contains(vmName, "control-plane") {
			role = "control-plane"
		} else if strings.Contains(vmName, "node") {
			role = "nodes"
		} else {
			// Skip VMs that don't match our criteria
			continue
		}

		vm := VM{
			Role: role,
			UUID: server.UniqueID,
		}
		vms = append(vms, vm)
	}

	if len(vms) == 0 {
		return &ExecuteClusterStartupResponse{
			Success: false,
			Message: "No VMs found with control-plane or nodes in their names",
		}, nil
	}

	// Convert VMs list to JSON string for the Ansible command
	vmsJSON, err := json.Marshal(vms)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VMs list: %w", err)
	}

	// Execute SSH command
	ansibleCommand := fmt.Sprintf("ansible-playbook inventory_creation.yaml -e 'vms_list=%s'", string(vmsJSON))

	err = c.executeSSHCommand("51.159.157.254", "root", ansibleCommand, 5*time.Minute)
	if err != nil {
		return &ExecuteClusterStartupResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to execute SSH command inventory creation: %v", err),
		}, nil
	} else {
		fmt.Println("Successfully executed inventory creation command via SSH.")
	}

	// Use extended timeout for k8s setup (30 minutes) with SSH host key checking disabled
	ansibleK8sCommand := "ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory.ini k8s_up.yaml -e 'ansible_ssh_common_args=\"-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null\"'"
	err = c.executeSSHCommand("51.159.157.254", "root", ansibleK8sCommand, 30*time.Minute)
	if err != nil {
		return &ExecuteClusterStartupResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to execute SSH command cluster startup: %v", err),
		}, nil
	} else {
		fmt.Println("Successfully executed cluster startup command via SSH.")
	}

	// After successful cluster startup, retrieve and process kubeconfig
	kubeconfigResponse, err := c.GetClusterKubeconfig("51.159.157.254") // TODO: Make dynamic
	if err != nil {
		fmt.Printf("Failed to get kubeconfig from Elemento: %v\n", err)
	} else if !kubeconfigResponse.Success {
		fmt.Printf("Failed to get kubeconfig: %s\n", kubeconfigResponse.Message)
	} else {
		kubeconfigContent := kubeconfigResponse.Kubeconfig
		fmt.Printf("Successfully retrieved kubeconfig from Elemento: %d bytes\n", len(kubeconfigContent))

		// Replace certificate-authority-data with insecure-skip-tls-verify
		// Note: Using simple string replacement since regex package would need to be imported
		certAuthorityPattern := "certificate-authority-data:"
		insecureSkipTLS := "    insecure-skip-tls-verify: true"

		// Find and replace certificate-authority-data lines
		lines := strings.Split(kubeconfigContent, "\n")
		for i, line := range lines {
			if strings.Contains(line, certAuthorityPattern) {
				lines[i] = insecureSkipTLS
			}
		}

		// Extract the original server IP from the kubeconfig and replace with localhost
		serverPrefix := "server: https://"
		serverSuffix := ":6443"
		var controllerIP string

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, serverPrefix) && strings.HasSuffix(trimmedLine, serverSuffix) {
				// Extract IP between "https://" and ":6443"
				start := len(serverPrefix)
				end := len(trimmedLine) - len(serverSuffix)
				if end > start {
					controllerIP = trimmedLine[start:end]
					fmt.Printf("CONTROLLER SERVER IP: %s\n", controllerIP)
					fmt.Printf("CONTROLLER SERVER SSH TUNNEL: ssh -nNT -L 6443:%s:6443 root@51.159.157.254\n", controllerIP) // TODO: Make dynamic

					// Replace server IP with localhost
					lines[i] = strings.Replace(line, controllerIP, "localhost", 1)
				}
				break
			}
		}

		// Rebuild the modified kubeconfig
		modifiedKubeconfig := strings.Join(lines, "\n")

		// Create ~/.kube directory if it doesn't exist
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			fmt.Printf("Failed to get user home directory from HOME environment variable\n")
		} else {
			kubeDir := homeDir + "/.kube"
			if err := os.MkdirAll(kubeDir, 0755); err != nil {
				fmt.Printf("Failed to create .kube directory: %v\n", err)
			} else {
				// Write kubeconfig file
				kubeconfigPath := kubeDir + "/kubeconfig" // TODO: Make dynamic with cluster name
				if err := os.WriteFile(kubeconfigPath, []byte(modifiedKubeconfig), 0600); err != nil {
					fmt.Printf("Failed to write kubeconfig file: %v\n", err)
				} else {
					fmt.Printf("Successfully wrote kubeconfig to %s\n", kubeconfigPath)
				}
			}
		}
	}

	return &ExecuteClusterStartupResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully executed cluster startup for %d VMs and saved kubeconfig to ~/.kube/kubeconfig", len(vms)),
	}, nil
}

type GetClusterKubeconfigResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Kubeconfig string `json:"kubeconfig,omitempty"`
}

func (c *Client) GetClusterKubeconfig(serverIP string) (*GetClusterKubeconfigResponse, error) {
	// Command to read the kubeconfig file from the remote server
	command := "cat kubeconfig"

	fmt.Printf("Attempting to read kubeconfig file from remote server\n")

	// Execute SSH command to read the kubeconfig file
	output, err := c.executeSSHCommandWithOutput(serverIP, "root", command, 30*time.Second)
	if err != nil {
		return &GetClusterKubeconfigResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to execute SSH command to get kubeconfig: %v", err),
		}, nil
	}

	if strings.TrimSpace(output) == "" {
		return &GetClusterKubeconfigResponse{
			Success: false,
			Message: "Kubeconfig file is empty or not found",
		}, nil
	}

	fmt.Println("Successfully retrieved kubeconfig from remote server")

	return &GetClusterKubeconfigResponse{
		Success:    true,
		Message:    "Successfully retrieved kubeconfig",
		Kubeconfig: output,
	}, nil
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

// executeSSHCommand executes a command on a remote server via SSH with specified timeout
func (c *Client) executeSSHCommand(host, user, command string, timeout time.Duration) error {
	fmt.Printf("Starting SSH connection to %s@%s (timeout: %v)\n", user, host, timeout)

	// Try to load SSH key for authentication
	var authMethods []ssh.AuthMethod

	// Try to load private key from standard locations
	keyPaths := []string{
		os.ExpandEnv("$HOME/.ssh/id_rsa"),
		os.ExpandEnv("$HOME/.ssh/id_ed25519"),
	}

	for _, keyPath := range keyPaths {
		fmt.Printf("Attempting to load SSH key from: %s\n", keyPath)
		if auth, err := c.loadSSHKey(keyPath); err == nil {
			authMethods = append(authMethods, auth)
			fmt.Printf("Successfully loaded SSH key from: %s\n", keyPath)
			break
		} else {
			fmt.Printf("Failed to load SSH key from %s: %v\n", keyPath, err)
		}
	}

	// If no key could be loaded, return error
	if len(authMethods) == 0 {
		return fmt.Errorf("no SSH authentication methods available - could not load SSH keys from standard locations")
	}

	// Create SSH client configuration with timeout
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: This is insecure for production
		Timeout:         30 * time.Second,            // Add timeout
	}

	fmt.Printf("Attempting to connect to SSH server %s:22\n", host)

	// Add a timeout context for the connection attempt
	connResult := make(chan struct {
		client *ssh.Client
		err    error
	}, 1)

	go func() {
		client, err := ssh.Dial("tcp", host+":22", config)
		connResult <- struct {
			client *ssh.Client
			err    error
		}{client: client, err: err}
	}()

	// Wait for connection or timeout
	var client *ssh.Client
	select {
	case result := <-connResult:
		client = result.client
		if result.err != nil {
			return fmt.Errorf("failed to connect to SSH server %s:22: %w", host, result.err)
		}
	case <-time.After(45 * time.Second): // 45 second timeout for connection
		return fmt.Errorf("SSH connection to %s:22 timed out after 45 seconds", host)
	}
	defer func() {
		fmt.Printf("Closing SSH connection\n")
		client.Close()
	}()

	fmt.Printf("SSH connection established successfully\n")

	// Create a session
	fmt.Printf("Creating SSH session\n")
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() {
		fmt.Printf("Closing SSH session\n")
		session.Close()
	}()

	fmt.Printf("SSH session created successfully\n")

	// Execute the command
	fmt.Printf("Executing SSH command: %s\n", command)
	if timeout > 10*time.Minute {
		fmt.Printf("This is a long-running command (timeout: %v). This may take some time...\n", timeout)
	}

	// Use a channel to handle timeout for command execution
	type result struct {
		output []byte
		err    error
	}

	resultChan := make(chan result, 1)

	go func() {
		output, err := session.CombinedOutput(command)
		resultChan <- result{output: output, err: err}
	}()

	// Progress indicator for long-running commands
	var ticker *time.Ticker
	if timeout > 10*time.Minute {
		ticker = time.NewTicker(60 * time.Second) // Print progress every minute
		defer ticker.Stop()
	}

	startTime := time.Now()

	for {
		select {
		case res := <-resultChan:
			duration := time.Since(startTime)
			if res.err != nil {
				fmt.Printf("SSH command failed after %v: %v\n", duration.Truncate(time.Second), res.err)
				return fmt.Errorf("command execution failed: %w, output: %s", res.err, string(res.output))
			}
			fmt.Printf("SSH command completed successfully after %v\n", duration.Truncate(time.Second))
			fmt.Printf("SSH command output: %s\n", string(res.output))
			return nil
		case <-func() <-chan time.Time {
			if ticker != nil {
				return ticker.C
			}
			return make(chan time.Time) // Never triggers if no ticker
		}():
			elapsed := time.Since(startTime)
			remaining := timeout - elapsed
			fmt.Printf("Command still running... Elapsed: %v, Remaining: %v\n",
				elapsed.Truncate(time.Second), remaining.Truncate(time.Second))
		case <-time.After(timeout):
			elapsed := time.Since(startTime)
			fmt.Printf("SSH command execution timed out after %v\n", elapsed.Truncate(time.Second))
			return fmt.Errorf("SSH command execution timed out after %v", timeout)
		}
	}
}

// executeSSHCommandWithOutput executes a command on a remote server via SSH and returns the output
func (c *Client) executeSSHCommandWithOutput(host, user, command string, timeout time.Duration) (string, error) {
	fmt.Printf("Starting SSH connection to %s@%s (timeout: %v)\n", user, host, timeout)

	// Try to load SSH key for authentication
	var authMethods []ssh.AuthMethod

	// Try to load private key from standard locations
	keyPaths := []string{
		os.ExpandEnv("$HOME/.ssh/id_rsa"),
		os.ExpandEnv("$HOME/.ssh/id_ed25519"),
	}

	for _, keyPath := range keyPaths {
		fmt.Printf("Attempting to load SSH key from: %s\n", keyPath)
		if auth, err := c.loadSSHKey(keyPath); err == nil {
			authMethods = append(authMethods, auth)
			fmt.Printf("Successfully loaded SSH key from: %s\n", keyPath)
			break
		} else {
			fmt.Printf("Failed to load SSH key from %s: %v\n", keyPath, err)
		}
	}

	// If no key could be loaded, return error
	if len(authMethods) == 0 {
		return "", fmt.Errorf("no SSH authentication methods available - could not load SSH keys from standard locations")
	}

	// Create SSH client configuration with timeout
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: This is insecure for production
		Timeout:         30 * time.Second,            // Add timeout
	}

	fmt.Printf("Attempting to connect to SSH server %s:22\n", host)

	// Add a timeout context for the connection attempt
	connResult := make(chan struct {
		client *ssh.Client
		err    error
	}, 1)

	go func() {
		client, err := ssh.Dial("tcp", host+":22", config)
		connResult <- struct {
			client *ssh.Client
			err    error
		}{client: client, err: err}
	}()

	// Wait for connection or timeout
	var client *ssh.Client
	select {
	case result := <-connResult:
		client = result.client
		if result.err != nil {
			return "", fmt.Errorf("failed to connect to SSH server %s:22: %w", host, result.err)
		}
	case <-time.After(45 * time.Second): // 45 second timeout for connection
		return "", fmt.Errorf("SSH connection to %s:22 timed out after 45 seconds", host)
	}
	defer func() {
		fmt.Printf("Closing SSH connection\n")
		client.Close()
	}()

	fmt.Printf("SSH connection established successfully\n")

	// Create a session
	fmt.Printf("Creating SSH session\n")
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() {
		fmt.Printf("Closing SSH session\n")
		session.Close()
	}()

	fmt.Printf("SSH session created successfully\n")

	// Execute the command and capture output
	fmt.Printf("Executing SSH command: %s\n", command)

	// Use a channel to handle timeout for command execution
	type result struct {
		output []byte
		err    error
	}

	resultChan := make(chan result, 1)

	go func() {
		output, err := session.CombinedOutput(command)
		resultChan <- result{output: output, err: err}
	}()

	startTime := time.Now()

	select {
	case res := <-resultChan:
		duration := time.Since(startTime)
		if res.err != nil {
			fmt.Printf("SSH command failed after %v: %v\n", duration.Truncate(time.Second), res.err)
			return "", fmt.Errorf("command execution failed: %w, output: %s", res.err, string(res.output))
		}
		fmt.Printf("SSH command completed successfully after %v\n", duration.Truncate(time.Second))
		fmt.Printf("SSH command output length: %d bytes\n", len(res.output))
		return string(res.output), nil
	case <-time.After(timeout):
		elapsed := time.Since(startTime)
		fmt.Printf("SSH command execution timed out after %v\n", elapsed.Truncate(time.Second))
		return "", fmt.Errorf("SSH command execution timed out after %v", timeout)
	}
}

func (c *Client) loadSSHKey(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	// Create the Signer for this private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// ------------------------------ ERROR HANDLING -----------------------------
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %d - %s", e.StatusCode, e.Message)
}
