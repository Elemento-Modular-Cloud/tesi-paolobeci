package ecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// NodeupClient is a client for the nodeup API.
type NodeupClient struct {
	client *Client
}

type VM struct {
	Role string `json:"role"`
	UUID string `json:"uuid"`
}

type ExecuteClusterStartupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (nc *NodeupClient) ExecuteClusterStartup(ctx context.Context, clusterName string) (*ExecuteClusterStartupResponse, error) {
	// Get compute instances
	computeResponse, err := nc.client.GetCompute()
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

	err = nc.executeSSHCommand("51.159.157.254", "root", ansibleCommand, 5*time.Minute)
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
	err = nc.executeSSHCommand("51.159.157.254", "root", ansibleK8sCommand, 30*time.Minute)
	if err != nil {
		return &ExecuteClusterStartupResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to execute SSH command cluster startup: %v", err),
		}, nil
	} else {
		fmt.Println("Successfully executed cluster startup command via SSH.")
	}

	// After successful cluster startup, retrieve and process kubeconfig
	kubeconfigResponse, err := nc.GetClusterKubeconfig("51.159.157.254") // TODO: Make dynamic
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

func (c *NodeupClient) GetClusterKubeconfig(serverIP string) (*GetClusterKubeconfigResponse, error) {
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

// -------------------------- UTILS FUNCTIONS --------------------------

// executeSSHCommand executes a command on a remote server via SSH with specified timeout
func (c *NodeupClient) executeSSHCommand(host, user, command string, timeout time.Duration) error {
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
func (c *NodeupClient) executeSSHCommandWithOutput(host, user, command string, timeout time.Duration) (string, error) {
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

func (c *NodeupClient) loadSSHKey(keyPath string) (ssh.AuthMethod, error) {
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
