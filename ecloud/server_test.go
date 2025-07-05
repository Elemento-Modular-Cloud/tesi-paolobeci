package ecloud

import (
	"testing"
)

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
