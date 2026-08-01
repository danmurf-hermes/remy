package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DefaultProvider != "ollama" {
		t.Errorf("expected default provider 'ollama', got %q", cfg.DefaultProvider)
	}
	if cfg.Memory.WorkingMemoryTurns != 20 {
		t.Errorf("expected 20 working memory turns, got %d", cfg.Memory.WorkingMemoryTurns)
	}
	if _, ok := cfg.Providers["ollama"]; !ok {
		t.Error("expected ollama provider in defaults")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig on missing file should return defaults, got error: %v", err)
	}
	if cfg.DefaultProvider != "ollama" {
		t.Errorf("expected default provider, got %q", cfg.DefaultProvider)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.DefaultProvider != "test-provider" {
		t.Errorf("expected 'test-provider', got %q", cfg.DefaultProvider)
	}
	if cfg.Memory.WorkingMemoryTurns != 10 {
		t.Errorf("expected 10, got %d", cfg.Memory.WorkingMemoryTurns)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{invalid json}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := DefaultConfig()
	cfg.DefaultProvider = "saved-provider"

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.DefaultProvider != "saved-provider" {
		t.Errorf("expected 'saved-provider', got %q", loaded.DefaultProvider)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".remy", "config.json")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
