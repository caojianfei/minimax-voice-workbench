package api

import (
	"encoding/hex"
	"fmt"
	"minimax-voice-workbench/internal/database"
	"minimax-voice-workbench/internal/model"
	"minimax-voice-workbench/pkg/minimax"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ListVoices returns combined list of voices (DB)
func ListVoices(c *gin.Context) {
	var voices []model.Voice
	if err := database.DB.Find(&voices).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 1, "Failed to fetch voices")
		return
	}
	SuccessResponse(c, voices)
}

// SyncVoices fetches voices from Minimax and updates DB
func SyncVoices(c *gin.Context) {
	keyIDStr := c.Query("key_id")
	keyID, _ := strconv.Atoi(keyIDStr)

	var apiKey model.ApiKey
	if err := database.DB.First(&apiKey, keyID).Error; err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "Invalid API Key ID")
		return
	}

	client := minimax.NewClient(apiKey.Key)
	resp, err := client.GetVoices("all")
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 3, "Sync Failed: "+err.Error())
		return
	}

	count := 0
	// Helper to upsert
	upsert := func(vInfo minimax.VoiceInfo, vType string) {
		var existing model.Voice
		if database.DB.Where("voice_id = ?", vInfo.VoiceID).First(&existing).Error == nil {
			// Update?
			return
		}
		newVoice := model.Voice{
			Name:    vInfo.VoiceName,
			VoiceID: vInfo.VoiceID,
			Type:    vType,
		}
		if newVoice.Name == "" {
			newVoice.Name = vInfo.VoiceID
		}

		database.DB.Create(&newVoice)
		count++
	}

	for _, v := range resp.SystemVoices {
		upsert(v, "system")
	}
	for _, v := range resp.VoiceCloning {
		upsert(v, "cloned")
	}
	for _, v := range resp.VoiceGeneration {
		upsert(v, "generated")
	}

	SuccessResponse(c, map[string]int{"added": count})
}

// CloneVoice requests...
func CloneVoice(c *gin.Context) {
	// 1. Parse Form
	name := c.PostForm("name")
	keyIDStr := c.PostForm("key_id")
	if name == "" || keyIDStr == "" {
		ErrorResponse(c, http.StatusBadRequest, 2, "name and key_id are required")
		return
	}

	keyID, _ := strconv.Atoi(keyIDStr)
	var apiKey model.ApiKey
	if err := database.DB.First(&apiKey, keyID).Error; err != nil {
		ErrorResponse(c, http.StatusBadRequest, 3, "Invalid API Key ID")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 4, "File upload required")
		return
	}

	// 2. Save file temporarily
	tempDir := "uploads"
	os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename))
	if err := c.SaveUploadedFile(fileHeader, tempPath); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 5, "Failed to save file")
		return
	}
	defer os.Remove(tempPath)

	// 3. Upload to Minimax
	client := minimax.NewClient(apiKey.Key)
	uploadResp, err := client.UploadFile(tempPath, "voice_clone")
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 6, "Minimax Upload Failed: "+err.Error())
		return
	}

	// 4. Call Voice Clone
	voiceID := fmt.Sprintf("Clone_%d", time.Now().Unix())

	cloneReq := &minimax.VoiceCloneRequest{
		FileID:  uploadResp.FileID,
		VoiceID: voiceID,
	}

	_, err = client.VoiceClone(cloneReq)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 7, "Minimax Voice Clone Failed: "+err.Error())
		return
	}

	// 5. Save to DB
	voice := model.Voice{
		Name:    name,
		VoiceID: voiceID,
		Type:    "cloned",
	}

	if err := database.DB.Create(&voice).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 8, "Failed to save voice to DB")
		return
	}

	SuccessResponse(c, voice)
}

// DesignVoiceRequest ...
type DesignVoiceRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	PreviewText string `json:"preview_text" binding:"required"`
	KeyID       uint   `json:"key_id" binding:"required"`
	Name        string `json:"name"`
	Watermark   bool   `json:"watermark"`
}

func DesignVoice(c *gin.Context) {
	var req DesignVoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 1, "Invalid design request")
		return
	}

	var apiKey model.ApiKey
	if err := database.DB.First(&apiKey, req.KeyID).Error; err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "Invalid API Key ID")
		return
	}

	client := minimax.NewClient(apiKey.Key)

	designReq := &minimax.VoiceDesignRequest{
		Prompt:      req.Prompt,
		PreviewText: req.PreviewText,
	}

	resp, err := client.VoiceDesign(designReq)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 3, "Design Failed: "+err.Error())
		return
	}

	// Make Preview File
	// resp.TrialAudio is hex encoded
	audioBytes, err := hex.DecodeString(resp.TrialAudio)
	var previewPath string
	if err == nil {
		outputDir := "generated"
		os.MkdirAll(outputDir, 0755)
		filename := fmt.Sprintf("preview_%s.mp3", resp.VoiceID)
		filepathStr := filepath.Join(outputDir, filename)
		os.WriteFile(filepathStr, audioBytes, 0644)
		previewPath = "/files/" + filename
	}

	// Save Voice to DB
	voice := model.Voice{
		Name:    req.Name,
		VoiceID: resp.VoiceID,
		Type:    "generated",
		Preview: previewPath,
	}
	if voice.Name == "" {
		voice.Name = "Designed " + resp.VoiceID[:8]
	}

	if err := database.DB.Create(&voice).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 4, "Failed to save generated voice")
		return
	}

	SuccessResponse(c, voice)
}

func DeleteVoice(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	keyIDStr := c.Query("key_id")

	var voice model.Voice
	if err := database.DB.First(&voice, id).Error; err != nil {
		ErrorResponse(c, http.StatusNotFound, 1, "Voice not found")
		return
	}

	// Attempt remote delete if key_id provided and voice is remote type
	if keyIDStr != "" && (voice.Type == "cloned" || voice.Type == "generated") {
		keyID, _ := strconv.Atoi(keyIDStr)
		var apiKey model.ApiKey
		if err := database.DB.First(&apiKey, keyID).Error; err == nil {
			client := minimax.NewClient(apiKey.Key)
			mapping := map[string]string{
				"cloned":    "voice_cloning",
				"generated": "voice_generation",
			}
			if vType, ok := mapping[voice.Type]; ok {
				client.DeleteVoice(vType, voice.VoiceID)
				// Log error but proceed to delete local?
			}
		}
	}

	if err := database.DB.Delete(&model.Voice{}, id).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 9, "Failed to delete voice")
		return
	}
	SuccessResponse(c, nil)
}
