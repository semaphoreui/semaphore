package api

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Digital-Data-Co/forge/api/helpers"
)

// FileSecurityConfig holds configuration for file security
type FileSecurityConfig struct {
	MaxFileSize    int64    // Maximum file size in bytes
	AllowedTypes   []string // Allowed MIME types
	AllowedExts    []string // Allowed file extensions
	ScanForMalware bool     // Whether to scan for malware (placeholder)
}

// DefaultFileSecurityConfig returns default security configuration
func DefaultFileSecurityConfig() FileSecurityConfig {
	return FileSecurityConfig{
		MaxFileSize: 32 * 1024 * 1024, // 32MB
		AllowedTypes: []string{
			"application/xml",
			"text/xml",
			"application/octet-stream", // For binary files
		},
		AllowedExts: []string{
			".xml",
			".arf",
		},
		ScanForMalware: false, // Would require integration with antivirus
	}
}

// ComplianceFileSecurityConfig returns security config for compliance files
func ComplianceFileSecurityConfig() FileSecurityConfig {
	return FileSecurityConfig{
		MaxFileSize: 32 * 1024 * 1024, // 32MB
		AllowedTypes: []string{
			"application/xml",
			"text/xml",
		},
		AllowedExts: []string{
			".xml",
		},
		ScanForMalware: false,
	}
}

// ValidateFileUpload validates a file upload against security rules
func ValidateFileUpload(file *multipart.FileHeader, config FileSecurityConfig) error {
	// Check file size
	if file.Size > config.MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.MaxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedExtension(ext, config.AllowedExts) {
		return fmt.Errorf("file extension '%s' is not allowed. Allowed extensions: %s",
			ext, strings.Join(config.AllowedExts, ", "))
	}

	// Check MIME type
	if !isAllowedMIMEType(file.Header.Get("Content-Type"), config.AllowedTypes) {
		return fmt.Errorf("file type '%s' is not allowed. Allowed types: %s",
			file.Header.Get("Content-Type"), strings.Join(config.AllowedTypes, ", "))
	}

	// Additional filename validation
	if err := validateFilename(file.Filename); err != nil {
		return err
	}

	return nil
}

// isAllowedExtension checks if file extension is allowed
func isAllowedExtension(ext string, allowedExts []string) bool {
	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

// isAllowedMIMEType checks if MIME type is allowed
func isAllowedMIMEType(mimeType string, allowedTypes []string) bool {
	// Handle multipart/form-data with boundary
	if strings.Contains(mimeType, "multipart/form-data") {
		return true
	}

	for _, allowed := range allowedTypes {
		if mimeType == allowed {
			return true
		}
	}
	return false
}

// validateFilename validates filename for security issues
func validateFilename(filename string) error {
	// Check for path traversal attempts
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("filename contains invalid characters")
	}

	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("filename contains null bytes")
	}

	// Check filename length
	if len(filename) > 255 {
		return fmt.Errorf("filename too long")
	}

	// Check for empty filename
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	return nil
}

// ScanFileForMalware performs basic file content scanning
func ScanFileForMalware(content []byte) error {
	// This is a placeholder for malware scanning
	// In a real implementation, you would integrate with:
	// - ClamAV
	// - VirusTotal API
	// - Commercial antivirus solutions

	// Basic checks for common malicious patterns
	maliciousPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"eval(",
		"document.cookie",
		"document.write",
	}

	contentStr := strings.ToLower(string(content))
	for _, pattern := range maliciousPatterns {
		if strings.Contains(contentStr, pattern) {
			return fmt.Errorf("file contains potentially malicious content: %s", pattern)
		}
	}

	return nil
}

// CalculateFileChecksum calculates MD5 checksum of file content
func CalculateFileChecksum(content []byte) string {
	hash := md5.Sum(content)
	return fmt.Sprintf("%x", hash)
}

// SecureFileUploadMiddleware creates middleware for secure file uploads
func SecureFileUploadMiddleware(config FileSecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to POST requests with file uploads
			if r.Method != "POST" {
				next.ServeHTTP(w, r)
				return
			}

			// Parse multipart form with size limit
			err := r.ParseMultipartForm(config.MaxFileSize)
			if err != nil {
				helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
					"error": "Failed to parse multipart form",
				})
				return
			}

			// Validate each uploaded file
			for _, files := range r.MultipartForm.File {
				for _, file := range files {
					if err := ValidateFileUpload(file, config); err != nil {
						helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
							"error": err.Error(),
						})
						return
					}

					// Optional: Scan file content for malware
					if config.ScanForMalware {
						fileHandle, err := file.Open()
						if err != nil {
							helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
								"error": "Failed to open uploaded file",
							})
							return
						}

						content, err := io.ReadAll(fileHandle)
						fileHandle.Close()
						if err != nil {
							helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
								"error": "Failed to read uploaded file",
							})
							return
						}

						if err := ScanFileForMalware(content); err != nil {
							helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
								"error": err.Error(),
							})
							return
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
