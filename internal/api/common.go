package api

import (
	"errors"
	"minimax-voice-workbench/internal/database"
	"minimax-voice-workbench/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"` // 0 for success, others for error
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse sends a standard success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ErrorResponse sends a standard error response
func ErrorResponse(c *gin.Context, httpCode int, errCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: message,
	})
}

// getEffectiveKey returns the specified key or the default key
func getEffectiveKey(keyID uint) (*model.ApiKey, error) {
	var apiKey model.ApiKey
	var err error

	if keyID > 0 {
		err = database.DB.First(&apiKey, keyID).Error
	} else {
		// Try to find default key
		err = database.DB.Where("is_default = ?", true).First(&apiKey).Error
		// Fallback: if no default key, try get the first one
		if err != nil {
			err = database.DB.Order("created_at asc").First(&apiKey).Error
		}
	}

	if err != nil {
		return nil, errors.New("no valid API key found")
	}
	return &apiKey, nil
}

