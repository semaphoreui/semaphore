package db_lib

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/semaphoreui/semaphore/util"
)

type ProxyGitClient struct {
	keyInstaller AccessKeyInstaller
}

type repositoryRequest struct {
	GitURL    string `json:"git_url"`
	GitBranch string `json:"git_branch"`
	SSHKeyID  *int   `json:"ssh_key_id,omitempty"`
}

type repositoryResponse struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Archive []byte `json:"archive"`
}

func (c ProxyGitClient) Clone(r GitRepository) error {
	r.Logger.Log("Requesting Repository from server: " + r.Repository.GitURL)

	// Create request payload
	req := repositoryRequest{
		GitURL:    r.Repository.GitURL,
		GitBranch: r.Repository.GitBranch,
	}
	if r.Repository.SSHKeyID != 0 {
		req.SSHKeyID = &r.Repository.SSHKeyID
	}

	// Request repository from server
	archive, err := c.requestRepository(req)
	if err != nil {
		r.Logger.Log("Unable to request repository from server: " + err.Error())
		return err
	}

	// Create target directory
	targetPath := r.GetFullPath()
	err = os.MkdirAll(targetPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract archive to target directory
	err = c.extractArchive(archive.Archive, targetPath)
	if err != nil {
		r.Logger.Log("Unable to extract repository archive: " + err.Error())
		return err
	}

	r.Logger.Log("Repository extracted successfully")
	return nil
}

func (c ProxyGitClient) Pull(r GitRepository) error {
	r.Logger.Log("Updating Repository via server: " + r.Repository.GitURL)
	
	// For proxy mode, we'll just re-clone since the server provides fresh archives
	// This simplifies the implementation and ensures we always have the latest version
	err := os.RemoveAll(r.GetFullPath())
	if err != nil {
		return fmt.Errorf("failed to remove existing repository: %w", err)
	}
	
	return c.Clone(r)
}

func (c ProxyGitClient) Checkout(r GitRepository, target string) error {
	// For proxy mode, checkout is not supported as we receive the repository at the correct commit
	// The server should provide the repository at the requested commit
	r.Logger.Log("Checkout to " + target + " - using repository as provided by server")
	return nil
}

func (c ProxyGitClient) CanBePulled(r GitRepository) bool {
	// In proxy mode, we can always "pull" by requesting a fresh copy from the server
	return true
}

func (c ProxyGitClient) GetLastCommitMessage(r GitRepository) (string, error) {
	// Try to read from .git/COMMIT_EDITMSG or use git command if available
	gitDir := filepath.Join(r.GetFullPath(), ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "Repository cloned via proxy", nil
	}
	
	// Fallback to using cmd git client for local operations
	cmdClient := CmdGitClient{keyInstaller: c.keyInstaller}
	return cmdClient.GetLastCommitMessage(r)
}

func (c ProxyGitClient) GetLastCommitHash(r GitRepository) (string, error) {
	// Try to read from local git info or use git command if available
	gitDir := filepath.Join(r.GetFullPath(), ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "unknown", nil
	}
	
	// Fallback to using cmd git client for local operations
	cmdClient := CmdGitClient{keyInstaller: c.keyInstaller}
	return cmdClient.GetLastCommitHash(r)
}

func (c ProxyGitClient) GetLastRemoteCommitHash(r GitRepository) (string, error) {
	// For proxy mode, we can't check remote directly, so return local hash
	return c.GetLastCommitHash(r)
}

func (c ProxyGitClient) GetRemoteBranches(r GitRepository) ([]string, error) {
	// For proxy mode, return current branch since we can't query remote
	return []string{r.Repository.GitBranch}, nil
}

func (c ProxyGitClient) requestRepository(req repositoryRequest) (*repositoryResponse, error) {
	// Marshal request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := util.Config.WebHost + "/api/internal/repositories/archive"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if util.Config.Runner != nil && util.Config.Runner.Token != "" {
		httpReq.Header.Set("X-Runner-Token", util.Config.Runner.Token)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response repositoryResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c ProxyGitClient) extractArchive(archiveData []byte, targetPath string) error {
	// Create gzip reader
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip if this is not a regular file or directory
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			continue
		}

		// Clean the path to prevent path traversal
		cleanPath := filepath.Clean(header.Name)
		if strings.Contains(cleanPath, "..") {
			continue
		}

		targetFilePath := filepath.Join(targetPath, cleanPath)

		// Create directory if it's a directory entry
		if header.Typeflag == tar.TypeDir {
			err = os.MkdirAll(targetFilePath, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetFilePath, err)
			}
			continue
		}

		// Create parent directories if they don't exist
		err = os.MkdirAll(filepath.Dir(targetFilePath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", targetFilePath, err)
		}

		// Create and write file
		file, err := os.OpenFile(targetFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetFilePath, err)
		}

		_, err = io.Copy(file, tarReader)
		file.Close()
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetFilePath, err)
		}
	}

	return nil
}