// Package main is the entry point for the Remy personal assistant binary.
package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/danmurf/remy/internal/app"
	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/llm"
	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/scheduler"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var version = "dev"

func main() {
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
	defer func() {
		_ = store.Close()
	}()

	embedder := memory.NewEmbedder(providerCfg.Endpoint, providerCfg.EmbeddingModel)

	personaLoader := app.NewPersonaLoader()

	sched := scheduler.NewScheduler(store, nil)

	application, err := app.NewApp(cfg, store, llmProvider, embedder, personaLoader, sched)
	if err != nil {
		log.Fatalf("Error creating app: %v", err)
	}

	// Start the scheduler
	sched.Start(context.Background())

	fmt.Printf("Remy %s starting...\n", version)

	// Create Wails app with options
	err = wails.Run(&options.App{
		Title:  "Remy",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  application.Startup,
		OnShutdown: application.Shutdown,
		Bind: []any{
			application,
		},
	})

	if err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
