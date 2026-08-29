package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"facebook-responder/internal/ai"
	"facebook-responder/internal/config"
	"facebook-responder/internal/facebook"
	"facebook-responder/internal/storage"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("exe: %v", err)
	}
	dataDir := filepath.Join(filepath.Dir(exe), "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("data dir: %v", err)
	}
	configPath := filepath.Join(dataDir, "config.json")
	storePath := filepath.Join(dataDir, "comments.json")

	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	store, err := storage.NewStore(storePath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	c := cfg.Get()
	var fb facebook.Client
	var aiC ai.Client
	if c.UsarMock {
		fb = facebook.NewMockClient()
		aiC = ai.NewMock()
	} else {
		fb = facebook.NewGraphClient(c.PageID)
		aiC = ai.NewGemini(c.GeminiAPIKey)
	}

	app := NewApp(cfg, store, fb, aiC)

	if err := wails.Run(&options.App{
		Title:  "Facebook Comments AI Responder",
		Width:  1200,
		Height: 820,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 24, G: 26, B: 32, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	}); err != nil {
		log.Fatalf("wails: %v", err)
	}
}
