package minimax

// T2ARequest represents request for Text-to-Audio (Sync & Async)
type T2ARequest struct {
	Model             string         `json:"model"` // speech-01-turbo, speech-01-hd
	Text              string         `json:"text"`
	Stream            bool           `json:"stream,omitempty"` // For sync
	VoiceSetting      VoiceSetting   `json:"voice_setting"`
	AudioSetting      AudioSetting   `json:"audio_setting"`
	PronunciationDict map[string]any `json:"pronunciation_dict,omitempty"`
}

type VoiceSetting struct {
	VoiceID string  `json:"voice_id"`
	Speed   float64 `json:"speed,omitempty"`
	Vol     float64 `json:"vol,omitempty"`
	Pitch   float64 `json:"pitch,omitempty"`
	Emotion string  `json:"emotion,omitempty"` // happy, sad, etc.
}

type AudioSetting struct {
	SampleRate int    `json:"sample_rate,omitempty"` // 32000
	Bitrate    int    `json:"bitrate,omitempty"`     // 128000
	Format     string `json:"format,omitempty"`      // mp3, wav
	Channel    int    `json:"channel,omitempty"`     // 1
}

// T2AResponse represents response from Sync T2A
type T2AResponse struct {
	Data      T2AData   `json:"data"`
	ExtraInfo ExtraInfo `json:"extra_info"`
	BaseResp  BaseResp  `json:"base_resp"`
	TraceID   string    `json:"trace_id"`
}

type T2AData struct {
	Audio  string `json:"audio"` // Hex/Base64
	Status int    `json:"status"`
}

type ExtraInfo struct {
	AudioLength int `json:"audio_length"`
}

type BaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// Async T2A
type T2AAsyncResponse struct {
	TaskID   int64    `json:"task_id"`
	BaseResp BaseResp `json:"base_resp"`
}

type T2AAsyncQueryResponse struct {
	TaskID   int64    `json:"task_id"`
	Status   string   `json:"status"` // Processing, Success, Failed, Expired
	FileID   int64    `json:"file_id"`
	BaseResp BaseResp `json:"base_resp"`
}

// File Retrieve (for async download)
type FileRetrieveResponse struct {
	File     FileRetrieveData `json:"file"`
	BaseResp BaseResp         `json:"base_resp"`
}

type FileRetrieveData struct {
	DownloadURL string `json:"download_url"`
}

// Voice Cloning
type VoiceCloneRequest struct {
	FileID      string      `json:"file_id"`  // Uploaded file ID
	VoiceID     string      `json:"voice_id"` // Custom ID
	ClonePrompt ClonePrompt `json:"clone_prompt,omitempty"`
	Text        string      `json:"text,omitempty"` // For demo generation
	Model       string      `json:"model,omitempty"`
}

type ClonePrompt struct {
	PromptAudio string `json:"prompt_audio"`
	PromptText  string `json:"prompt_text"`
}

type VoiceCloneResponse struct {
	BaseResp  BaseResp `json:"base_resp"`
	DemoAudio string   `json:"demo_audio"` // URL
}

// Upload
type UploadResponse struct {
	FileID   string   `json:"file_id"` // String in upload, Int64 in async? Check docs. Upload doc says string.
	BaseResp BaseResp `json:"base_resp"`
}

// Voice Design
type VoiceDesignRequest struct {
	Prompt      string `json:"prompt"`
	PreviewText string `json:"preview_text"`
	VoiceID     string `json:"voice_id,omitempty"`
}

type VoiceDesignResponse struct {
	VoiceID    string   `json:"voice_id"`
	TrialAudio string   `json:"trial_audio"` // Hex
	BaseResp   BaseResp `json:"base_resp"`
}

// Voice Management
type GetVoicesRequest struct {
	VoiceType string `json:"voice_type"` // all, system, voice_cloning, voice_generation
}

type GetVoicesResponse struct {
	SystemVoices    []VoiceInfo `json:"system_voice"`
	VoiceCloning    []VoiceInfo `json:"voice_cloning"`
	VoiceGeneration []VoiceInfo `json:"voice_generation"`
	BaseResp        BaseResp    `json:"base_resp"`
}

type VoiceInfo struct {
	VoiceID   string `json:"voice_id"`
	VoiceName string `json:"voice_name"`
}

type DeleteVoiceRequest struct {
	VoiceType string `json:"voice_type"`
	VoiceID   string `json:"voice_id"`
}

type DeleteVoiceResponse struct {
	BaseResp BaseResp `json:"base_resp"`
}
