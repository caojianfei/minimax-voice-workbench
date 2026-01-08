package api

import (
	"encoding/hex"
	"fmt"
	"io"
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

	apiKey, err := getEffectiveKey(uint(keyID))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "Invalid API Key or No Default Key")
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
		// Use Unscoped to include soft-deleted records
		err := database.DB.Unscoped().Where("voice_id = ?", vInfo.VoiceID).First(&existing).Error

		if err == nil {
			// Record exists (may be soft-deleted)
			if existing.DeletedAt.Valid {
				// Restore soft-deleted record
				database.DB.Unscoped().Model(&existing).Updates(map[string]interface{}{
					"deleted_at": nil,
					"name":       vInfo.VoiceName,
					"type":       vType,
				})
				count++
			}
			// If not deleted, just skip (already exists)
			return
		}

		// Create new record
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
	promptText := c.PostForm("prompt_text") // Optional: text for prompt audio
	demoText := c.PostForm("demo_text")     // Optional: text for demo generation
	speechModel := c.PostForm("model")      // Optional: speech model
	noiseReduction := c.PostForm("noise_reduction") == "true"
	volumeNorm := c.PostForm("volume_normalization") == "true"
	watermark := c.PostForm("watermark") == "true"

	if name == "" {
		ErrorResponse(c, http.StatusBadRequest, 2, "name is required")
		return
	}

	keyID, _ := strconv.Atoi(keyIDStr)
	apiKey, err := getEffectiveKey(uint(keyID))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 3, "Invalid API Key or No Default Key")
		return
	}

	// 2. Get main clone audio file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 4, "File upload required")
		return
	}

	// 3. Save main file temporarily
	tempDir := "uploads"
	os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename))
	if err := c.SaveUploadedFile(fileHeader, tempPath); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 5, "Failed to save file")
		return
	}
	defer os.Remove(tempPath)

	// 4. Upload main clone audio to Minimax
	client := minimax.NewClient(apiKey.Key)
	uploadResp, err := client.UploadFile(tempPath, "voice_clone")
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 6, "Minimax Upload Failed: "+err.Error())
		return
	}

	// 5. Handle optional prompt audio
	var clonePrompt *minimax.ClonePrompt
	promptFileHeader, err := c.FormFile("prompt_file")
	if err == nil && promptFileHeader != nil {
		// Save prompt file temporarily
		promptTempPath := filepath.Join(tempDir, fmt.Sprintf("%d_prompt_%s", time.Now().Unix(), promptFileHeader.Filename))
		if err := c.SaveUploadedFile(promptFileHeader, promptTempPath); err == nil {
			defer os.Remove(promptTempPath)

			// Upload prompt audio
			promptUploadResp, err := client.UploadFile(promptTempPath, "prompt_audio")
			if err == nil {
				clonePrompt = &minimax.ClonePrompt{
					PromptAudio: promptUploadResp.File.FileID,
					PromptText:  promptText,
				}
			}
		}
	}

	// 6. Call Voice Clone
	voiceID := fmt.Sprintf("Clone_%d", time.Now().Unix())

	// Set default model if demo text provided
	if demoText != "" && speechModel == "" {
		speechModel = "speech-2.6-hd"
	}

	cloneReq := &minimax.VoiceCloneRequest{
		FileID:                  uploadResp.File.FileID,
		VoiceID:                 voiceID,
		ClonePrompt:             clonePrompt,
		Text:                    demoText,
		Model:                   speechModel,
		LanguageBoost:           "auto",
		NeedNoiseReduction:      noiseReduction,
		NeedVolumeNormalization: volumeNorm,
		AigcWatermark:           watermark,
	}

	cloneResp, err := client.VoiceClone(cloneReq)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 7, "Minimax Voice Clone Failed: "+err.Error())
		return
	}

	// 7. Download demo audio if available
	var demoAudioPath string
	if cloneResp.DemoAudio != "" {
		outputDir := "generated"
		os.MkdirAll(outputDir, 0755)
		filename := fmt.Sprintf("demo_%s.mp3", voiceID)
		demoFilePath := filepath.Join(outputDir, filename)

		// Download from URL
		if err := downloadAudioFromURL(cloneResp.DemoAudio, demoFilePath); err == nil {
			demoAudioPath = "/files/" + filename
		}
	}

	// 8. Save to DB
	voice := model.Voice{
		Name:      name,
		VoiceID:   voiceID,
		Type:      "cloned",
		DemoAudio: demoAudioPath,
	}

	if err := database.DB.Create(&voice).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 8, "Failed to save voice to DB")
		return
	}

	SuccessResponse(c, voice)
}

// Helper function to download audio from URL
func downloadAudioFromURL(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// DesignVoiceRequest ...
type DesignVoiceRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	PreviewText string `json:"preview_text" binding:"required"`
	KeyID       uint   `json:"key_id"`
	Name        string `json:"name"`
	Watermark   bool   `json:"watermark"`
}

func DesignVoice(c *gin.Context) {
	var req DesignVoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 1, "Invalid design request")
		return
	}

	apiKey, err := getEffectiveKey(req.KeyID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "Invalid API Key or No Default Key")
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
	if voice.Type == "cloned" || voice.Type == "generated" {
		keyID, _ := strconv.Atoi(keyIDStr)
		apiKey, err := getEffectiveKey(uint(keyID))
		if err == nil {
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
