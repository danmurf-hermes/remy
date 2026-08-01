// Package main is the entry point for the Remy personal assistant binary.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/danmurf/remy/internal/agent"
	"github.com/danmurf/remy/internal/app"
	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/interface/telegram"
	"github.com/danmurf/remy/internal/llm"
	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/scheduler"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var version = "dev"

// wailsEmitter implements app.Emitter using the Wails runtime.
type wailsEmitter struct {
	ctx context.Context
}

func (e wailsEmitter) Emit(event string, data any) error {
	runtime.EventsEmit(e.ctx, event, data)
	return nil
}

func main() {
	daemonMode := flag.Bool("daemon", false, "Run in daemon mode (no GUI, Telegram only)")
	showHelp := flag.Bool("help", false, "Show this help message")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Remy — Personal AI Assistant

Usage:
  remy                    Start the desktop GUI
  remy init               Initialize Remy configuration
  remy --daemon           Run in daemon mode (Telegram + scheduler, no GUI)
  remy --help             Show this help message

Flags:
  --daemon    Run in daemon mode (no GUI, Telegram only)
  --help      Show this help message

Commands:
  init        Create ~/.remy/ directory, default config, and default persona

Documentation:
  https://github.com/danmurf/remy#readme
`)
	}
	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Handle "init" subcommand
	if flag.NArg() > 0 && flag.Arg(0) == "init" {
		runInit()
		return
	}

	cfgPath, err := config.ConfigPath()
	if err != nil {
		log.Fatalf("Error determining config path: %v", err)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	providerCfg, ok := cfg.Providers[cfg.DefaultProvider]
	if !ok {
		log.Fatalf("Default provider %q not found in config", cfg.DefaultProvider)
	}

	llmProvider, err := llm.NewProvider(providerCfg)
	if err != nil {
		log.Fatalf("Error creating LLM provider: %v", err)
	}

	dbPath := cfg.Memory.DBPath
	if dbPath != "" && dbPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}

	store, err := memory.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	embedder := memory.NewEmbedder(providerCfg.Endpoint, providerCfg.EmbeddingModel)

	personaLoader := app.NewPersonaLoader()

	sched := scheduler.NewScheduler(store, nil)

	agentCfg := &agent.Config{
		WorkingMemoryTurns:        cfg.Memory.WorkingMemoryTurns,
		UserID:                    "default-user",
		SessionID:                 "default",
		Interface:                 "gui",
		PersonaDir:                cfg.Persona.Directory,
		ActivePersona:             cfg.Persona.Active,
		QuickConsolidationDelayMs: cfg.Memory.QuickConsolidationDelayMs,
		DeepConsolidationDelayMs:  cfg.Memory.DeepConsolidationDelayMs,
	}

	ag := agent.NewAgent(store, llmProvider, embedder, personaLoader, sched, agentCfg)
	sched.SetAgent(ag)

	// Start Telegram interface if enabled
	var tgInterface *telegram.Interface
	if cfg.Interfaces.Telegram.Enabled {
		tgCfg := &cfg.Interfaces.Telegram
		tgInterface = telegram.New(ag, store, tgCfg)
		if err := tgInterface.Start(context.Background()); err != nil {
			log.Fatalf("Error starting Telegram interface: %v", err)
		}
		log.Printf("Telegram interface started (bot token: %s)", maskToken(tgCfg.BotToken))
	}

	sched.Start(context.Background())

	fmt.Printf("Remy %s starting...\n", version)

	if *daemonMode {
		// Daemon mode: run Telegram + scheduler only, block until signal
		log.Println("Running in daemon mode (no GUI)")
		select {}
	}

	application, err := app.NewApp(cfg, cfgPath, store, ag)
	if err != nil {
		_ = store.Close()
		log.Fatalf("Error creating app: %v", err)
	}

	err = wails.Run(&options.App{
		Title:  "Remy",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			application.SetEmitter(wailsEmitter{ctx: ctx})
			application.Startup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			application.Shutdown(ctx)
			if tgInterface != nil {
				tgInterface.Stop()
			}
		},
		Bind: []any{
			application,
		},
		Linux: &linux.Options{
			ProgramName: "Remy",
		},
	})

	_ = store.Close()

	if err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

// maskToken returns a masked version of a bot token for logging.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// runInit initializes the Remy configuration directory.
// This is a placeholder — the full implementation is in Stage 12.
func runInit() {
	fmt.Println("Remy init — setting up your configuration...")

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	remyDir := filepath.Join(home, ".remy")
	personaDir := filepath.Join(remyDir, "personas")

	// Create directories
	for _, dir := range []string{remyDir, personaDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Error creating directory %s: %v", dir, err)
		}
	}

	// Create default config if it doesn't exist
	cfgPath := filepath.Join(remyDir, "config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		defaultCfg := config.DefaultConfig()
		if err := config.SaveConfig(cfgPath, defaultCfg); err != nil {
			log.Fatalf("Error saving default config: %v", err)
		}
		fmt.Println("  ✓ Created default config:", cfgPath)
	} else {
		fmt.Println("  ✓ Config already exists:", cfgPath)
	}

	// Create default persona if it doesn't exist
	defaultPersona := filepath.Join(personaDir, "default.md")
	if _, err := os.Stat(defaultPersona); os.IsNotExist(err) {
		content := `---
name: default
---

You are Remy, a personal AI assistant. You are helpful, warm, and conversational. You remember past conversations and use that context to provide better responses. You can help with questions, tasks, reminders, and general conversation.
`
		if err := os.WriteFile(defaultPersona, []byte(content), 0644); err != nil {
			log.Fatalf("Error creating default persona: %v", err)
		}
		fmt.Println("  ✓ Created default persona:", defaultPersona)
	} else {
		fmt.Println("  ✓ Default persona already exists:", defaultPersona)
	}

	fmt.Println()
	fmt.Println("Remy is ready! Next steps:")
	fmt.Println("  1. Make sure Ollama is running with a chat model (e.g., llama3.1:8b)")
	fmt.Println("  2. Make sure nomic-embed-text is installed for memory support")
	fmt.Println("  3. Run 'remy' to start the desktop GUI")
	fmt.Println("  4. Run 'remy --daemon' for Telegram-only mode")
}
