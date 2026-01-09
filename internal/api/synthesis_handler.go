package api

import (
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

func ListSynthesisTasks(c *gin.Context) {
	var tasks []model.SynthesisTask
	query := database.DB.Model(&model.SynthesisTask{})

	// Filters
	if text := c.Query("text"); text != "" {
		query = query.Where("text LIKE ?", "%"+text+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if voiceID := c.Query("voice_id"); voiceID != "" {
		query = query.Where("voice_id = ?", voiceID)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// Add 24h to include the end date fully
			query = query.Where("created_at < ?", t.Add(24*time.Hour))
		}
	}

	if err := query.Order("created_at desc").Find(&tasks).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 1, "Failed to fetch tasks")
		return
	}
	SuccessResponse(c, tasks)
}

type GenerateSpeechRequest struct {
	KeyID uint `json:"key_id"`
	minimax.T2ARequest
}

func GenerateSpeech(c *gin.Context) {
	var req GenerateSpeechRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "Invalid request body")
		return
	}

	if req.Text == "" && req.TextFileID == 0 {
		ErrorResponse(c, http.StatusBadRequest, 5, "Text or TextFileID is required")
		return
	}

	apiKey, err := getEffectiveKey(req.KeyID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 3, "Invalid API Key or No Default Key")
		return
	}

	client := minimax.NewClient(apiKey.Key)

	t2aReq := &req.T2ARequest

	resp, err := client.T2AAsync(t2aReq)
	task := model.SynthesisTask{
		Text:    req.Text,
		VoiceID: req.VoiceSetting.VoiceID,
		Status:  "processing",
	}
	if req.TextFileID > 0 {
		task.Text = fmt.Sprintf("FileID: %d", req.TextFileID)
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
}

func CheckTaskStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	keyIDStr := c.Query("key_id")

	var task model.SynthesisTask
	if err := database.DB.First(&task, id).Error; err != nil {
		ErrorResponse(c, http.StatusNotFound, 1, "Task not found")
		return
	}

	if task.Status == "success" || task.Status == "failed" {
		SuccessResponse(c, task)
		return
	}

	if task.TaskID == 0 {
		ErrorResponse(c, http.StatusBadRequest, 2, "Not an async task")
		return
	}

	keyID, _ := strconv.Atoi(keyIDStr)
	apiKey, err := getEffectiveKey(uint(keyID))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 3, "No valid API Key available")
		return
	}

	client := minimax.NewClient(apiKey.Key)
	qResp, err := client.T2AAsyncQuery(task.TaskID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 4, "Query Failed: "+err.Error())
		return
	}

	statusLower := qResp.Status

	if statusLower == "Success" {
		fResp, err := client.RetrieveFile(qResp.FileID)
		if err != nil {
			task.Error = "Retrieve failed: " + err.Error()
		} else {
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

func UploadTextFile(c *gin.Context) {
	keyIDStr := c.PostForm("key_id")
	keyID, _ := strconv.Atoi(keyIDStr)

	apiKey, err := getEffectiveKey(uint(keyID))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 1, "Invalid API Key or No Default Key")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, 2, "File upload required")
		return
	}

	// Validate extension
	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".txt" && ext != ".zip" {
		ErrorResponse(c, http.StatusBadRequest, 3, "Only .txt and .zip files are allowed")
		return
	}

	tempDir := "uploads"
	os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename))
	if err := c.SaveUploadedFile(fileHeader, tempPath); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 4, "Failed to save file")
		return
	}
	defer os.Remove(tempPath)

	client := minimax.NewClient(apiKey.Key)
	resp, err := client.UploadFile(tempPath, "t2a_async_input")
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 5, "Minimax Upload Failed: "+err.Error())
		return
	}

	SuccessResponse(c, resp.File)
}
