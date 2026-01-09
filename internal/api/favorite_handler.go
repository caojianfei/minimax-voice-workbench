package api

import (
	"net/http"

	"minimax-voice-workbench/internal/database"
	"minimax-voice-workbench/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListFavorites(c *gin.Context) {
	var rows []model.Voice
	if err := database.DB.Model(&model.Voice{}).Select("voice_id").Where("is_favorite = ?", true).Find(&rows).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 1, "Failed to list favorites")
		return
	}

	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.VoiceID)
	}

	SuccessResponse(c, ids)
}

func ToggleFavorite(c *gin.Context) {
	voiceID := c.Param("voice_id")
	if voiceID == "" {
		ErrorResponse(c, http.StatusBadRequest, 1, "Missing voice_id")
		return
	}

	var voice model.Voice
	if err := database.DB.Where("voice_id = ?", voiceID).First(&voice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, 2, "Voice not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, 3, "Failed to query voice")
		return
	}

	next := !voice.IsFavorite
	if err := database.DB.Model(&voice).Update("is_favorite", next).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 4, "Failed to update favorite")
		return
	}

	SuccessResponse(c, gin.H{"voice_id": voiceID, "favorite": next})
}
