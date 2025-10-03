package lockdown

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// LockdownRepository represents an Ansible Lockdown repository
type LockdownRepository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	Language    string `json:"language"`
	Stars       int    `json:"stargazers_count"`
	UpdatedAt   string `json:"updated_at"`
}

// LockdownRole represents a specific compliance role within a repository
type LockdownRole struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OS          string `json:"os"`
	Version     string `json:"version"`
	Framework   string `json:"framework"` // CIS or STIG
	Repository  string `json:"repository"`
	Tasks       []LockdownTask `json:"tasks,omitempty"`
}

// LockdownTask represents an Ansible task from a Lockdown role
type LockdownTask struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Module      string            `json:"module"`
	Args        map[string]interface{} `json:"args"`
	Tags        []string          `json:"tags"`
	When        string            `json:"when,omitempty"`
}

// LockdownService handles interactions with Ansible Lockdown repositories
type LockdownService struct {
	httpClient *http.Client
	baseURL    string
}

// NewLockdownService creates a new LockdownService instance
func NewLockdownService() *LockdownService {
	return &LockdownService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// GetAvailableRepositories returns all available Ansible Lockdown repositories
func (s *LockdownService) GetAvailableRepositories(ctx context.Context) ([]LockdownRepository, error) {
	url := fmt.Sprintf("%s/orgs/ansible-lockdown/repos?per_page=100&sort=updated", s.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var repositories []LockdownRepository
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Filter for compliance repositories
	var complianceRepos []LockdownRepository
	for _, repo := range repositories {
		if s.isComplianceRepository(repo) {
			complianceRepos = append(complianceRepos, repo)
		}
	}
	
	return complianceRepos, nil
}

// GetComplianceRoles returns available compliance roles for a specific OS and framework
func (s *LockdownService) GetComplianceRoles(ctx context.Context, os, framework string) ([]LockdownRole, error) {
	repositories, err := s.GetAvailableRepositories(ctx)
	if err != nil {
		return nil, err
	}
	
	var roles []LockdownRole
	for _, repo := range repositories {
		if s.matchesCriteria(repo, os, framework) {
			role := LockdownRole{
				Name:        s.extractRoleName(repo.Name),
				Description: repo.Description,
				OS:          s.extractOS(repo.Name),
				Version:     s.extractVersion(repo.Name),
				Framework:   s.extractFramework(repo.Name),
				Repository:  repo.FullName,
			}
			roles = append(roles, role)
		}
	}
	
	return roles, nil
}

// GetRoleTasks fetches and parses Ansible tasks from a specific role
func (s *LockdownService) GetRoleTasks(ctx context.Context, repository, rolePath string) ([]LockdownTask, error) {
	// This would typically clone the repository and parse the tasks/main.yml file
	// For now, we'll return a placeholder implementation
	log.Warnf("GetRoleTasks not fully implemented for %s/%s", repository, rolePath)
	
	// Placeholder tasks based on common Ansible Lockdown patterns
	tasks := []LockdownTask{
		{
			Name:        "Install required packages",
			Description: "Ensure required packages are installed for compliance",
			Module:      "package",
			Args: map[string]interface{}{
				"name": "{{ compliance_packages }}",
				"state": "present",
			},
			Tags: []string{"compliance", "packages"},
		},
		{
			Name:        "Configure system settings",
			Description: "Apply compliance configuration settings",
			Module:      "lineinfile",
			Args: map[string]interface{}{
				"path": "/etc/security/limits.conf",
				"line": "* hard core 0",
			},
			Tags: []string{"compliance", "system"},
		},
		{
			Name:        "Set file permissions",
			Description: "Ensure proper file permissions for compliance",
			Module:      "file",
			Args: map[string]interface{}{
				"path": "/etc/passwd",
				"mode": "0644",
			},
			Tags: []string{"compliance", "permissions"},
		},
	}
	
	return tasks, nil
}

// isComplianceRepository checks if a repository is a compliance-related repository
func (s *LockdownService) isComplianceRepository(repo LockdownRepository) bool {
	name := strings.ToUpper(repo.Name)
	return strings.Contains(name, "CIS") || strings.Contains(name, "STIG")
}

// matchesCriteria checks if a repository matches the specified OS and framework
func (s *LockdownService) matchesCriteria(repo LockdownRepository, os, framework string) bool {
	name := strings.ToUpper(repo.Name)
	
	// Check framework
	hasFramework := false
	if framework == "CIS" && strings.Contains(name, "CIS") {
		hasFramework = true
	} else if framework == "STIG" && strings.Contains(name, "STIG") {
		hasFramework = true
	}
	
	if !hasFramework {
		return false
	}
	
	// Check OS (case-insensitive)
	osUpper := strings.ToUpper(os)
	return strings.Contains(name, osUpper)
}

// extractRoleName extracts a readable role name from repository name
func (s *LockdownService) extractRoleName(repoName string) string {
	// Convert "RHEL8-CIS" to "RHEL 8 CIS"
	parts := strings.Split(repoName, "-")
	if len(parts) >= 2 {
		return fmt.Sprintf("%s %s", parts[0], parts[1])
	}
	return repoName
}

// extractOS extracts the OS from repository name
func (s *LockdownService) extractOS(repoName string) string {
	// Extract OS from repository name like "RHEL8-CIS" -> "RHEL8"
	parts := strings.Split(repoName, "-")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "Unknown"
}

// extractVersion extracts the version from repository name
func (s *LockdownService) extractVersion(repoName string) string {
	// Extract version from repository name like "RHEL8-CIS" -> "8"
	parts := strings.Split(repoName, "-")
	if len(parts) >= 1 {
		os := parts[0]
		// Extract numeric version
		for i, char := range os {
			if char >= '0' && char <= '9' {
				return os[i:]
			}
		}
	}
	return "Unknown"
}

// extractFramework extracts the framework from repository name
func (s *LockdownService) extractFramework(repoName string) string {
	name := strings.ToUpper(repoName)
	if strings.Contains(name, "CIS") {
		return "CIS"
	} else if strings.Contains(name, "STIG") {
		return "STIG"
	}
	return "Unknown"
}

// GetSupportedOS returns list of supported operating systems
func (s *LockdownService) GetSupportedOS() []string {
	return []string{
		"RHEL7", "RHEL8", "RHEL9",
		"UBUNTU18", "UBUNTU20", "UBUNTU22", "UBUNTU24",
		"DEBIAN11", "DEBIAN12",
		"AMAZON2", "AMAZON2023",
		"SUSE15",
		"WINDOWS10", "WINDOWS11", "WINDOWS2016", "WINDOWS2019", "WINDOWS2022",
	}
}

// GetSupportedFrameworks returns list of supported compliance frameworks
func (s *LockdownService) GetSupportedFrameworks() []string {
	return []string{"CIS", "STIG"}
}
