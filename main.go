package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"minimax-voice-workbench/internal/api"
	"minimax-voice-workbench/internal/config"
	"minimax-voice-workbench/internal/database"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist/* web/dist/assets/*
var staticFS embed.FS

func main() {
	// Initialize Config
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize Database
	database.InitDB(config.GlobalConfig.Storage.BasePath)

	r := gin.Default()

	// API Routes
	api.SetupRouter(r)

	// Static Frontend Serving
	// Check if we are running in dev mode or prod (embedded)
	// For now, simpler: always try to serve embedded if available, else standard fs

	// Simplify: In dev, we might run separate frontend, but for "single executable" requirement:
	// We assume "web/dist" exists on build.
	// To make "go run main.go" work before fetch, we need to handle "pattern not found".
	// But go:embed requires pattern to exist.
	// So we might need to create a dummy web/dist if not exists, or comment out for dev?
	// User Requirement: "Generate a single executable".
	// I will use a separate function to mount static to avoid build errors if dist missing during dev steps.
	// Actually, just serve "/" from "web/dist" if embed works.

	// Setup Embed FS
	distFS, err := fs.Sub(staticFS, "web/dist")
	if err == nil {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			relPath := strings.TrimPrefix(path, "/")
			if relPath == "" {
				relPath = "index.html"
			}

			// If file exists in distFS, serve it
			if b, err := fs.ReadFile(distFS, relPath); err == nil {
				contentType := mime.TypeByExtension(filepath.Ext(relPath))
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				c.Data(http.StatusOK, contentType, b)
				return
			}
			// Fallback to index.html for Vue Router history mode
			if b, err := fs.ReadFile(distFS, "index.html"); err == nil {
				contentType := mime.TypeByExtension(".html")
				if contentType == "" {
					contentType = "text/html; charset=utf-8"
				}
				c.Data(http.StatusOK, contentType, b)
				return
			}
			c.Status(http.StatusNotFound)
		})
	} else {
		log.Println("Static FS not found or invalid (expected during dev before build):", err)
	}

	port := config.GlobalConfig.Server.Port
	addr := fmt.Sprintf(":%d", port)

	// Make sure generated dir exists
	// Open Browser
	go func() {
		OpenBrowser("http://localhost" + addr)
	}()

	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Println("Failed to open browser:", err)
	}
}
