package ecloud

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

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
		fmt.Printf("❌ Error: \n")
		t.Errorf("Test '%s' failed: %v", name, err)
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
	client.timeout = 120 * time.Second

	var serverID string

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

	// testEndpoint(t, "Health Check Compute", func() error {
	// 	resp, err := client.HealthCheckCompute()
	// 	if err != nil || *resp != "This is an Elemento Matcher Client!" {
	// 		return err
	// 	}
	// 	fmt.Printf("\nHealth Check Response:\n%s\n", prettyPrint(resp))
	// 	return nil
	// })

	testEndpoint(t, "Can Allocate Compute", func() error {
		req := schema.CanAllocateComputeRequest{
			Slots:         2,
			Overprovision: 2,
			AllowSMT:      false,
			Archs:         []string{"X86_64"},
			Flags:         []string{"sse2"},
			Ramsize:       2048,
			ReqECC:        false,
			Misc:          schema.Misc{OsFamily: "linux", OsFlavour: "pop"},
			Pci:           []string{},
		}
		resp, err := client.CanAllocateCompute(req)
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
			Overprovision: 2,
			AllowSMT:      false,
			Archs:         []string{"X86_64"},
			Flags:         []string{"sse2"},
			Ramsize:       2048,
			ReqECC:        false,
			Misc:          schema.Misc{OsFamily: "linux", OsFlavour: "ubuntu"},
			Pci:           []string{},
			Volumes:       []map[string]string{}, // {"vid": "volume_id"}
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
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			resp, err := client.ComputeStatus()
			if err != nil {
				return err
			}
			fmt.Printf("\nCompute Status Response (attempt %d/%d):\n%s\n", i+1, maxRetries, prettyPrint(resp))

			// Save the created server's uniqueID if available
			if len(*resp) > 0 {
				serverID = (*resp)[0].UniqueID
				fmt.Printf("\nCreated server ID: %s\n", serverID)
				return nil
			}

			if i < maxRetries-1 {
				fmt.Printf("No server found, retrying in 1 second...\n")
				time.Sleep(time.Second)
			}
		}
		return fmt.Errorf("no server found after %d attempts", maxRetries)
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
			LocalIndex: serverID,
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

	// testEndpoint(t, "Health Check Storage", func() error {
	// 	resp, err := client.HealthCheckStorage()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Printf("\nHealth Check Response:\n%s\n", prettyPrint(resp))
	// 	return nil
	// })

	testEndpoint(t, "Can Create Storage", func() error {
		reqBody := map[string]interface{}{
			"size": 100,
		}
		resp, err := client.CanCreateStorage(reqBody)
		if err != nil {
			return err
		}
		fmt.Printf("\nCan Create Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Create Storage", func() error {
		req := map[string]interface{}{
			"size":      100,
			"name":      "test-volume",
			"bootable":  true,
			"readonly":  false,
			"shareable": false,
			"private":   true,
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
			"volumeID": "d596ec1f15f7444b93e294c3cdbc1905", // TODO: cambia con quello del create
		}
		resp, err := client.GetStorageByID(req)
		if err != nil {
			return err
		}
		fmt.Printf("\nGet Storage By ID Response:\n%s\n", prettyPrint(resp))
		return nil
	})

	testEndpoint(t, "Delete Storage", func() error {
		req := schema.DeleteStorageRequest{
			VolumeID: "ffffffff-fffff-ffff-ffff-ffffffffffff", // TODO: make this value the one of the create
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
