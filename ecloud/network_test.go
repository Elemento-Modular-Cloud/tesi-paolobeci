package ecloud

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

func TestNetworkFunctionalities(t *testing.T) {
	ctx := context.Background()

	// Create a real client
	client, err := NewClient("test-network-app", "1.0.0")
	if err != nil {
		t.Errorf("Skipping test: failed to create client: %v", err)
	}

	// Login
	body := &schema.LoginRequest{
		Username: os.Getenv("ECL_USERNAME"),
		Password: os.Getenv("ECL_PASSWORD"),
	}
	_, err = client.Login(body)
	if err != nil {
		t.Errorf("Skipping test: failed to login: %v", err)
		return
	}

	// Create the network client
	networkClient := &NetworkClient{client: client}

	// Parse IP range for testing
	_, ipRange, err := net.ParseCIDR("192.168.80.1/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	// Test variables
	testNetworkName := "test-network-integration"
	var createdNetwork *Network
	var networkID string

	t.Run("1_CreateNetwork", func(t *testing.T) {
		t.Logf("Testing network creation...")

		opts := NetworkCreateOpts{
			Name:    testNetworkName,
			IPRange: ipRange,
			Labels: map[string]string{
				"environment": "test",
				"purpose":     "integration-test",
			},
		}

		t.Logf("Creating network with options: %+v", opts)
		_, _, err := networkClient.Create(ctx, opts)

		if err != nil {
			t.Errorf("Create network returned error (may be expected if API is not available): %v", err)
			t.Skip("Skipping remaining tests due to API unavailability")
			return
		}

		t.Logf("Create OK")
	})

	t.Run("2_ListNetworks", func(t *testing.T) {
		if t.Failed() {
			t.Skip("Skipping due to previous test failure")
		}

		t.Logf("Testing network listing...")

		// Test listing all networks
		networks, _, err := networkClient.List(ctx, NetworkListOpts{})
		if err != nil {
			t.Errorf("List networks returned error: %v", err)
			return
		}

		t.Logf("Listed %d networks", len(networks))
		// t.Logf("List response: %+v", resp)

		for i, network := range networks {
			t.Logf("Network %d: ID=%s, Name=%s, IPRange=%v", i+1, network.ID, network.Name, network.IPRange)
		}

		// Test listing with name filter
		t.Logf("Testing network listing with name filter...")
		filteredNetworks, _, err := networkClient.List(ctx, NetworkListOpts{Name: testNetworkName})
		if err != nil {
			t.Errorf("List networks with filter returned error: %v", err)
			return
		}

		t.Logf("Found %d networks with name '%s'", len(filteredNetworks), testNetworkName)

		if len(filteredNetworks) > 0 {
			// Update our network info if we found it
			createdNetwork = filteredNetworks[0]
			networkID = createdNetwork.ID
			t.Logf("Found created network: ID=%s, Name=%s", networkID, createdNetwork.Name)
		}
	})

	t.Run("3_GetNetworkByName", func(t *testing.T) {
		if t.Failed() {
			t.Skip("Skipping due to previous test failure")
		}

		t.Logf("Testing get network by name...")

		network, _, err := networkClient.GetByName(ctx, testNetworkName)
		if err != nil {
			t.Errorf("GetByName returned error: %v", err)
			return
		}

		// t.Logf("GetByName response: %+v", resp)

		if network != nil {
			t.Logf("Found network by name: ID=%s, Name=%s, IPRange=%v", network.ID, network.Name, network.IPRange)
			t.Logf("Network details: Created=%v, Protection=%v, Routes=%s", network.Created, network.Protection, network.Routes)

			// Update our references
			createdNetwork = network
			networkID = network.ID
		} else {
			t.Errorf("Network '%s' not found", testNetworkName)
		}

		// Test with empty name
		emptyNetwork, _, err := networkClient.GetByName(ctx, "")
		if err != nil {
			t.Logf("GetByName with empty name returned error: %v", err)
		} else if emptyNetwork != nil {
			t.Error("Expected nil network for empty name")
		} else {
			t.Logf("GetByName with empty name correctly returned nil")
		}
	})

	t.Run("4_GetNetworkByID", func(t *testing.T) {
		if t.Failed() || networkID == "" {
			t.Skip("Skipping due to previous test failure or missing network ID")
		}

		t.Logf("Testing get network by ID: %s", networkID)

		network, resp, err := networkClient.GetByID(ctx, networkID)
		if err != nil {
			t.Errorf("GetByID returned error: %v", err)
			return
		}

		t.Logf("GetByID response: %+v", resp)

		if network != nil {
			t.Logf("Found network by ID: ID=%s, Name=%s, IPRange=%v", network.ID, network.Name, network.IPRange)
		} else {
			t.Errorf("Network not found by ID")
		}
	})

	t.Run("5_DeleteNetwork", func(t *testing.T) {
		if t.Failed() || createdNetwork == nil {
			t.Skip("Skipping due to previous test failure or missing network")
		}

		t.Logf("Testing network deletion for network: %s", createdNetwork.Name)

		_, deleteResp, err := networkClient.Delete(ctx, networkID)
		if err != nil {
			t.Errorf("Delete network returned error: %v", err)
			return
		}

		t.Logf("Delete API response: %+v", deleteResp)
		t.Logf("Network '%s' deletion request completed", createdNetwork.Name)

		// Wait a moment for deletion to process
		time.Sleep(2 * time.Second)

		// Verify deletion by trying to find the network again
		t.Logf("Verifying network deletion...")
		network, _, err := networkClient.GetByName(ctx, testNetworkName)
		if err != nil {
			t.Logf("Verification GetByName returned error: %v", err)
		} else if network == nil {
			t.Log("Network successfully deleted - not found in subsequent search")
		} else {
			t.Errorf("Network still exists after deletion: %+v", network)
		}
	})
}
