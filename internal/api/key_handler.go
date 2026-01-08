package api

import (
	"minimax-voice-workbench/internal/database"
	"minimax-voice-workbench/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	Remark   string `json:"remark"`
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

	// Check if this is the first key
	var count int64
	database.DB.Model(&model.ApiKey{}).Count(&count)
	isDefault := count == 0

	apiKey := model.ApiKey{
		Platform:  req.Platform,
		Key:       req.Key,
		Remark:    req.Remark,
		IsDefault: isDefault,
	}

	result := database.DB.Create(&apiKey)
	if result.Error != nil {
		ErrorResponse(c, http.StatusInternalServerError, 3, "Failed to save key")
		return
	}

	SuccessResponse(c, apiKey)
}

// SetDefaultKey sets a key as default
func SetDefaultKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 4, "Invalid ID format")
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Reset all to false
		if err := tx.Model(&model.ApiKey{}).Where("1 = 1").Update("is_default", false).Error; err != nil {
			return err
		}
		// Set target to true
		result := tx.Model(&model.ApiKey{}).Where("id = ?", id).Update("is_default", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, 6, "Key not found")
		} else {
			ErrorResponse(c, http.StatusInternalServerError, 7, "Failed to set default key")
		}
		return
	}

	SuccessResponse(c, nil)
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
