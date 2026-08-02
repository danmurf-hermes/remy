// Package config provides configuration loading, saving, and defaults
// for the Remy personal assistant. Configuration is stored as JSON
// in ~/.remy/config.json.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProviderConfig holds the connection and model settings for an LLM provider
// (e.g., Ollama, OpenAI). Each named provider in the config has one of these.
type ProviderConfig struct {
	Endpoint       string         `json:"endpoint"`
	APIKey         string         `json:"api_key"`
	ChatModel      string         `json:"chat_model"`
	EmbeddingModel string         `json:"embedding_model"`
	Parameters     map[string]any `json:"parameters,omitempty"`
}

// MemoryConfig controls the SQLite-backed memory system — database path,
// working memory size, and consolidation timing.
type MemoryConfig struct {
	DBPath                    string `json:"db_path"`
	WorkingMemoryTurns        int    `json:"working_memory_turns"`
	QuickConsolidationDelayMs int    `json:"quick_consolidation_delay_ms"`
	DeepConsolidationDelayMs  int    `json:"deep_consolidation_delay_ms"`
}

// PersonaConfig specifies the active persona and the directory where
// persona Markdown files are stored.
type PersonaConfig struct {
	Active    string `json:"active"`
	Directory string `json:"directory"`
}

// TelegramConfig controls the optional Telegram bot interface.
type TelegramConfig struct {
	Enabled      bool     `json:"enabled"`
	BotToken     string   `json:"bot_token"`
	AllowedUsers []string `json:"allowed_users"`
}

// InterfacesConfig holds configuration for all user-facing interfaces
// (e.g., Telegram).
type InterfacesConfig struct {
	Telegram TelegramConfig `json:"telegram"`
}

// Config is the top-level configuration for Remy. It includes provider
// definitions, memory settings, persona selection, and interface options.
type Config struct {
	Providers       map[string]ProviderConfig `json:"providers"`
	DefaultProvider string                    `json:"default_provider"`
	Memory          MemoryConfig              `json:"memory"`
	Persona         PersonaConfig             `json:"persona"`
	Interfaces      InterfacesConfig          `json:"interfaces"`
}

const defaultProviderName = "ollama"

// DefaultConfig returns a Config with sensible defaults for a local Ollama
// setup. The returned config can be modified and saved via SaveConfig.
func DefaultConfig() *Config {
	return &Config{
		Providers: map[string]ProviderConfig{
			defaultProviderName: {
				Endpoint:       "http://localhost:11434/v1",
				APIKey:         "",
				ChatModel:      "llama3.1:8b",
				EmbeddingModel: "nomic-embed-text",
				Parameters: map[string]any{
					"temperature": 0.7,
					"max_tokens":  4096,
				},
			},
		},
		DefaultProvider: defaultProviderName,
		Memory: MemoryConfig{
			DBPath:                    "~/.remy/memory.db",
			WorkingMemoryTurns:        20,
			QuickConsolidationDelayMs: 30000,  // 30 seconds
			DeepConsolidationDelayMs:  300000, // 5 minutes
		},
		Persona: PersonaConfig{
			Active:    "default",
			Directory: "~/.remy/personas/",
		},
		Interfaces: InterfacesConfig{
			Telegram: TelegramConfig{ //nolint:gosec // bot token is a placeholder env var ref
				Enabled:      false,
				BotToken:     "${REMY_TELEGRAM_BOT_TOKEN}",
				AllowedUsers: []string{},
			},
		},
	}
}

func remyDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".remy"), nil
}

// ConfigPath returns the expected path to the config file
// (~/.remy/config.json), creating the ~/.remy directory if needed.
func ConfigPath() (string, error) { //nolint:revive // stutters but is the clearest name
	dir, err := remyDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig reads and parses a JSON config file. If the file does not
// exist, it returns DefaultConfig without an error.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is user-provided config path
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// SaveConfig writes the given config as indented JSON to the specified path,
// creating parent directories as needed.
func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
