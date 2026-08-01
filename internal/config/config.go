package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ProviderConfig struct {
	Endpoint       string         `json:"endpoint"`
	APIKey         string         `json:"api_key"`
	ChatModel      string         `json:"chat_model"`
	EmbeddingModel string         `json:"embedding_model"`
	Parameters     map[string]any `json:"parameters,omitempty"`
}

type MemoryConfig struct {
	DBPath                    string `json:"db_path"`
	WorkingMemoryTurns        int    `json:"working_memory_turns"`
	QuickConsolidationDelayMs int    `json:"quick_consolidation_delay_ms"`
	DeepConsolidationDelayMs  int    `json:"deep_consolidation_delay_ms"`
}

type PersonaConfig struct {
	Active    string `json:"active"`
	Directory string `json:"directory"`
}

type TelegramConfig struct {
	Enabled      bool     `json:"enabled"`
	BotToken     string   `json:"bot_token"`
	AllowedUsers []string `json:"allowed_users"`
}

type InterfacesConfig struct {
	Telegram TelegramConfig `json:"telegram"`
}

type Config struct {
	Providers       map[string]ProviderConfig `json:"providers"`
	DefaultProvider string                    `json:"default_provider"`
	Memory          MemoryConfig              `json:"memory"`
	Persona         PersonaConfig             `json:"persona"`
	Interfaces      InterfacesConfig          `json:"interfaces"`
}

const defaultProviderName = "ollama"

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
			QuickConsolidationDelayMs: 300000,
			DeepConsolidationDelayMs:  1800000,
		},
		Persona: PersonaConfig{
			Active:    "default",
			Directory: "~/.remy/personas/",
		},
		Interfaces: InterfacesConfig{
			Telegram: TelegramConfig{
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

func ConfigPath() (string, error) {
	dir, err := remyDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
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

func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
