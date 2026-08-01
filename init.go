// Package main provides the init command for first-time Remy setup.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/persona"
)

// initCmd runs the first-time setup for Remy: checks dependencies,
// creates the ~/.remy/ directory structure, generates default config
// and persona, and prints next steps.
func initCmd() error {
	fmt.Println("🔧 Remy Initial Setup")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	// 1. Check Ollama is running
	fmt.Print("🔍 Checking Ollama... ")
	ollamaOK, ollamaEndpoint := checkOllama()
	if !ollamaOK {
		fmt.Println("❌ NOT FOUND")
		fmt.Println()
		fmt.Println("Ollama is not running. Please start Ollama first:")
		fmt.Println("  ollama serve")
		fmt.Println()
		fmt.Println("Or configure a different provider in ~/.remy/config.json")
		fmt.Println("after running this command.")
		fmt.Println()
	} else {
		fmt.Println("✅ Found at", ollamaEndpoint)
	}

	// 2. Check required models
	if ollamaOK {
		fmt.Print("🔍 Checking chat model (llama3.1:8b)... ")
		if checkModel(ollamaEndpoint, "llama3.1:8b") {
			fmt.Println("✅ Found")
		} else {
			fmt.Println("❌ NOT FOUND")
			fmt.Println("  Run: ollama pull llama3.1:8b")
		}

		fmt.Print("🔍 Checking embedding model (nomic-embed-text)... ")
		if checkModel(ollamaEndpoint, "nomic-embed-text") {
			fmt.Println("✅ Found")
		} else {
			fmt.Println("❌ NOT FOUND")
			fmt.Println("  Run: ollama pull nomic-embed-text")
		}
	}

	fmt.Println()

	// 3. Create ~/.remy/ directory structure
	fmt.Print("📁 Creating ~/.remy/ directory... ")
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}
	remyDir := filepath.Join(home, ".remy")
	personaDir := filepath.Join(remyDir, "personas")

	if err := os.MkdirAll(personaDir, 0o750); err != nil {
		return fmt.Errorf("creating ~/.remy/personas/: %w", err)
	}
	fmt.Println("✅")

	// 4. Generate default config.json
	cfgPath := filepath.Join(remyDir, "config.json")
	if _, err := os.Stat(cfgPath); err == nil {
		fmt.Println("📄 config.json already exists — skipping")
	} else {
		fmt.Print("📄 Generating default config.json... ")
		cfg := config.DefaultConfig()
		if err := config.SaveConfig(cfgPath, cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Println("✅")
	}

	// 5. Generate default persona
	defaultPersonaPath := filepath.Join(personaDir, "default.md")
	if _, err := os.Stat(defaultPersonaPath); err == nil {
		fmt.Println("📄 personas/default.md already exists — skipping")
	} else {
		fmt.Print("📄 Generating default persona... ")
		defaultPersona := &persona.Persona{
			Name: "default",
			Body: `You are Remy, a helpful, knowledgeable, and friendly personal AI assistant.

You have access to long-term memory (facts about the user, past conversations),
a scratchpad for notes, and a scheduler for reminders and recurring tasks.

Be concise but thorough. Adapt your tone to match the user's communication style.
When you don't know something, say so rather than making things up.

You can:
- Answer questions and have conversations
- Remember facts about the user and their preferences
- Set reminders and manage schedules
- Search through past conversations and memories
- Help with creative and analytical tasks

Always be respectful, honest, and helpful.`,
		}
		if err := persona.SavePersona(defaultPersonaPath, defaultPersona); err != nil {
			return fmt.Errorf("saving default persona: %w", err)
		}
		fmt.Println("✅")
	}

	fmt.Println()
	fmt.Println("✅ Setup complete!")
	fmt.Println()
	fmt.Println("📋 Next steps:")
	fmt.Println()
	fmt.Println("  1. Start Remy:")
	fmt.Println("     ./remy")
	fmt.Println()
	fmt.Println("  2. (Optional) Connect Telegram:")
	fmt.Println("     - Set REMY_TELEGRAM_BOT_TOKEN environment variable")
	fmt.Println("     - Set interfaces.telegram.enabled to true in config.json")
	fmt.Println("     - Run: ./remy --daemon")
	fmt.Println()
	fmt.Println("  3. Customize your experience:")
	fmt.Println("     - Edit ~/.remy/config.json to change providers, models, etc.")
	fmt.Println("     - Edit ~/.remy/personas/default.md to change Remy's personality")
	fmt.Println("     - Add new persona files to ~/.remy/personas/")
	fmt.Println()

	return nil
}

// checkOllama checks if Ollama is running at the default endpoint
// or any common alternative. Returns true and the endpoint URL if found.
func checkOllama() (found bool, endpoint string) {
	endpoints := []string{
		"http://localhost:11434/v1",
		"http://localhost:11434",
		"http://127.0.0.1:11434/v1",
		"http://127.0.0.1:11434",
	}

	client := &http.Client{Timeout: 2 * time.Second}

	for _, ep := range endpoints {
		// Build the models URL based on whether the endpoint already has /v1
		var modelsURL string
		if strings.HasSuffix(ep, "/v1") {
			modelsURL = ep + "/models"
		} else {
			modelsURL = ep + "/v1/models"
		}

		req, err := http.NewRequest("GET", modelsURL, http.NoBody)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Return the base endpoint (with /v1)
			baseEndpoint := ep
			if !strings.HasSuffix(ep, "/v1") {
				baseEndpoint = ep + "/v1"
			}
			return true, baseEndpoint
		}
	}

	return false, ""
}

// checkModel checks if a specific model is available in Ollama.
func checkModel(endpoint, model string) bool {
	// Try the OpenAI-compatible /v1/models endpoint
	modelsURL := strings.TrimSuffix(endpoint, "/v1") + "/v1/models"
	client := &http.Client{Timeout: 2 * time.Second}

	req, err := http.NewRequest("GET", modelsURL, http.NoBody)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	// Parse the response to check for the model
	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return false
	}

	for _, m := range modelsResp.Data {
		if m.ID == model {
			return true
		}
	}

	return false
}

// firstRunCheck checks if Remy has been initialized and returns
// a descriptive message if not. Returns empty string if everything is fine.
func firstRunCheck() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	remyDir := filepath.Join(home, ".remy")
	cfgPath := filepath.Join(remyDir, "config.json")

	if _, err := os.Stat(remyDir); os.IsNotExist(err) {
		return fmt.Sprintf(
			"Remy has not been initialized yet.\n\n"+
				"Run 'remy init' to set up your configuration:\n"+
				"  %s/remy init\n\n"+
				"This will:\n"+
				"  - Check that Ollama is running\n"+
				"  - Check for required AI models\n"+
				"  - Create ~/.remy/ with default config and persona\n",
			filepath.Dir(remyDir))
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return fmt.Sprintf(
			"Configuration file not found at %s.\n\n"+
				"Run 'remy init' to generate a default configuration:\n"+
				"  %s/remy init\n",
			cfgPath, filepath.Dir(remyDir))
	}

	return ""
}
