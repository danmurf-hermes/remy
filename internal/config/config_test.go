package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danmurf/remy/internal/config"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		check   func(t *testing.T, cfg *config.Config)
		wantErr bool
	}{
		{
			name: "default config",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.json")
			},
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.DefaultProvider != "ollama" {
					t.Errorf("expected default provider 'ollama', got %q", cfg.DefaultProvider)
				}
				if cfg.Memory.WorkingMemoryTurns != 20 {
					t.Errorf("expected 20 working memory turns, got %d", cfg.Memory.WorkingMemoryTurns)
				}
				if _, ok := cfg.Providers["ollama"]; !ok {
					t.Error("expected ollama provider in defaults")
				}
			},
		},
		{
			name: "file not found returns defaults",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.json")
			},
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.DefaultProvider != "ollama" {
					t.Errorf("expected default provider, got %q", cfg.DefaultProvider)
				}
			},
		},
		{
			name: "valid file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "config.json")
				content := `{
					"default_provider": "test-provider",
					"providers": {
						"test-provider": {
							"endpoint": "http://test:11434/v1",
							"api_key": "",
							"chat_model": "test-model",
							"embedding_model": "test-embed"
						}
					},
					"memory": {
						"db_path": "~/.remy/memory.db",
						"working_memory_turns": 10,
						"quick_consolidation_delay_ms": 60000,
						"deep_consolidation_delay_ms": 300000
					},
					"persona": {
						"active": "default",
						"directory": "~/.remy/personas/"
					},
					"interfaces": {
						"telegram": {
							"enabled": false,
							"bot_token": "",
							"allowed_users": []
						}
					}
				}`
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
				return path
			},
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.DefaultProvider != "test-provider" {
					t.Errorf("expected 'test-provider', got %q", cfg.DefaultProvider)
				}
				if cfg.Memory.WorkingMemoryTurns != 10 {
					t.Errorf("expected 10, got %d", cfg.Memory.WorkingMemoryTurns)
				}
			},
		},
		{
			name: "invalid JSON",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "config.json")
				if err := os.WriteFile(path, []byte(`{invalid json}`), 0o600); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "save and load round-trip",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "config.json")
				cfg := config.DefaultConfig()
				cfg.DefaultProvider = "saved-provider"
				if err := config.SaveConfig(path, cfg); err != nil {
					t.Fatalf("SaveConfig failed: %v", err)
				}
				return path
			},
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.DefaultProvider != "saved-provider" {
					t.Errorf("expected 'saved-provider', got %q", cfg.DefaultProvider)
				}
			},
		},
		{
			name: "config path",
			setup: func(_ *testing.T) string {
				return "" // not used
			},
			check: func(t *testing.T, _ *config.Config) {
				path, err := config.ConfigPath()
				if err != nil {
					t.Fatalf("ConfigPath failed: %v", err)
				}
				home, _ := os.UserHomeDir()
				expected := filepath.Join(home, ".remy", "config.json")
				if path != expected {
					t.Errorf("expected %q, got %q", expected, path)
				}
			},
		},
		{
			name: "save config creates directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "subdir", "config.json")
				cfg := config.DefaultConfig()
				if err := config.SaveConfig(path, cfg); err != nil {
					t.Fatalf("SaveConfig: %v", err)
				}
				return path
			},
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.DefaultProvider != "ollama" {
					t.Errorf("expected 'ollama', got %q", cfg.DefaultProvider)
				}
			},
		},
		{
			name: "load config read error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				// Use a directory as the path to trigger a read error
				return dir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			if tt.name == "config path" {
				tt.check(t, nil)
				return
			}
			cfg, err := config.LoadConfig(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}
