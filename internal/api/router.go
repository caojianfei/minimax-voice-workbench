package api

import (
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

	api := r.Group("/api")
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
	r.Static("/files", "./generated")
}
