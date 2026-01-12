package api

import (
	"path/filepath"

	"minimax-voice-workbench/internal/auth"
	"minimax-voice-workbench/internal/config"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	// Enable CORS for dev
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Login Route (Public, Rate Limited)
	r.POST("/api/login", auth.LoginRateLimitMiddleware(), LoginHandler)

	// Public Config
	r.GET("/api/config", GetPublicConfig)

	api := r.Group("/api")
	// Apply Auth Middleware
	api.Use(auth.AuthMiddleware())
	{
		// Keys
		api.GET("/keys", ListKeys)
		api.POST("/keys", AddKey)
		api.DELETE("/keys/:id", DeleteKey)
		api.PUT("/keys/:id/default", SetDefaultKey)

		// Voices
		api.GET("/voices", ListVoices)
		api.POST("/voices/clone", CloneVoice)
		api.POST("/voices/sync", SyncVoices)
		api.POST("/voices/design", DesignVoice)
		api.POST("/voices/preview", GeneratePreview)
		api.DELETE("/voices/:id", DeleteVoice)

		// Favorites
		api.GET("/favorites", ListFavorites)
		api.POST("/favorites/:voice_id/toggle", ToggleFavorite)

		// Synthesis
		api.GET("/synthesis", ListSynthesisTasks)
		api.POST("/synthesis", GenerateSpeech)
		api.POST("/synthesis/upload", UploadTextFile)
		api.GET("/synthesis/:id/status", CheckTaskStatus)
		api.DELETE("/synthesis/:id", DeleteSynthesisTask)
	}

	// Static files for generated audio
	// Use configured storage path
	basePath := config.GlobalConfig.Storage.BasePath
	// Default to "./generated" if basePath is current dir, effectively.
	// Actually config defaults basePath to CWD.
	generatedPath := filepath.Join(basePath, "generated")
	r.Static("/files", generatedPath)
}
