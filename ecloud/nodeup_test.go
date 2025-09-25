package ecloud

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// TestSSHFunctions tests SSH connection and command execution functionality
func TestSSHFunctions(t *testing.T) {
	fmt.Printf("\n%s\nSSH Function Tests\n%s\n",
		separator,
		subSeparator)

	// Create context for the test
	ctx := context.Background()

	// Initialize client and nodeup client
	baseClient, err := NewClient("SSH-TESTER", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create NodeupClient which has the SSH methods
	nodeupClient := &NodeupClient{client: baseClient}

	testEndpoint(t, "Load SSH Key", func() error {
		keyPath := "/Users/paolob/Developer/elemento/tesi-paolobeci/ecloud/ssh_keys/paolo.beci@gmail.com_private.txt"
		_, err := nodeupClient.loadSSHKey(keyPath)
		if err != nil {
			fmt.Printf("SSH Key load error: %v\n", err)
			return err
		}
		fmt.Printf("✅ SSH key loaded successfully\n")
		return nil
	})

	testEndpoint(t, "Basic SSH Connection Test", func() error {
		// Test basic SSH connection without executing commands
		keyPath := "/Users/paolob/Developer/elemento/tesi-paolobeci/ecloud/ssh_keys/paolo.beci@gmail.com_private.txt"
		authMethod, err := nodeupClient.loadSSHKey(keyPath)
		if err != nil {
			return fmt.Errorf("failed to load SSH key: %v", err)
		}

		config := &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         30 * time.Second,
		}

		host := "51.159.157.254"
		fmt.Printf("Testing SSH connection to %s:22...\n", host)

		conn, err := ssh.Dial("tcp", host+":22", config)
		if err != nil {
			fmt.Printf("❌ SSH connection failed: %v\n", err)
			return err
		}
		defer conn.Close()

		fmt.Printf("✅ SSH connection successful\n")
		return nil
	})

	testEndpoint(t, "Execute SSH Command - whoami", func() error {
		err := nodeupClient.executeSSHCommand("51.159.157.254", "root", "whoami", 30*time.Second)
		if err != nil {
			fmt.Printf("SSH command execution failed: %v\n", err)
			return err
		}
		fmt.Printf("✅ Command executed successfully\n")
		return nil
	})

	testEndpoint(t, "Execute SSH Command - hostname", func() error {
		err := nodeupClient.executeSSHCommand("51.159.157.254", "root", "hostname", 30*time.Second)
		if err != nil {
			fmt.Printf("SSH command execution failed: %v\n", err)
			return err
		}
		fmt.Printf("✅ Command executed successfully\n")
		return nil
	})

	testEndpoint(t, "Execute SSH Command - ls /", func() error {
		err := nodeupClient.executeSSHCommand("51.159.157.254", "root", "ls /", 30*time.Second)
		if err != nil {
			fmt.Printf("SSH command execution failed: %v\n", err)
			return err
		}
		fmt.Printf("✅ Command executed successfully\n")
		return nil
	})

	testEndpoint(t, "Test SSH Connection", func() error {
		err := nodeupClient.executeSSHCommand("51.159.157.254", "root", "echo 'SSH connection test successful'", 30*time.Second)
		if err != nil {
			fmt.Printf("SSH connection test failed: %v\n", err)
			return err
		}
		fmt.Printf("✅ SSH connection test passed\n")
		return nil
	})

	testEndpoint(t, "Test Long-Running SSH Command", func() error {
		// Test the long-running command functionality with a simple command first
		err := nodeupClient.executeSSHCommand("51.159.157.254", "root", "sleep 10 && echo 'Long running command test completed'", 2*time.Minute)
		if err != nil {
			fmt.Printf("Long-running SSH command test failed: %v\n", err)
			return err
		}
		fmt.Printf("✅ Long-running SSH command test passed\n")
		return nil
	})

	testEndpoint(t, "Execute Cluster Startup", func() error {
		resp, err := nodeupClient.ExecuteClusterStartup(ctx, "test-cluster")
		if err != nil {
			fmt.Printf("Cluster startup failed: %v\n", err)
			return err
		}
		fmt.Printf("Cluster startup response: %s\n", prettyPrint(resp))
		return nil
	})
}
