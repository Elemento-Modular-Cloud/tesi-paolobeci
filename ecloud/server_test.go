package ecloud

import (
	"context"
	"testing"
	"time"
	
	// "github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Log("=== Starting comprehensive Server Create test ===")

	// Step 1: Create client
	t.Log("Step 1: Creating client...")
	client, err := NewClient("test-app", "1.0.0")
	if err != nil {
		t.Logf("❌ Failed to create client: %v", err)
		t.SkipNow()
	}
	t.Logf("✓ Client created successfully: %T", client)

	// Step 2: Create ServerClient
	t.Log("Step 2: Creating ServerClient...")
	serverClient := &ServerClient{client: client}
	t.Logf("✓ ServerClient created: %+v", serverClient)

	// Step 3: Define test cases for different server configurations
	testCases := []struct {
		name        string
		opts        ServerCreateOpts
		description string
		expectError bool
	}{
		{
			name: "basic_ubuntu_server",
			opts: ServerCreateOpts{
				Name: "test-server-ubuntu",
				ServerType: &ServerType{
					Name:         "neon",
					Cores:        2,
					Memory:       2.0,
					Disk:         50,
					Architecture: ArchitectureX86_64,
				},
				Image: "ubuntu-24-04",
				SSHKeys: []*SSHKey{
					{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test-key",
					},
				},
				Datacenter: &Datacenter{
					ID:   1,
					Name: "Test Datacenter",
				},
				UserData:         "#!/bin/bash\necho 'Hello World' > /tmp/test.txt",
				StartAfterCreate: &[]bool{true}[0],
				Labels: map[string]string{
					"test":        "true",
					"environment": "testing",
				},
			},
			description: "Basic Ubuntu server with SSH key and user data",
			expectError: false,
		},
	}

	// Step 4: Execute test cases
	for _, tc := range testCases {
		t.Logf("\n--- Testing: %s ---", tc.description)
		t.Logf("Test case name: %s", tc.name)
		t.Logf("Expected error: %v", tc.expectError)

		// Log the input options
		t.Logf("Input ServerCreateOpts:")
		t.Logf("  Name: %q", tc.opts.Name)
		if tc.opts.ServerType != nil {
			t.Logf("  ServerType:")
			t.Logf("    Name: %q", tc.opts.ServerType.Name)
			t.Logf("    Cores: %d", tc.opts.ServerType.Cores)
			t.Logf("    Memory: %.1f GB", tc.opts.ServerType.Memory)
			t.Logf("    Disk: %d GB", tc.opts.ServerType.Disk)
			t.Logf("    Architecture: %s", tc.opts.ServerType.Architecture)
		} else {
			t.Logf("  ServerType: nil")
		}
		t.Logf("  Image: %q", tc.opts.Image)
		t.Logf("  SSH Keys count: %d", len(tc.opts.SSHKeys))
		for i, key := range tc.opts.SSHKeys {
			t.Logf("    SSH Key %d: %s...", i+1, key.PublicKey[:min(50, len(key.PublicKey))])
		}
		if tc.opts.Datacenter != nil {
			t.Logf("  Datacenter:")
			t.Logf("    ID: %d", tc.opts.Datacenter.ID)
			t.Logf("    Name: %q", tc.opts.Datacenter.Name)
		} else {
			t.Logf("  Datacenter: nil")
		}
		t.Logf("  UserData length: %d characters", len(tc.opts.UserData))
		if tc.opts.StartAfterCreate != nil {
			t.Logf("  StartAfterCreate: %v", *tc.opts.StartAfterCreate)
		} else {
			t.Logf("  StartAfterCreate: nil")
		}
		t.Logf("  Labels count: %d", len(tc.opts.Labels))
		for k, v := range tc.opts.Labels {
			t.Logf("    %s: %q", k, v)
		}

		// Call the Create function
		t.Logf("Calling ServerClient.Create()...")
		result, response, err := serverClient.Create(ctx, tc.opts)

		// Handle expected errors
		if tc.expectError {
			if err != nil {
				t.Logf("✓ Expected error occurred: %v", err)
				t.Logf("  Error type: %T", err)
			} else {
				t.Logf("❌ Expected error but Create succeeded")
				t.Logf("  Result: %+v", result)
			}
			continue
		}

		// Handle unexpected errors
		if err != nil {
			t.Logf("❌ Create returned unexpected error: %v", err)
			t.Logf("  Error type: %T", err)
			continue
		}

		// Success case - validate results
		t.Logf("✓ Create succeeded!")
		t.Logf("Result validation:")

		// Validate ServerCreateResult
		if result.Server != nil {
			t.Logf("  ✓ Server created:")
			t.Logf("    ID: %q", result.Server.ID)
			t.Logf("    Name: %q", result.Server.Name)
			t.Logf("    Status: %s", result.Server.Status)
			t.Logf("    Created: %s", result.Server.Created.Format(time.RFC3339))
			t.Logf("    Labels count: %d", len(result.Server.Labels))
			t.Logf("    Volumes count: %d", len(result.Server.Volumes))
		} else {
			t.Logf("  ❌ Server is nil in result")
		}

		if result.RootPassword != "" {
			t.Logf("  ✓ Root password provided (length: %d)", len(result.RootPassword))
		} else {
			t.Logf("  ⚠️  No root password in result")
		}

		// Validate Response
		if response != nil {
			t.Logf("  ✓ Response object is non-nil: %+v", response)
		} else {
			t.Logf("  ❌ Response is nil")
		}

		// Additional validations
		t.Logf("Additional validations:")
		if result.Server != nil && result.Server.ID != "" {
			t.Logf("  ✓ Server ID is not empty")
		}
		if result.Server != nil && result.Server.Name == tc.opts.Name {
			t.Logf("  ✓ Server name matches input")
		}
		if result.Server != nil && len(result.Server.Labels) > 0 {
			t.Logf("  ✓ Server has labels")
		}

		t.Logf("--- Completed test: %s ---\n", tc.description)
	}

	t.Log("=== Comprehensive Server Create test completed ===")
}

// Helper function for Go versions that don't have min built-in
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestConvertServerSize(t *testing.T) {
	tests := []struct {
		name            string
		size            string
		expectedSlots   int
		expectedRamsize int
		expectError     bool
	}{
		{
			name:            "Helium size",
			size:            "helium",
			expectedSlots:   1,
			expectedRamsize: 512,
			expectError:     false,
		},
		{
			name:            "Neon size",
			size:            "neon",
			expectedSlots:   2,
			expectedRamsize: 2048,
			expectError:     false,
		},
		{
			name:            "Argon2 size",
			size:            "argon2",
			expectedSlots:   4,
			expectedRamsize: 4096,
			expectError:     false,
		},
		{
			name:            "Argon size",
			size:            "argon",
			expectedSlots:   6,
			expectedRamsize: 4096,
			expectError:     false,
		},
		{
			name:            "Kripton size",
			size:            "kripton",
			expectedSlots:   8,
			expectedRamsize: 8192,
			expectError:     false,
		},
		{
			name:            "Case insensitive - HELIUM",
			size:            "HELIUM",
			expectedSlots:   1,
			expectedRamsize: 512,
			expectError:     false,
		},
		{
			name:            "Case insensitive - Neon",
			size:            "Neon",
			expectedSlots:   2,
			expectedRamsize: 2048,
			expectError:     false,
		},
		{
			name:            "Invalid size",
			size:            "invalid",
			expectedSlots:   0,
			expectedRamsize: 0,
			expectError:     true,
		},
		{
			name:            "Empty size",
			size:            "",
			expectedSlots:   0,
			expectedRamsize: 0,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ConvertServerSize(tt.size)

			if tt.expectError {
				if err == nil {
					t.Errorf("ConvertServerSize() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertServerSize() unexpected error: %v", err)
				return
			}

			if config.Slots != tt.expectedSlots {
				t.Errorf("ConvertServerSize() slots = %v, want %v", config.Slots, tt.expectedSlots)
			}

			if config.Ramsize != tt.expectedRamsize {
				t.Errorf("ConvertServerSize() ramsize = %v, want %v", config.Ramsize, tt.expectedRamsize)
			}
		})
	}
}

func TestCreateBootVolume(t *testing.T) {
	ctx := context.Background()

	// Mock client
	mockClient, _ := NewClient("test", "1")

	// Call the function with a mock saveCloudInitToFile
	volumeIDs, err := createBootVolume(
		ctx,
		mockClient,
		"test-server",
		"ubuntu",
		[]string{"ssh-rsa AAA..."},
		"data provided by kops",
	)
	if err != nil {
		t.Fatalf("createBootVolume returned error: %v", err)
	}

	// Check results
	expected := []string{"boot-volume-id", "cloudinit-volume-id"}
	if len(volumeIDs) != len(expected) {
		t.Fatalf("expected %d volume IDs, got %d", len(expected), len(volumeIDs))
	}
	for i, id := range expected {
		if volumeIDs[i] != id {
			t.Errorf("expected volume ID %q, got %q", id, volumeIDs[i])
		}
	}
}
