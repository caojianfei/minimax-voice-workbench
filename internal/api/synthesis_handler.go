package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"minimax-voice-workbench/internal/database"
	"minimax-voice-workbench/internal/model"
	"minimax-voice-workbench/pkg/minimax"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ListSynthesisTasks 获取语音合成任务列表
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

// GenerateSpeech 提交异步语音合成任务
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

	payloadBytes, _ := json.Marshal(t2aReq)

	resp, err := client.T2AAsync(t2aReq)
	task := model.SynthesisTask{
		Text:           req.Text,
		VoiceID:        req.VoiceSetting.VoiceID,
		Format:         req.AudioSetting.Format,
		SampleRate:     req.AudioSetting.AudioSampleRate,
		Channel:        req.AudioSetting.Channel,
		Status:         "processing",
		RequestPayload: string(payloadBytes),
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

var taskMutexes sync.Map

// CheckTaskStatus 检查异步任务状态并下载结果
func CheckTaskStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	keyIDStr := c.Query("key_id")

	// First check: Read directly from DB to see if it's already done
	var task model.SynthesisTask
	if err := database.DB.First(&task, id).Error; err != nil {
		ErrorResponse(c, http.StatusNotFound, 1, "Task not found")
		return
	}

	if task.Status == "success" || task.Status == "failed" {
		SuccessResponse(c, task)
		return
	}

	// Concurrency control: Lock based on task ID
	// This prevents multiple requests from triggering download simultaneously
	muVal, _ := taskMutexes.LoadOrStore(id, &sync.Mutex{})
	mu := muVal.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	// Second check: Re-fetch task status after acquiring lock
	// Another request might have finished the task while we were waiting for the lock
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

	switch statusLower {
	case "Success":
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
	case "Failed", "Expired":
		task.Status = "failed"
		task.Error = "Remote status: " + statusLower
	default:
		task.Status = "processing"
	}

	database.DB.Save(&task)
	SuccessResponse(c, task)
}

// downloadFile 从指定 URL 下载音频文件并保存到本地
func downloadFile(url string, task *model.SynthesisTask) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	outputDir := filepath.Join("generated", "audios")
	os.MkdirAll(outputDir, 0755)

	ext := task.Format
	if ext == "" {
		ext = "mp3"
	}
	filename := fmt.Sprintf("audio_%d.%s", task.ID, ext)

	outputPath := filepath.Join(outputDir, filename)

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	contentType := resp.Header.Get("Content-Type")
	mediaType, params, parseErr := mime.ParseMediaType(contentType)
	if parseErr == nil && strings.HasPrefix(mediaType, "multipart/") && params["boundary"] != "" {
		reader := multipart.NewReader(resp.Body, params["boundary"])
		part, err := reader.NextPart()
		if err != nil {
			return err
		}
		defer part.Close()

		if _, err := io.Copy(out, part); err != nil {
			return err
		}
	} else {
		if _, err := io.Copy(out, resp.Body); err != nil {
			return err
		}
	}

	task.Output = "/files/audios/" + filename
	return nil
}

// DeleteSynthesisTask 删除语音合成任务及其对应的音频文件
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

// UploadTextFile 上传文本文件用于异步语音合成
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
