package minimax

// T2ARequest represents request for Text-to-Audio (Sync & Async)
type T2ARequest struct {
	Model             string         `json:"model"` // speech-01-turbo, speech-01-hd, speech-2.6-hd, speech-2.6-turbo
	Text              string         `json:"text"`
	TextFileID        int64          `json:"text_file_id,omitempty"`
	LanguageBoost     string         `json:"language_boost,omitempty"` // auto, Chinese, English, etc.
	VoiceSetting      VoiceSetting   `json:"voice_setting"`
	AudioSetting      AudioSetting   `json:"audio_setting"`
	PronunciationDict map[string]any `json:"pronunciation_dict,omitempty"`
	VoiceModify       VoiceModify    `json:"voice_modify,omitempty"`
	AigcWatermark     bool           `json:"aigc_watermark,omitempty"`
}

type VoiceModify struct {
	Pitch        int    `json:"pitch,omitempty"`
	Intensity    int    `json:"intensity,omitempty"`
	Timbre       int    `json:"timbre,omitempty"`
	SoundEffects string `json:"sound_effects,omitempty"` // spacious_echo, etc.
}

type VoiceSetting struct {
	VoiceID              string  `json:"voice_id"`
	Speed                float64 `json:"speed,omitempty"`
	Vol                  float64 `json:"vol,omitempty"`
	Pitch                int     `json:"pitch,omitempty"`
	Emotion              string  `json:"emotion,omitempty"` // happy, sad, etc.
	EnglishNormalization bool    `json:"english_normalization,omitempty"`
}

type AudioSetting struct {
	AudioSampleRate int64  `json:"audio_sample_rate,omitempty"` // 32000
	Bitrate         int64  `json:"bitrate,omitempty"`           // 128000
	Format          string `json:"format,omitempty"`            // mp3, wav
	Channel         int64  `json:"channel,omitempty"`           // 1, 2
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
	FileID                  int64        `json:"file_id"`  // Uploaded file ID
	VoiceID                 string       `json:"voice_id"` // Custom ID
	ClonePrompt             *ClonePrompt `json:"clone_prompt,omitempty"`
	Text                    string       `json:"text,omitempty"`                      // For demo generation, max 1000 chars
	Model                   string       `json:"model,omitempty"`                     // speech-2.6-hd, speech-2.6-turbo, etc.
	LanguageBoost           string       `json:"language_boost,omitempty"`            // auto, Chinese, English, etc.
	NeedNoiseReduction      bool         `json:"need_noise_reduction,omitempty"`      // Default false
	NeedVolumeNormalization bool         `json:"need_volume_normalization,omitempty"` // Default false
	AigcWatermark           bool         `json:"aigc_watermark,omitempty"`            // Default false
}

type ClonePrompt struct {
	PromptAudio int64  `json:"prompt_audio"`
	PromptText  string `json:"prompt_text"`
}

type VoiceCloneResponse struct {
	InputSensitive     bool     `json:"input_sensitive"`
	InputSensitiveType int      `json:"input_sensitive_type"`
	DemoAudio          string   `json:"demo_audio"` // URL or empty
	BaseResp           BaseResp `json:"base_resp"`
}

// Upload
type UploadResponse struct {
	File     UploadFileData `json:"file"`
	BaseResp BaseResp       `json:"base_resp"`
}

type UploadFileData struct {
	FileID    int64  `json:"file_id"`
	Bytes     int64  `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
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
