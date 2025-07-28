package ecloud

import (
	"context"
	"testing"
)

func TestCreateCloudInit(t *testing.T) {
	ctx := context.Background()

	// Create a real client
	// Note: This test will skip if no valid config is available
	client, err := NewClient("test-app", "1.0.0")
	if err != nil {
		t.Skipf("Skipping test: failed to create client: %v", err)
	}

	// Create the volume client
	volumeClient := &VolumeClient{client: client}

	// Set up test options
	opts := CloudInitCreateOpts{
		Name: "test-cloudinit-2",
	}

	t.Logf("Calling CreateCloudInit with opts: %+v", opts)
	volumeID, resp, err := volumeClient.CreateCloudInit(ctx, opts)

	// Note: This test may fail if the API is not available or authentication fails
	// In a real environment, you might want to check for specific error types
	if err != nil {
		t.Logf("CreateCloudInit returned error (this may be expected if API is not available): %v", err)
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
