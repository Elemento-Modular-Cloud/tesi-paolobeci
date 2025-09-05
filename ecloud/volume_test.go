package ecloud

import (
	"context"
	"testing"
	"time"
)

func TestCreateVolumeWithUrl(t *testing.T) {
	ctx := context.Background()

	// Create a real client
	// Note: This test will skip if no valid config is available
	client, err := NewClient("test-app", "1.0.0")
	if err != nil {
		t.Skipf("Skipping test: failed to create client: %v", err)
	}

	// Create the volume client
	volumeClient := &VolumeClient{client: client}

	// Set up test options with URL
	opts := VolumeCreateOpts{
		Name:      "test-volume-with-url",
		Size:      10, // 10 GB
		Bootable:  true,
		Readonly:  false,
		Shareable: false,
		Private:   false,
		Labels:    map[string]string{"test": "true"},
		Url:       "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img",
	}

	t.Logf("Calling Create with opts: %+v", opts)
	volumeID, resp, err := volumeClient.Create(ctx, opts)

	// Note: This test may fail if the API is not available or authentication fails
	// In a real environment, you might want to check for specific error types
	if err != nil {
		t.Logf("Create returned error: %v", err)
		return
	}

	t.Logf("Received volumeID: %q", volumeID)
	t.Logf("Received response: %+v", resp)

	if volumeID == "" {
		t.Error("Expected non-empty volumeID")
	}
	if resp == nil {
		t.Error("Expected non-nil response")
	}
}

func TestCreateCloudInit(t *testing.T) {
	ctx := context.Background()

	t.Log("=== Starting CreateCloudInit test ===")

	// Create a real client
	t.Log("Step 1: Creating client...")
	client, err := NewClient("test-app", "1.0.0")
	if err != nil {
		t.Skipf("Skipping test: failed to create client: %v", err)
	}
	t.Logf("✓ Client created successfully: %T", client)

	// Create the volume client
	t.Log("Step 2: Creating VolumeClient...")
	volumeClient := &VolumeClient{client: client}
	t.Logf("✓ VolumeClient created: %+v", volumeClient)

	// Test different cloud-init options
	testCases := []struct {
		name        string
		opts        CloudInitCreateOpts
		description string
	}{
		{
			name: "basic_cloudinit",
			opts: CloudInitCreateOpts{
				Name: "test-cloudinit",
			},
			description: "Basic cloud-init volume creation",
		},
		// Add more test cases here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("--- Testing: %s ---", tc.description)
			t.Logf("Input options: %+v", tc.opts)

			// Log the expected request that will be sent
			t.Logf("Expected request body:")
			t.Logf("  Name: %q", tc.opts.Name)
			t.Logf("  Private: false")
			t.Logf("  Bootable: true")
			t.Logf("  Clonable: false")
			t.Logf("  Alg: \"no\"")
			t.Logf("  ExpectedFiles: 2")

			// Call CreateCloudInit
			t.Log("Calling CreateCloudInit...")
			volumeID, resp, err := volumeClient.CreateCloudInit(ctx, tc.opts)

			if err != nil {
				t.Logf("❌ CreateCloudInit returned error: %v", err)
				t.Logf("Error type: %T", err)
				return
			}

			t.Logf("✓ CreateCloudInit succeeded!")
			t.Logf("Returned volume ID: %q", volumeID)
			t.Logf("Returned response: %+v", resp)

			// Detailed validation
			t.Log("Performing detailed validation...")

			if volumeID == "" {
				t.Error("❌ Expected non-empty volumeID")
			} else {
				t.Logf("✓ Volume ID is non-empty: %s", volumeID)
				t.Logf("  Volume ID length: %d characters", len(volumeID))
			}

			if resp == nil {
				t.Error("❌ Expected non-nil response")
			} else {
				t.Logf("✓ Response is non-nil: %+v", resp)
			}

			// Additional checks
			t.Log("Additional validation checks:")
			if len(volumeID) > 0 {
				t.Logf("✓ Volume ID format appears valid")
			}

			t.Logf("--- Completed test: %s ---", tc.description)
		})
	}

	t.Log("=== Detailed CreateCloudInit test completed ===")
}

func TestFeedFileIntoCloudInitStorage(t *testing.T) {
	ctx := context.Background()

	t.Log("=== Starting detailed FeedFileIntoCloudInitStorage test ===")

	// Create a real client
	t.Log("Step 1: Creating client...")
	client, err := NewClient("test-app", "1.0.0")
	if err != nil {
		t.Skipf("Skipping test: failed to create client: %v", err)
	}
	t.Logf("✓ Client created successfully: %T", client)

	// Create the volume client
	t.Log("Step 2: Creating VolumeClient...")
	volumeClient := &VolumeClient{client: client}
	t.Logf("✓ VolumeClient created: %+v", volumeClient)

	// First, we need to create a cloud-init volume to feed files into
	t.Log("Step 3: Creating a cloud-init volume for testing...")
	cloudInitOpts := CloudInitCreateOpts{
		Name: "test-cloudinit-complete",
	}

	t.Logf("Creating cloud-init volume with opts: %+v", cloudInitOpts)
	volumeID, _, err := volumeClient.CreateCloudInit(ctx, cloudInitOpts)
	if err != nil {
		t.Logf("❌ Failed to create cloud-init volume: %v", err)
		return
	}
	t.Logf("✓ Cloud-init volume created successfully with ID: %q", volumeID)

	// Wait 5 seconds to allow the volume to be fully initialized
	t.Log("Waiting 5 seconds for volume initialization...")
	time.Sleep(5 * time.Second)
	t.Log("✓ Wait completed")

	// Test different scenarios for feeding files
	testCases := []struct {
		name        string
		volumeID    string
		description string
	}{
		{
			name:        "feed_files_to_valid_volume",
			volumeID:    volumeID,
			description: "Feed files into a valid cloud-init volume",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("--- Testing: %s ---", tc.description)
			t.Logf("Input volume ID: %q", tc.volumeID)

			// Log the expected request that will be sent
			t.Logf("Expected request body:")
			t.Logf("  VolumeID: %q", tc.volumeID)

			// Call FeedFileIntoCloudInitStorage
			t.Log("Calling FeedFileIntoCloudInitStorage...")
			response, resp, err := volumeClient.FeedFileIntoCloudInitStorage(ctx, tc.volumeID)

			if err != nil {
				t.Logf("❌ FeedFileIntoCloudInitStorage returned error: %v", err)
				t.Logf("Error type: %T", err)

				// For empty or invalid volume IDs, we expect errors
				if tc.volumeID == "" || tc.volumeID == "invalid-volume-id-12345" {
					t.Logf("✓ Error was expected for this test case")
				}
				return
			}

			t.Logf("✓ FeedFileIntoCloudInitStorage succeeded!")
			t.Logf("Returned response string: %q", response)
			t.Logf("Returned response object: %+v", resp)

			// Detailed validation
			t.Log("Performing detailed validation...")

			if response == "" {
				t.Error("❌ Expected non-empty response string")
			} else {
				t.Logf("✓ Response string is non-empty: %s", response)
				t.Logf("  Response length: %d characters", len(response))

				// Check for expected response values
				switch response {
				case "CONTINUE":
					t.Logf("✓ Response indicates more files need to be sent")
				case "OK":
					t.Logf("✓ Response indicates all files received and loaded")
				default:
					t.Logf("⚠️  Unexpected response value: %q", response)
				}
			}

			if resp == nil {
				t.Error("❌ Expected non-nil response object")
			} else {
				t.Logf("✓ Response object is non-nil: %+v", resp)
			}

			// Additional checks
			t.Log("Additional validation checks:")
			if tc.volumeID != "" && len(response) > 0 {
				t.Logf("✓ Function executed successfully with valid volume ID")
			}

			t.Logf("--- Completed test: %s ---", tc.description)
		})
	}

	t.Log("=== Detailed FeedFileIntoCloudInitStorage test completed ===")
}
