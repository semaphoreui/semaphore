package compliance

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/tz"
	log "github.com/sirupsen/logrus"
)

// ContentService handles SCAP content operations
type ContentService struct {
	store db.Store
}

// NewContentService creates a new content service
func NewContentService(store db.Store) *ContentService {
	return &ContentService{
		store: store,
	}
}

// ScapProfileInfo represents a profile discovered from a DataStream
type ScapProfileInfo struct {
	ID          string `xml:"id,attr" json:"id"`
	Title       string `xml:"title" json:"title"`
	Description string `xml:"description" json:"description"`
	Severity    string `xml:"severity,attr" json:"severity"`
}

// ScapContentInfo represents content information from oscap info
type ScapContentInfo struct {
	Profiles []ScapProfileInfo `xml:"profiles>profile" json:"profiles"`
}

// UploadContent uploads and processes a SCAP DataStream file
func (s *ContentService) UploadContent(projectID, userID int, name, source string, contentData []byte) (*db.ScapContent, []*db.ScapProfile, error) {
	// Create content directory for the project
	contentDir := filepath.Join("uploads", "scap", fmt.Sprintf("project_%d", projectID))
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create content directory: %v", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d.xml", strings.ReplaceAll(name, " ", "_"), tz.Now().Unix())
	filePath := filepath.Join(contentDir, filename)

	// Write content to file
	if err := os.WriteFile(filePath, contentData, 0644); err != nil {
		return nil, nil, fmt.Errorf("failed to write content file: %v", err)
	}

	// Validate the SCAP content and extract profiles
	profiles, err := s.extractProfiles(filePath)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, nil, fmt.Errorf("failed to extract profiles: %v", err)
	}

	// Create content record
	content := &db.ScapContent{
		ProjectID:  projectID,
		Name:       name,
		Source:     &source,
		DsXmlPath:  &filePath,
		UploadedBy: userID,
		Created:    tz.Now(),
	}

	// Save to database
	if err := s.store.CreateScapContent(content); err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, nil, fmt.Errorf("failed to save content: %v", err)
	}

	// Create profile records
	var dbProfiles []*db.ScapProfile
	for _, profile := range profiles {
		dbProfile := &db.ScapProfile{
			ContentID:   content.ID,
			ProfileID:   profile.ID,
			Title:       profile.Title,
			Description: &profile.Description,
			Severity:    &profile.Severity,
		}

		if err := s.store.CreateScapProfile(dbProfile); err != nil {
			log.Warnf("Failed to save profile %s: %v", profile.ID, err)
			continue
		}

		dbProfiles = append(dbProfiles, dbProfile)
	}

	return content, dbProfiles, nil
}

// extractProfiles uses oscap to extract profile information from a DataStream
func (s *ContentService) extractProfiles(filePath string) ([]ScapProfileInfo, error) {
	// Run oscap info to get profile information
	cmd := exec.Command("oscap", "info", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("oscap info failed: %v", err)
	}

	// Parse XML output
	var contentInfo ScapContentInfo
	if err := xml.Unmarshal(output, &contentInfo); err != nil {
		return nil, fmt.Errorf("failed to parse oscap output: %v", err)
	}

	return contentInfo.Profiles, nil
}

// GetContentsByProject retrieves all SCAP contents for a project
func (s *ContentService) GetContentsByProject(projectID int) ([]*db.ScapContent, error) {
	return s.store.GetScapContentsByProject(projectID)
}

// GetContent retrieves a specific SCAP content
func (s *ContentService) GetContent(id int) (*db.ScapContent, error) {
	return s.store.GetScapContent(id)
}

// GetProfilesByContent retrieves all profiles for a content
func (s *ContentService) GetProfilesByContent(contentID int) ([]*db.ScapProfile, error) {
	return s.store.GetScapProfilesByContent(contentID)
}

// DeleteContent deletes a SCAP content and its associated files
func (s *ContentService) DeleteContent(id int) error {
	// Get content to find file path
	content, err := s.store.GetScapContent(id)
	if err != nil {
		return fmt.Errorf("failed to get content: %v", err)
	}

	// Delete associated profiles first
	if err := s.store.DeleteScapProfilesByContent(id); err != nil {
		log.Warnf("Failed to delete profiles for content %d: %v", id, err)
	}

	// Delete the content record
	if err := s.store.DeleteScapContent(id); err != nil {
		return fmt.Errorf("failed to delete content: %v", err)
	}

	// Delete the file if it exists
	if content.DsXmlPath != nil {
		if err := os.Remove(*content.DsXmlPath); err != nil && !os.IsNotExist(err) {
			log.Warnf("Failed to delete content file %s: %v", *content.DsXmlPath, err)
		}
	}

	return nil
}

// ValidateOscapInstallation checks if oscap is available on the system
func (s *ContentService) ValidateOscapInstallation() error {
	cmd := exec.Command("oscap", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oscap is not installed or not in PATH: %v", err)
	}
	return nil
}

// GetContentFile returns the file path for a content
func (s *ContentService) GetContentFile(contentID int) (string, error) {
	content, err := s.store.GetScapContent(contentID)
	if err != nil {
		return "", err
	}

	if content.DsXmlPath == nil {
		return "", fmt.Errorf("content file path is not set")
	}

	return *content.DsXmlPath, nil
}
