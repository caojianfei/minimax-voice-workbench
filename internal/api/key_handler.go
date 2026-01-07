package api

import (
	"minimax-voice-workbench/internal/database"
	"minimax-voice-workbench/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListKeys returns all API keys
func ListKeys(c *gin.Context) {
	var keys []model.ApiKey
	result := database.DB.Find(&keys)
	if result.Error != nil {
		ErrorResponse(c, http.StatusInternalServerError, 1, "Failed to fetch keys")
		return
	}
	SuccessResponse(c, keys)
}

// AddKeyRequest defines the body for adding a key
type AddKeyRequest struct {
	Platform string `json:"platform"`
	Key      string `json:"key" binding:"required"`
}

// AddKey creates a new API key
func AddKey(c *gin.Context) {
	var req AddKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "Invalid request body")
		return
	}

	if req.Platform == "" {
		req.Platform = "minimax"
	}

	apiKey := model.ApiKey{
		Platform: req.Platform,
		Key:      req.Key,
	}

	result := database.DB.Create(&apiKey)
	if result.Error != nil {
		ErrorResponse(c, http.StatusInternalServerError, 3, "Failed to save key")
		return
	}

	SuccessResponse(c, apiKey)
}

// DeleteKey removes an API key
func DeleteKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 4, "Invalid ID format")
		return
	}

	result := database.DB.Delete(&model.ApiKey{}, id)
	if result.Error != nil {
		ErrorResponse(c, http.StatusInternalServerError, 5, "Failed to delete key")
		return
	}

	if result.RowsAffected == 0 {
		ErrorResponse(c, http.StatusNotFound, 6, "Key not found")
		return
	}

	SuccessResponse(c, nil)
}
