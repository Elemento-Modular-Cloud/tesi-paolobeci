package ecloud

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

const (
	separator    = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	subSeparator = "────────────────────────────────────────"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error pretty printing: %v", err)
	}
	return string(b)
}

func testEndpoint(t *testing.T, name string, fn func() error) {
	fmt.Printf("\n%s\n%s\n%s\n",
		separator,
		name,
		subSeparator)

	err := fn()
	if err != nil {
		// Format error message for better readability
		errMsg := err.Error()
		if strings.Contains(errMsg, "<!doctype html>") {
			errMsg = "Server returned HTML error page"
		}
		fmt.Printf("❌ Error: %s\n", errMsg)
		t.Error(errMsg)
	} else {
		fmt.Printf("✅ Success\n")
	}
}

func TestAPIEndpoints(t *testing.T) {
	// Print test configuration
	fmt.Printf("\n%s\nAPI Test Configuration\n%s\n",
		separator,
		subSeparator)

	// Check environment variables
	username := os.Getenv("ECL_USERNAME")
	password := os.Getenv("ECL_PASSWORD")
	if username == "" || password == "" {
		fmt.Printf("⚠️  Warning: ECL_USERNAME or ECL_PASSWORD not set\n")
	}

	// Initialize client
	client, err := NewClient(
		"API-TESTER",
		"1.0.0",
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Authentication Tests
	fmt.Printf("\n%s\nAuthentication Tests\n%s\n",
		separator,
		subSeparator)

	// Test login
	loginResp, err := client.Login(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
	} else {
		fmt.Printf("\nLogin Response:\n%s\n", prettyPrint(loginResp))
	}

	testEndpoint(t, "Status Login", func() error {
		resp, err := client.StatusLogin()
		if err != nil {
			return err
		}
		fmt.Printf("\nStatus Login Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	// Compute Tests
	fmt.Printf("\n%s\nCompute Tests\n%s\n",
		separator,
		subSeparator)

	testEndpoint(t, "Health Check Compute", func() error {
		resp, err := client.HealthCheckCompute()
		if err != nil {
			return err
		}
		fmt.Printf("\nHealth Check Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Can Allocate Compute", func() error {
		resp, err := client.CanAllocateCompute()
		if err != nil {
			return err
		}
		fmt.Printf("\nCan Allocate Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Create Compute", func() error {
		req := schema.CreateComputeRequest{
			Name:          "test-vm",
			Slots:         2,
			Overprovision: false,
			AllowSMT:      false,
			Archs:         []string{"x86_64"},
			Flags:         []string{},
			Ramsize:       2048,
			ReqECC:        false,
			Misc:          []string{},
			Pci:           []string{},
			Volumes:       []string{},
			Netdevs:       []string{},
		}
		resp, err := client.CreateCompute(req)
		if err != nil {
			return err
		}
		fmt.Printf("\nCreate Compute Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Compute Status", func() error {
		resp, err := client.ComputeStatus()
		if err != nil {
			return err
		}
		fmt.Printf("\nCompute Status Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Compute Templates", func() error {
		resp, err := client.ComputeTemplates()
		if err != nil {
			return err
		}
		fmt.Printf("\nCompute Templates Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Delete Compute", func() error {
		req := schema.DeleteComputeRequest{
			Name: "test-vm",
		}
		resp, err := client.DeleteCompute(req)
		if err != nil {
			return err
		}
		fmt.Printf("\nDelete Compute Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	// Storage Tests
	fmt.Printf("\n%s\nStorage Tests\n%s\n",
		separator,
		subSeparator)

	testEndpoint(t, "Health Check Storage", func() error {
		resp, err := client.HealthCheckStorage()
		if err != nil {
			return err
		}
		fmt.Printf("\nHealth Check Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Can Create Storage", func() error {
		resp, err := client.CanCreateStorage()
		if err != nil {
			return err
		}
		fmt.Printf("\nCan Create Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Create Storage", func() error {
		req := map[string]interface{}{
			"name": "test-volume",
			"size": 1024,
		}
		resp, err := client.CreateStorage(req)
		if err != nil {
			return err
		}
		fmt.Printf("\nCreate Storage Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Get Storage", func() error {
		resp, err := client.GetStorage()
		if err != nil {
			return err
		}
		fmt.Printf("\nGet Storage Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Get Storage By ID", func() error {
		req := map[string]string{
			"volumeID": "test-volume-id",
		}
		resp, err := client.GetStorageByID(req)
		if err != nil {
			return err
		}
		fmt.Printf("\nGet Storage By ID Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Delete Storage", func() error {
		req := map[string]string{
			"volumeID": "test-volume-id",
		}
		resp, err := client.DeleteStorage(req)
		if err != nil {
			return err
		}
		fmt.Printf("\nDelete Storage Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	// Cleanup Tests
	fmt.Printf("\n%s\nCleanup Tests\n%s\n",
		separator,
		subSeparator)

	testEndpoint(t, "Logout", func() error {
		resp, err := client.Logout()
		if err != nil {
			return err
		}
		fmt.Printf("\nLogout Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	// Test Summary
	fmt.Printf("\n%s\nTest Summary\n%s\n",
		separator,
		subSeparator)
	fmt.Println("All tests completed. Check the output above for details.")
}
