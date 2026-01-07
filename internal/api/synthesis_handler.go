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

// ListSynthesisTasks returns history
func ListSynthesisTasks(c *gin.Context) {
	var tasks []model.SynthesisTask
	// Order by created_at desc
	if err := database.DB.Order("created_at desc").Find(&tasks).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 1, "Failed to fetch tasks")
		return
	}
	SuccessResponse(c, tasks)
}

// GenerateSpeechRequest ...
type GenerateSpeechRequest struct {
	Text    string  `json:"text" binding:"required"`
	VoiceID string  `json:"voice_id" binding:"required"`
	Speed   float64 `json:"speed"`
	Vol     float64 `json:"vol"`
	KeyID   uint    `json:"key_id" binding:"required"`
	Model   string  `json:"model"`
	Mode    string  `json:"mode"` // "sync" or "async"
}

func GenerateSpeech(c *gin.Context) {
	var req GenerateSpeechRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "Invalid request body")
		return
	}

	var apiKey model.ApiKey
	if err := database.DB.First(&apiKey, req.KeyID).Error; err != nil {
		ErrorResponse(c, http.StatusBadRequest, 3, "Invalid API Key ID")
		return
	}

	// Defaults
	if req.Model == "" {
		req.Model = "speech-01-turbo"
	}
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	if req.Vol == 0 {
		req.Vol = 1.0
	}

	client := minimax.NewClient(apiKey.Key)

	// Async Mode
	if req.Mode == "async" {
		t2aReq := &minimax.T2ARequest{
			Model: req.Model,
			Text:  req.Text,
			VoiceSetting: minimax.VoiceSetting{
				VoiceID: req.VoiceID,
				Speed:   req.Speed,
				Vol:     req.Vol,
			},
			AudioSetting: minimax.AudioSetting{
				SampleRate: 32000,
				Bitrate:    128000,
				Format:     "mp3",
				Channel:    1,
			},
		}

		resp, err := client.T2AAsync(t2aReq)
		task := model.SynthesisTask{
			Text:    req.Text,
			VoiceID: req.VoiceID,
			Status:  "processing",
		}

		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			database.DB.Create(&task)
			ErrorResponse(c, http.StatusInternalServerError, 4, "Async Submit Failed: "+err.Error())
			return
		}

		task.TaskID = resp.TaskID
		database.DB.Create(&task)
		SuccessResponse(c, task)
		return
	}

	// Sync Mode
	t2aReq := &minimax.T2ARequest{
		Model:  req.Model,
		Text:   req.Text,
		Stream: false,
		VoiceSetting: minimax.VoiceSetting{
			VoiceID: req.VoiceID,
			Speed:   req.Speed,
			Vol:     req.Vol,
		},
		AudioSetting: minimax.AudioSetting{
			SampleRate: 32000,
			Bitrate:    128000,
			Format:     "mp3",
			Channel:    1,
		},
	}

	resp, err := client.T2A(t2aReq)

	task := model.SynthesisTask{
		Text:    req.Text,
		VoiceID: req.VoiceID,
		Status:  "failed",
	}

	if err != nil {
		task.Error = err.Error()
		database.DB.Create(&task)
		ErrorResponse(c, http.StatusInternalServerError, 4, "Synthesis Failed: "+err.Error())
		return
	}

	// Decode Audio
	audioBytes, err := hex.DecodeString(resp.Data.Audio)
	if err != nil {
		task.Error = "Hex decode failed"
		database.DB.Create(&task)
		ErrorResponse(c, http.StatusInternalServerError, 5, "Failed to decode audio")
		return
	}

	// Save to file
	outputDir := "generated"
	os.MkdirAll(outputDir, 0755)
	filename := fmt.Sprintf("audio_%d.mp3", time.Now().UnixNano())
	outputPath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(outputPath, audioBytes, 0644); err != nil {
		task.Error = "Save file failed"
		database.DB.Create(&task)
		ErrorResponse(c, http.StatusInternalServerError, 6, "Failed to save audio file")
		return
	}

	task.Status = "success"
	task.Output = "/files/" + filename
	database.DB.Create(&task)

	SuccessResponse(c, task)
}

// CheckTaskStatus checks status of async task and downloads if done
func CheckTaskStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	keyIDStr := c.Query("key_id") // Pass key_id to check status

	var task model.SynthesisTask
	if err := database.DB.First(&task, id).Error; err != nil {
		ErrorResponse(c, http.StatusNotFound, 1, "Task not found")
		return
	}

	// If already success or failed (and verified), just return
	if task.Status == "success" || task.Status == "failed" {
		SuccessResponse(c, task)
		return
	}

	if task.TaskID == 0 {
		ErrorResponse(c, http.StatusBadRequest, 2, "Not an async task")
		return
	}

	keyID, _ := strconv.Atoi(keyIDStr)
	var apiKey model.ApiKey
	if err := database.DB.First(&apiKey, keyID).Error; err != nil {
		// Use first available key as fallback or error?
		// For async check, any valid key *might* not work if tasks are scoped to user/account.
		// We assume same account key is used.
		var keys []model.ApiKey
		database.DB.Find(&keys)
		if len(keys) > 0 {
			apiKey = keys[0]
		} else {
			ErrorResponse(c, http.StatusBadRequest, 3, "No API Key available")
			return
		}
	}

	client := minimax.NewClient(apiKey.Key)
	qResp, err := client.T2AAsyncQuery(task.TaskID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 4, "Query Failed: "+err.Error())
		return
	}

	// Update Status
	// Status: Processing, Success, Failed, Expired
	statusLower := qResp.Status // Check case? Docs say "Processing", "Success".

	if statusLower == "Success" {
		// Download file
		// Need FileID?
		// Response has file_id
		fResp, err := client.RetrieveFile(qResp.FileID)
		if err != nil {
			task.Error = "Retrieve failed: " + err.Error()
			// Don't mark failed yet, maybe retry?
		} else {
			// Download URL
			err := downloadFile(fResp.File.DownloadURL, &task)
			if err != nil {
				task.Error = "Download failed: " + err.Error()
			} else {
				task.Status = "success"
			}
		}
	} else if statusLower == "Failed" || statusLower == "Expired" {
		task.Status = "failed"
		task.Error = "Remote status: " + statusLower
	} else {
		task.Status = "processing"
	}

	database.DB.Save(&task)
	SuccessResponse(c, task)
}

func downloadFile(url string, task *model.SynthesisTask) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outputDir := "generated"
	os.MkdirAll(outputDir, 0755)
	filename := fmt.Sprintf("async_audio_%d.mp3", time.Now().UnixNano())
	outputPath := filepath.Join(outputDir, filename)

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	task.Output = "/files/" + filename
	return nil
}

func DeleteSynthesisTask(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var task model.SynthesisTask
	if err := database.DB.First(&task, id).Error; err == nil {
		if len(task.Output) > 7 {
			fName := task.Output[7:]
			os.Remove(filepath.Join("generated", fName))
		}
	}

	if err := database.DB.Delete(&model.SynthesisTask{}, id).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 7, "Failed to delete task")
		return
	}
	SuccessResponse(c, nil)
}
