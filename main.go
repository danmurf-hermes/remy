// Package main is the entry point for the Remy personal assistant binary.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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
	// Handle the "init" subcommand before flag parsing
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := initCmd(); err != nil {
			slog.Error("Init failed", "error", err)
			os.Exit(1)
		}
		return
	}

	// First-run detection
	if msg := firstRunCheck(); msg != "" {
		fmt.Println("╔══════════════════════════════════════════════════════╗")
		fmt.Println("║               Welcome to Remy!                      ║")
		fmt.Println("╚══════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Print(msg)
		fmt.Println("Starting with default configuration...")
		fmt.Println()
	}

	daemonMode := flag.Bool("daemon", false, "Run in daemon mode (no GUI, Telegram only)")
	showHelp := flag.Bool("help", false, "Show this help message")
	logFormat := flag.String("log-format", "text", "Log format: text or json")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Remy — Personal AI Assistant

Usage:
  remy                    Start the desktop GUI
  remy init               Initialize Remy configuration
  remy --daemon           Run in daemon mode (Telegram + scheduler, no GUI)
  remy --help             Show this help message

Flags:
  --daemon        Run in daemon mode (no GUI, Telegram only)
  --help          Show this help message
  --log-format    Log format: text or json (default: text)
  --log-level     Log level: debug, info, warn, error (default: info)

Commands:
  init            Create ~/.remy/ directory, default config, and default persona

Documentation:
  https://github.com/danmurf/remy#readme
`)
	}
	flag.Parse()

	// Configure structured logging
	var level slog.Level
	switch strings.ToLower(*logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if strings.ToLower(*logFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Handle "init" subcommand
	if flag.NArg() > 0 && flag.Arg(0) == "init" {
		runInit()
		return
	}

	slog.Info("Remy starting...", "version", version, "daemon", *daemonMode)
	runApp(*daemonMode)
}

// runApp initializes all Remy components and starts the application.
func runApp(daemonMode bool) {
	cfgPath, err := config.ConfigPath()
	if err != nil {
		slog.Error("Error determining config path", "error", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		slog.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	providerCfg, ok := cfg.Providers[cfg.DefaultProvider]
	if !ok {
		slog.Error("Default provider not found in config", "provider", cfg.DefaultProvider)
		os.Exit(1)
	}

	llmProvider, err := llm.NewProvider(providerCfg)
	if err != nil {
		slog.Error("Error creating LLM provider", "error", err)
		os.Exit(1)
	}

	dbPath := cfg.Memory.DBPath
	if dbPath != "" && dbPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Error getting home directory", "error", err)
			os.Exit(1)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}

	store, err := memory.NewStore(dbPath)
	if err != nil {
		slog.Error("Error opening database", "error", err, "db_path", dbPath)
		os.Exit(1)
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
			slog.Error("Error starting Telegram interface", "error", err)
			os.Exit(1)
		}
		slog.Info("Telegram interface started", "bot_token", maskToken(tgCfg.BotToken))
	}

	sched.Start(context.Background())

	slog.Info("Application initialized", "version", version, "daemon", daemonMode)

	if daemonMode {
		// Daemon mode: run Telegram + scheduler only, block until signal
		slog.Info("Running in daemon mode (no GUI)")
		select {}
	}

	application, err := app.NewApp(cfg, cfgPath, store, ag)
	if err != nil {
		_ = store.Close()
		slog.Error("Error creating app", "error", err)
		os.Exit(1)
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
		slog.Error("Error running application", "error", err)
		os.Exit(1)
	}

	slog.Info("Remy shutdown complete")
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
	slog.Info("Remy init — setting up your configuration...")

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Error getting home directory", "error", err)
		os.Exit(1)
	}

	remyDir := filepath.Join(home, ".remy")
	personaDir := filepath.Join(remyDir, "personas")

	// Create directories
	for _, dir := range []string{remyDir, personaDir} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			slog.Error("Error creating directory", "dir", dir, "error", err)
			os.Exit(1)
		}
	}

	// Create default config if it doesn't exist
	cfgPath := filepath.Join(remyDir, "config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		defaultCfg := config.DefaultConfig()
		if err := config.SaveConfig(cfgPath, defaultCfg); err != nil {
			slog.Error("Error saving default config", "error", err)
			os.Exit(1)
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
		if err := os.WriteFile(defaultPersona, []byte(content), 0o600); err != nil {
			slog.Error("Error creating default persona", "error", err)
			os.Exit(1)
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
