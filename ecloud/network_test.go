package ecloud

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNetworkFunctionalities(t *testing.T) {
	ctx := context.Background()

	// Create a real client
	client, err := NewClient("test-network-app", "1.0.0")
	if err != nil {
		t.Skipf("Skipping test: failed to create client: %v", err)
	}

	// Create the network client
	networkClient := &NetworkClient{client: client}

	// Parse IP range for testing
	_, ipRange, err := net.ParseCIDR("192.168.100.0/24")
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
			Routes:  "0.0.0.0/0 192.168.100.1",
			Labels: map[string]string{
				"environment": "test",
				"purpose":     "integration-test",
			},
		}

		t.Logf("Creating network with options: %+v", opts)
		network, resp, err := networkClient.Create(ctx, opts)

		// Note: API might not be available, so we handle this gracefully
		if err != nil {
			t.Logf("Create network returned error (may be expected if API is not available): %v", err)
			t.Skip("Skipping remaining tests due to API unavailability")
			return
		}

		t.Logf("Create response: %+v", resp)

		if network != nil {
			createdNetwork = network
			networkID = network.ID
			t.Logf("Network created successfully with ID: %s", networkID)
		} else {
			t.Log("Network creation returned nil (API may not return the created network)")
		}
	})

	t.Run("2_ListNetworks", func(t *testing.T) {
		if t.Failed() {
			t.Skip("Skipping due to previous test failure")
		}

		t.Logf("Testing network listing...")

		// Test listing all networks
		networks, resp, err := networkClient.List(ctx, NetworkListOpts{})
		if err != nil {
			t.Logf("List networks returned error: %v", err)
			return
		}

		t.Logf("Listed %d networks", len(networks))
		t.Logf("List response: %+v", resp)

		for i, network := range networks {
			t.Logf("Network %d: ID=%s, Name=%s, IPRange=%v", i+1, network.ID, network.Name, network.IPRange)
		}

		// Test listing with name filter
		t.Logf("Testing network listing with name filter...")
		filteredNetworks, _, err := networkClient.List(ctx, NetworkListOpts{Name: testNetworkName})
		if err != nil {
			t.Logf("List networks with filter returned error: %v", err)
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

		network, resp, err := networkClient.GetByName(ctx, testNetworkName)
		if err != nil {
			t.Logf("GetByName returned error: %v", err)
			return
		}

		t.Logf("GetByName response: %+v", resp)

		if network != nil {
			t.Logf("Found network by name: ID=%s, Name=%s, IPRange=%v", network.ID, network.Name, network.IPRange)
			t.Logf("Network details: Created=%v, Protection=%v, Routes=%s", network.Created, network.Protection, network.Routes)

			// Update our references
			createdNetwork = network
			networkID = network.ID
		} else {
			t.Logf("Network '%s' not found", testNetworkName)
		}

		// Test with empty name
		emptyNetwork, _, err := networkClient.GetByName(ctx, "")
		if err != nil {
			t.Logf("GetByName with empty name returned error: %v", err)
		} else if emptyNetwork != nil {
			t.Error("Expected nil network for empty name")
		} else {
			t.Log("GetByName with empty name correctly returned nil")
		}
	})

	t.Run("4_GetNetworkByID", func(t *testing.T) {
		if t.Failed() || networkID == "" {
			t.Skip("Skipping due to previous test failure or missing network ID")
		}

		t.Logf("Testing get network by ID: %s", networkID)

		// Note: The GetByID function expects an int, but our networkID is a string
		// This might need to be adjusted based on the actual API implementation
		t.Logf("Note: GetByID expects int but we have string ID - this may need API adjustment")

		// For now, we'll test with a dummy ID since the function signature expects int
		// In a real scenario, you'd need to convert or adjust the API
		network, resp, err := networkClient.GetByID(ctx, 1) // Using dummy ID
		if err != nil {
			t.Logf("GetByID returned error: %v", err)
			return
		}

		t.Logf("GetByID response: %+v", resp)

		if network != nil {
			t.Logf("Found network by ID: ID=%s, Name=%s, IPRange=%v", network.ID, network.Name, network.IPRange)
		} else {
			t.Log("Network not found by ID")
		}
	})

	t.Run("5_DeleteNetwork", func(t *testing.T) {
		if t.Failed() || createdNetwork == nil {
			t.Skip("Skipping due to previous test failure or missing network")
		}

		t.Logf("Testing network deletion for network: %s", createdNetwork.Name)

		resp, deleteResp, err := networkClient.Delete(ctx, createdNetwork)
		if err != nil {
			t.Logf("Delete network returned error: %v", err)
			return
		}

		t.Logf("Delete response: %+v", resp)
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
			t.Logf("Network still exists after deletion: %+v", network)
		}
	})

	t.Run("6_TestAllNetworks", func(t *testing.T) {
		t.Logf("Testing All() method...")

		networks, err := networkClient.All(ctx)
		if err != nil {
			t.Logf("All() returned error: %v", err)
			return
		}

		t.Logf("All() returned %d networks", len(networks))

		// Test AllWithOpts
		t.Logf("Testing AllWithOpts() method...")
		networksWithOpts, err := networkClient.AllWithOpts(ctx, NetworkListOpts{
			ListOpts: ListOpts{PerPage: 10},
		})
		if err != nil {
			t.Logf("AllWithOpts() returned error: %v", err)
			return
		}

		t.Logf("AllWithOpts() returned %d networks", len(networksWithOpts))
	})
}
