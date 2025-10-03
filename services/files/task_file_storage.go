package files

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/util"
)

// TaskFileStorage handles file storage for task artifacts
type TaskFileStorage struct {
	basePath string
}

// NewTaskFileStorage creates a new TaskFileStorage instance
func NewTaskFileStorage() *TaskFileStorage {
	basePath := filepath.Join(util.Config.TmpPath, "task_files")
	return &TaskFileStorage{
		basePath: basePath,
	}
}

// StoreFile stores a file from the temporary directory to permanent storage
func (s *TaskFileStorage) StoreFile(projectID, taskID int, originalPath, filename string) (*db.TaskFile, error) {
	// Create the storage directory structure
	storageDir := filepath.Join(s.basePath, fmt.Sprintf("project_%d", projectID), fmt.Sprintf("task_%d", taskID))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Read the original file
	fileData, err := os.ReadFile(originalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read original file: %w", err)
	}

	// Generate a unique filename to avoid conflicts
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)
	uniqueFilename := fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
	
	storagePath := filepath.Join(storageDir, uniqueFilename)

	// Write the file to permanent storage
	if err := os.WriteFile(storagePath, fileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file to storage: %w", err)
	}

	// Calculate file checksum
	hash := md5.Sum(fileData)
	checksum := hex.EncodeToString(hash[:])

	// Determine MIME type
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Create TaskFile record
	taskFile := &db.TaskFile{
		TaskID:      taskID,
		ProjectID:   projectID,
		Filename:    uniqueFilename,
		OriginalPath: originalPath,
		FileSize:    int64(len(fileData)),
		MimeType:    mimeType,
		Checksum:    checksum,
		Created:     time.Now(),
	}

	return taskFile, nil
}

// GetFilePath returns the full path to a stored task file
func (s *TaskFileStorage) GetFilePath(taskFile *db.TaskFile) string {
	return filepath.Join(s.basePath, fmt.Sprintf("project_%d", taskFile.ProjectID), fmt.Sprintf("task_%d", taskFile.TaskID), taskFile.Filename)
}

// ReadFile reads a task file from storage
func (s *TaskFileStorage) ReadFile(taskFile *db.TaskFile) ([]byte, error) {
	filePath := s.GetFilePath(taskFile)
	return os.ReadFile(filePath)
}

// DeleteFile deletes a task file from storage
func (s *TaskFileStorage) DeleteFile(taskFile *db.TaskFile) error {
	filePath := s.GetFilePath(taskFile)
	return os.Remove(filePath)
}

// DeleteTaskFiles deletes all files for a specific task
func (s *TaskFileStorage) DeleteTaskFiles(projectID, taskID int) error {
	taskDir := filepath.Join(s.basePath, fmt.Sprintf("project_%d", projectID), fmt.Sprintf("task_%d", taskID))
	return os.RemoveAll(taskDir)
}

// GetFileReader returns a reader for a task file
func (s *TaskFileStorage) GetFileReader(taskFile *db.TaskFile) (io.ReadCloser, error) {
	filePath := s.GetFilePath(taskFile)
	return os.Open(filePath)
}
