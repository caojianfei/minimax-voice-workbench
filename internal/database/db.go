package database

import (
	"io"
	"log"
	"minimax-voice-workbench/internal/model"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dataDir string) {
	if dataDir == "" {
		dataDir = "."
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	dbPath := filepath.Join(dataDir, "minimax.db")

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Perform Custom Migration (Move files and consolidate fields)
	// migrateVoiceStorage(DB)

	// Auto Migrate
	err = DB.AutoMigrate(&model.ApiKey{}, &model.Voice{}, &model.SynthesisTask{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database initialized successfully at", dbPath)
}

func migrateVoiceStorage(db *gorm.DB) {
	// Ensure new directory exists
	newDir := "generated/voices"
	if err := os.MkdirAll(newDir, 0755); err != nil {
		log.Printf("Migration Warning: Failed to create %s: %v", newDir, err)
	}

	// Define a struct to read old data including potentially dropped columns
	// We use a map to be flexible or raw SQL
	type OldVoice struct {
		ID        uint
		VoiceID   string
		Preview   string
		DemoAudio string
	}

	// Check if demo_audio column exists
	hasDemoAudio := db.Migrator().HasColumn(&model.Voice{}, "demo_audio")

	var voices []OldVoice
	// Select fields. If demo_audio doesn't exist, we just select NULL as demo_audio
	query := db.Table("voices").Select("id, voice_id, preview")
	if hasDemoAudio {
		query = db.Table("voices").Select("id, voice_id, preview, demo_audio")
	}

	if err := query.Scan(&voices).Error; err != nil {
		log.Printf("Migration Warning: Failed to fetch voices: %v", err)
		return
	}

	for _, v := range voices {
		// Determine source path and file
		currentPath := v.Preview
		if currentPath == "" && v.DemoAudio != "" {
			currentPath = v.DemoAudio
		}

		if currentPath == "" {
			continue
		}

		// Normalize path
		// Stored in DB as "/files/filename.mp3" or similar
		// We want to move it to "generated/voices/filename.mp3"
		// And update DB to "/files/voices/filename.mp3"

		if strings.HasPrefix(currentPath, "/files/voices/") {
			// Already migrated
			continue
		}

		filename := filepath.Base(currentPath)
		// Assuming /files/ maps to ./generated/
		// So /files/demo_123.mp3 -> ./generated/demo_123.mp3

		srcPath := filepath.Join("generated", filename)
		// Handle case where path might be relative or different
		if !strings.HasPrefix(currentPath, "/files/") {
			// Maybe it was stored as local path? Assume relative to project root if not /files/
			srcPath = currentPath
		}

		// Check if file exists at srcPath
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			// Try checking if it's already in newDir (manual move?)
			if _, err := os.Stat(filepath.Join(newDir, filename)); err == nil {
				// File exists in new dir, just update DB
				updateDBPath(db, v.ID, filename)
			}
			continue
		}

		// Move file
		dstPath := filepath.Join(newDir, filename)
		if err := moveFile(srcPath, dstPath); err != nil {
			log.Printf("Migration Error: Failed to move %s to %s: %v", srcPath, dstPath, err)
			continue
		}

		// Update DB
		updateDBPath(db, v.ID, filename)
	}
}

func moveFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	in.Close()
	return os.Remove(src)
}

func updateDBPath(db *gorm.DB, id uint, filename string) {
	newWebPath := "/files/voices/" + filename
	// Update preview and clear demo_audio (if it exists)
	// We use map to avoid struct issues
	updates := map[string]interface{}{
		"preview": newWebPath,
	}
	if db.Migrator().HasColumn(&model.Voice{}, "demo_audio") {
		updates["demo_audio"] = nil // or "" depending on schema, usually nil for string pointer or "" for string
	}

	db.Table("voices").Where("id = ?", id).Updates(updates)
}
