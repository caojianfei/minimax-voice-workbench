package model

import (
	"time"

	"gorm.io/gorm"
)

// ApiKey stores API keys for Minimax platform
type ApiKey struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Platform  string         `gorm:"size:50;default:'minimax'" json:"platform"`
	Key       string         `gorm:"size:255;not null" json:"key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Voice represents a voice profile (cloned or official)
type Voice struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	VoiceID   string         `gorm:"size:100;uniqueIndex;not null" json:"voice_id"` // Minimax Voice ID
	Type      string         `gorm:"size:20;default:'cloned'" json:"type"`          // cloned, system, generated
	Preview   string         `gorm:"size:255" json:"preview"`                       // Path to preview audio (for design)
	DemoAudio string         `gorm:"size:255" json:"demo_audio"`                    // Path to demo audio (for clone)
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// SynthesisTask tracks text-to-speech tasks
type SynthesisTask struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TaskID    int64          `gorm:"index" json:"task_id"` // For async tasks
	Text      string         `gorm:"type:text" json:"text"`
	VoiceID   string         `gorm:"size:100" json:"voice_id"`
	Output    string         `gorm:"size:255" json:"output"` // Path to generated audio
	Status    string         `gorm:"size:20;default:'pending'" json:"status"`
	Error     string         `gorm:"size:255" json:"error,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
