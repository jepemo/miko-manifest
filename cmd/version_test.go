package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		commit   string
		date     string
		expected []string
	}{
		{
			name:     "default values",
			version:  "dev",
			commit:   "none",
			date:     "unknown",
			expected: []string{"miko-manifest version dev", "commit: none", "built: unknown"},
		},
		{
			name:     "real build values",
			version:  "v1.0.0",
			commit:   "a1b2c3d",
			date:     "2025-08-27T10:00:00Z",
			expected: []string{"miko-manifest version v1.0.0", "commit: a1b2c3d", "built: 2025-08-27T10:00:00Z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersion := version
			origCommit := commit
			origDate := date

			// Set test values
			version = tt.version
			commit = tt.commit
			date = tt.date

			// Restore original values after test
			defer func() {
				version = origVersion
				commit = origCommit
				date = origDate
			}()

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the version command function directly
			versionCmd.Run(versionCmd, []string{})

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Check output
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestVersionCommandExists(t *testing.T) {
	// Test that the version command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Version command not found in root command")
	}
}

func TestVersionCommandStructure(t *testing.T) {
	// Test command properties
	if versionCmd.Use != "version" {
		t.Errorf("Expected Use to be 'version', got '%s'", versionCmd.Use)
	}

	if versionCmd.Short != "Show version information" {
		t.Errorf("Expected Short to be 'Show version information', got '%s'", versionCmd.Short)
	}

	expectedLong := "Display the version, commit hash, and build date of miko-manifest."
	if versionCmd.Long != expectedLong {
		t.Errorf("Expected Long to be '%s', got '%s'", expectedLong, versionCmd.Long)
	}
}
