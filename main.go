package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"minimax-voice-workbench/internal/api"
	"minimax-voice-workbench/internal/database"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist/*
var staticFS embed.FS

func main() {
	// Initialize Database
	database.InitDB(".")

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
			// If file exists in distFS, serve it
			if f, err := distFS.Open(path[1:]); err == nil {
				f.Close()
				c.FileFromFS(path, http.FS(distFS))
				return
			}
			// Fallback to index.html for Vue Router history mode
			c.FileFromFS("/", http.FS(distFS))
		})
	} else {
		log.Println("Static FS not found or invalid (expected during dev before build):", err)
	}

	// Make sure generated dir exists
	// Open Browser
	go func() {
		OpenBrowser("http://localhost:8080")
	}()

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
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
