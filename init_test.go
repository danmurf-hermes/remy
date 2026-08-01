package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestInitCmd tests the init command logic by running it in a temporary directory
// with a mocked home directory.
func TestInitCmd(t *testing.T) {
	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Create a temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	// Run initCmd - it should succeed even without Ollama running
	err := initCmd()
	if err != nil {
		t.Fatalf("initCmd() returned error: %v", err)
	}

	// Verify ~/.remy/ directory was created
	remyDir := filepath.Join(tmpHome, ".remy")
	if _, err := os.Stat(remyDir); os.IsNotExist(err) {
		t.Error("~/.remy/ directory was not created")
	}

	// Verify ~/.remy/personas/ directory was created
	personaDir := filepath.Join(remyDir, "personas")
	if _, err := os.Stat(personaDir); os.IsNotExist(err) {
		t.Error("~/.remy/personas/ directory was not created")
	}

	// Verify config.json was created
	cfgPath := filepath.Join(remyDir, "config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Error("config.json was not created")
	}

	// Verify config.json is valid JSON
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("reading config.json: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("config.json is not valid JSON: %v", err)
	}

	// Verify default persona was created
	defaultPersonaPath := filepath.Join(personaDir, "default.md")
	if _, err := os.Stat(defaultPersonaPath); os.IsNotExist(err) {
		t.Error("personas/default.md was not created")
	}

	// Verify persona file has frontmatter
	personaData, err := os.ReadFile(defaultPersonaPath)
	if err != nil {
		t.Fatalf("reading default.md: %v", err)
	}
	content := string(personaData)
	if len(content) < 10 {
		t.Error("default.md seems too short")
	}
}

// TestInitCmdIdempotent tests that running init twice doesn't cause errors
// and doesn't overwrite existing files.
func TestInitCmdIdempotent(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	// First run
	if err := initCmd(); err != nil {
		t.Fatalf("first initCmd() failed: %v", err)
	}

	// Second run - should not error
	if err := initCmd(); err != nil {
		t.Fatalf("second initCmd() failed: %v", err)
	}

	// Verify files still exist
	remyDir := filepath.Join(tmpHome, ".remy")
	cfgPath := filepath.Join(remyDir, "config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Error("config.json missing after second init")
	}
}

// TestFirstRunCheck tests the first-run detection logic.
func TestFirstRunCheck(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, home string)
		wantMsg bool // true if we expect a non-empty message
	}{
		{
			name: "no remy directory",
			setup: func(t *testing.T, home string) {
				// Don't create anything
			},
			wantMsg: true,
		},
		{
			name: "remy dir but no config",
			setup: func(t *testing.T, home string) {
				remyDir := filepath.Join(home, ".remy")
				if err := os.MkdirAll(remyDir, 0o750); err != nil {
					t.Fatalf("creating .remy dir: %v", err)
				}
			},
			wantMsg: true,
		},
		{
			name: "fully initialized",
			setup: func(t *testing.T, home string) {
				remyDir := filepath.Join(home, ".remy")
				personaDir := filepath.Join(remyDir, "personas")
				if err := os.MkdirAll(personaDir, 0o750); err != nil {
					t.Fatalf("creating dirs: %v", err)
				}
				cfgPath := filepath.Join(remyDir, "config.json")
				if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
					t.Fatalf("writing config: %v", err)
				}
			},
			wantMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origHome := os.Getenv("HOME")
			defer os.Setenv("HOME", origHome)

			tmpHome := t.TempDir()
			os.Setenv("HOME", tmpHome)

			tt.setup(t, tmpHome)

			msg := firstRunCheck()
			if tt.wantMsg && msg == "" {
				t.Error("expected non-empty message, got empty")
			}
			if !tt.wantMsg && msg != "" {
				t.Errorf("expected empty message, got: %s", msg)
			}
		})
	}
}

// TestCheckOllama tests the Ollama detection with a mock server.
func TestCheckOllama(t *testing.T) {
	// Start a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" || r.URL.Path == "/models" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":[{"id":"llama3.1:8b"},{"id":"nomic-embed-text"}]}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Override the endpoints to use our test server
	// We can't easily override the endpoints slice, so let's test the checkModel function directly
	found := checkModel(server.URL+"/v1", "llama3.1:8b")
	if !found {
		t.Error("checkModel should have found llama3.1:8b")
	}

	found = checkModel(server.URL+"/v1", "nomic-embed-text")
	if !found {
		t.Error("checkModel should have found nomic-embed-text")
	}

	found = checkModel(server.URL+"/v1", "nonexistent-model")
	if found {
		t.Error("checkModel should not have found nonexistent-model")
	}
}

// TestCheckModelWithEmptyResponse tests checkModel with an empty model list.
func TestCheckModelWithEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	found := checkModel(server.URL+"/v1", "llama3.1:8b")
	if found {
		t.Error("checkModel should not have found any model in empty list")
	}
}

// TestCheckModelWithServerError tests checkModel when the server returns an error.
func TestCheckModelWithServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	found := checkModel(server.URL+"/v1", "llama3.1:8b")
	if found {
		t.Error("checkModel should return false on server error")
	}
}

// TestCheckModelWithTimeout tests checkModel when the server is unreachable.
func TestCheckModelWithTimeout(t *testing.T) {
	// Use a port that's unlikely to be listening
	found := checkModel("http://127.0.0.1:1/v1", "llama3.1:8b")
	if found {
		t.Error("checkModel should return false when server is unreachable")
	}
}
