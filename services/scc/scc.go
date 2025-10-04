package scc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Digital-Data-Co/forge/db"
	log "github.com/sirupsen/logrus"
)

// SCCService handles integration with SCAP Compliance Checker
type SCCService struct {
	store         db.Store
	contentPath   string
	sccBinaryPath string
}

// NewSCCService creates a new SCC service
func NewSCCService(store db.Store, contentPath, sccBinaryPath string) *SCCService {
	return &SCCService{
		store:         store,
		contentPath:   contentPath,
		sccBinaryPath: sccBinaryPath,
	}
}

// SCCBenchmark represents a SCAP benchmark from NIWC Atlantic
type SCCBenchmark struct {
	Title           string    `json:"title"`
	Date            string    `json:"date"`
	STIGVersion     string    `json:"stig_version"`
	BenchmarkVersion string   `json:"benchmark_version"`
	DownloadURL     string    `json:"download_url"`
	FileName        string    `json:"file_name"`
	OS              string    `json:"os"`
	Type            string    `json:"type"` // "STIG", "CIS", etc.
	Size            int64     `json:"size"`
	LastModified    time.Time `json:"last_modified"`
}

// SCCScanResult represents the result of an SCC scan
type SCCScanResult struct {
	Host           string                 `json:"host"`
	Benchmark      string                 `json:"benchmark"`
	Profile        string                 `json:"profile"`
	Status         string                 `json:"status"` // "pass", "fail", "error"
	Score          float64                `json:"score"`
	TotalRules     int                    `json:"total_rules"`
	PassedRules    int                    `json:"passed_rules"`
	FailedRules    int                    `json:"failed_rules"`
	NotApplicable  int                    `json:"not_applicable"`
	NotChecked     int                    `json:"not_checked"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	Duration       time.Duration          `json:"duration"`
	ReportPath     string                 `json:"report_path"`
	Results        []SCCRuleResult        `json:"results"`
	Summary        map[string]interface{} `json:"summary"`
}

// SCCRuleResult represents an individual rule result
type SCCRuleResult struct {
	RuleID       string `json:"rule_id"`
	Title        string `json:"title"`
	Severity     string `json:"severity"`
	Status       string `json:"status"`
	Description  string `json:"description"`
	Rationale    string `json:"rationale"`
	Remediation  string `json:"remediation"`
	References   string `json:"references"`
	CheckContent string `json:"check_content"`
}

// GetAvailableBenchmarks returns the list of available SCAP benchmarks
func (s *SCCService) GetAvailableBenchmarks() ([]SCCBenchmark, error) {
	// This would typically fetch from the NIWC Atlantic repository
	// For now, we'll return a curated list of common benchmarks
	benchmarks := []SCCBenchmark{
		{
			Title:            "Red Hat Enterprise Linux 7 STIG SCAP Benchmark",
			Date:            "2025-09-16",
			STIGVersion:     "003.015",
			BenchmarkVersion: "003.015.13",
			DownloadURL:     "https://raw.githubusercontent.com/niwc-atlantic/scap-content-library/refs/heads/main/Current/U_RHEL_7_V3R15_STIG_SCAP_1-4_Benchmark-enhancedV13-signed.zip",
			FileName:        "U_RHEL_7_V3R15_STIG_SCAP_1-4_Benchmark-enhancedV13-signed.zip",
			OS:              "RHEL7",
			Type:            "STIG",
		},
		{
			Title:            "Red Hat Enterprise Linux 8 STIG SCAP Benchmark",
			Date:            "2025-09-16",
			STIGVersion:     "002.004",
			BenchmarkVersion: "002.004.16",
			DownloadURL:     "https://raw.githubusercontent.com/niwc-atlantic/scap-content-library/refs/heads/main/Current/U_RHEL_8_V2R4_STIG_SCAP_1-4_Benchmark-enhancedV16-signed.zip",
			FileName:        "U_RHEL_8_V2R4_STIG_SCAP_1-4_Benchmark-enhancedV16-signed.zip",
			OS:              "RHEL8",
			Type:            "STIG",
		},
		{
			Title:            "Red Hat Enterprise Linux 9 STIG SCAP Benchmark",
			Date:            "2025-09-16",
			STIGVersion:     "002.005",
			BenchmarkVersion: "002.005.7",
			DownloadURL:     "https://raw.githubusercontent.com/niwc-atlantic/scap-content-library/refs/heads/main/Current/U_RHEL_9_V2R5_STIG_SCAP_1-4_Benchmark-enhancedV7-signed.zip",
			FileName:        "U_RHEL_9_V2R5_STIG_SCAP_1-4_Benchmark-enhancedV7-signed.zip",
			OS:              "RHEL9",
			Type:            "STIG",
		},
		{
			Title:            "SUSE Linux Enterprise Server 15 STIG SCAP Benchmark",
			Date:            "2025-09-16",
			STIGVersion:     "002.005",
			BenchmarkVersion: "002.005.13",
			DownloadURL:     "https://raw.githubusercontent.com/niwc-atlantic/scap-content-library/refs/heads/main/Current/U_SLES_15_V2R5_STIG_SCAP_1-4_Benchmark-enhancedV13-signed.zip",
			FileName:        "U_SLES_15_V2R5_STIG_SCAP_1-4_Benchmark-enhancedV13-signed.zip",
			OS:              "SLES15",
			Type:            "STIG",
		},
		{
			Title:            "Apple macOS 13 (Ventura) STIG SCAP Benchmark",
			Date:            "2025-09-16",
			STIGVersion:     "001.005",
			BenchmarkVersion: "001.005.2",
			DownloadURL:     "https://raw.githubusercontent.com/niwc-atlantic/scap-content-library/refs/heads/main/Current/U_Apple_macOS_13_V1R5_STIG_SCAP_1-4_Benchmark-enhancedV2-signed.zip",
			FileName:        "U_Apple_macOS_13_V1R5_STIG_SCAP_1-4_Benchmark-enhancedV2-signed.zip",
			OS:              "macOS13",
			Type:            "STIG",
		},
		{
			Title:            "Apple macOS 14 (Sonoma) STIG SCAP Benchmark",
			Date:            "2025-09-16",
			STIGVersion:     "002.003",
			BenchmarkVersion: "002.003.2",
			DownloadURL:     "https://raw.githubusercontent.com/niwc-atlantic/scap-content-library/refs/heads/main/Current/U_Apple_macOS_14_V2R3_STIG_SCAP_1-4_Benchmark-enhancedV2-signed.zip",
			FileName:        "U_Apple_macOS_14_V2R3_STIG_SCAP_1-4_Benchmark-enhancedV2-signed.zip",
			OS:              "macOS14",
			Type:            "STIG",
		},
		{
			Title:            "Apple macOS 15 (Sequoia) STIG SCAP Benchmark",
			Date:            "2025-09-16",
			STIGVersion:     "001.004",
			BenchmarkVersion: "001.004.2",
			DownloadURL:     "https://raw.githubusercontent.com/niwc-atlantic/scap-content-library/refs/heads/main/Current/U_Apple_macOS_15_V1R4_STIG_SCAP_1-4_Benchmark-enhancedV2-signed.zip",
			FileName:        "U_Apple_macOS_15_V1R4_STIG_SCAP_1-4_Benchmark-enhancedV2-signed.zip",
			OS:              "macOS15",
			Type:            "STIG",
		},
	}

	return benchmarks, nil
}

// DownloadBenchmark downloads a SCAP benchmark from the NIWC Atlantic repository
func (s *SCCService) DownloadBenchmark(benchmark SCCBenchmark) error {
	// Create content directory if it doesn't exist
	if err := os.MkdirAll(s.contentPath, 0755); err != nil {
		return fmt.Errorf("failed to create content directory: %w", err)
	}

	// Download the benchmark file
	resp, err := http.Get(benchmark.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download benchmark: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download benchmark, status: %d", resp.StatusCode)
	}

	// Save the file
	filePath := filepath.Join(s.contentPath, benchmark.FileName)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read benchmark data: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save benchmark file: %w", err)
	}

	log.Infof("Downloaded SCAP benchmark: %s", benchmark.Title)
	return nil
}

// ExtractBenchmark extracts a downloaded benchmark ZIP file
func (s *SCCService) ExtractBenchmark(benchmark SCCBenchmark) (string, error) {
	filePath := filepath.Join(s.contentPath, benchmark.FileName)
	extractPath := filepath.Join(s.contentPath, strings.TrimSuffix(benchmark.FileName, ".zip"))

	// Use unzip command to extract
	cmd := exec.Command("unzip", "-o", filePath, "-d", extractPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to extract benchmark: %w, output: %s", err, string(output))
	}

	log.Infof("Extracted SCAP benchmark to: %s", extractPath)
	return extractPath, nil
}

// GetBenchmarkProfiles retrieves available profiles from a SCAP benchmark
func (s *SCCService) GetBenchmarkProfiles(benchmarkPath string) ([]string, error) {
	if s.sccBinaryPath == "" {
		// Fallback to oscap if SCC is not available
		return s.getProfilesWithOscap(benchmarkPath)
	}

	// Use SCC to get profiles
	cmd := exec.Command(s.sccBinaryPath, "--list-profiles", benchmarkPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles with SCC: %w, output: %s", err, string(output))
	}

	// Parse SCC output to extract profile names
	lines := strings.Split(string(output), "\n")
	var profiles []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Profile") {
			profiles = append(profiles, line)
		}
	}

	return profiles, nil
}

// getProfilesWithOscap gets profiles using oscap command
func (s *SCCService) getProfilesWithOscap(benchmarkPath string) ([]string, error) {
	// Find the XCCDF file
	xccdfFiles, err := filepath.Glob(filepath.Join(benchmarkPath, "*.xml"))
	if err != nil {
		return nil, fmt.Errorf("failed to find XCCDF files: %w", err)
	}

	if len(xccdfFiles) == 0 {
		return nil, fmt.Errorf("no XCCDF files found in benchmark")
	}

	// Use oscap to get profiles
	cmd := exec.Command("oscap", "info", xccdfFiles[0])
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles with oscap: %w, output: %s", err, string(output))
	}

	// Parse oscap output to extract profile names
	lines := strings.Split(string(output), "\n")
	var profiles []string
	inProfiles := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Profiles:") {
			inProfiles = true
			continue
		}
		if inProfiles && strings.HasPrefix(line, "Id:") {
			// Extract profile ID
			parts := strings.Split(line, "Id:")
			if len(parts) > 1 {
				profileID := strings.TrimSpace(parts[1])
				if profileID != "" {
					profiles = append(profiles, profileID)
				}
			}
		}
		if inProfiles && line == "" {
			break
		}
	}

	return profiles, nil
}

// RunSCCScan executes an SCC scan on a target host
func (s *SCCService) RunSCCScan(ctx context.Context, host string, benchmarkPath, profile string) (*SCCScanResult, error) {
	startTime := time.Now()
	
	// Create output directory
	outputDir := filepath.Join(s.contentPath, "reports", fmt.Sprintf("scan_%s_%d", host, startTime.Unix()))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	result := &SCCScanResult{
		Host:      host,
		Benchmark: benchmarkPath,
		Profile:   profile,
		StartTime: startTime,
		ReportPath: outputDir,
	}

	// Determine if we should use SCC or oscap
	var cmd *exec.Cmd
	if s.sccBinaryPath != "" {
		// Use SCC
		cmd = exec.CommandContext(ctx, s.sccBinaryPath,
			"--benchmark", benchmarkPath,
			"--profile", profile,
			"--target", host,
			"--output-dir", outputDir,
			"--format", "json")
	} else {
		// Use oscap
		xccdfFiles, err := filepath.Glob(filepath.Join(benchmarkPath, "*.xml"))
		if err != nil || len(xccdfFiles) == 0 {
			return nil, fmt.Errorf("failed to find XCCDF files in benchmark")
		}

		resultsFile := filepath.Join(outputDir, "results.xml")
		reportFile := filepath.Join(outputDir, "report.html")
		
		cmd = exec.CommandContext(ctx, "oscap", "xccdf", "eval",
			"--profile", profile,
			"--results", resultsFile,
			"--report", reportFile,
			xccdfFiles[0])
	}

	// Execute the scan
	output, err := cmd.CombinedOutput()
	endTime := time.Now()
	result.EndTime = endTime
	result.Duration = endTime.Sub(startTime)

	if err != nil {
		result.Status = "error"
		log.Errorf("SCC scan failed: %v, output: %s", err, string(output))
		return result, fmt.Errorf("SCC scan failed: %w", err)
	}

	// Parse results
	if err := s.parseScanResults(result, outputDir); err != nil {
		log.Errorf("Failed to parse scan results: %v", err)
		result.Status = "error"
		return result, nil // Return partial results
	}

	result.Status = "completed"
	log.Infof("SCC scan completed for host %s with profile %s", host, profile)
	return result, nil
}

// parseScanResults parses the scan results from output files
func (s *SCCService) parseScanResults(result *SCCScanResult, outputDir string) error {
	// Try to parse results.xml first (oscap format)
	resultsFile := filepath.Join(outputDir, "results.xml")
	if _, err := os.Stat(resultsFile); err == nil {
		return s.parseOscapResults(result, resultsFile)
	}

	// Try to parse JSON results (SCC format)
	jsonFile := filepath.Join(outputDir, "results.json")
	if _, err := os.Stat(jsonFile); err == nil {
		return s.parseSCCResults(result, jsonFile)
	}

	// If no results file found, create a basic result
	result.Status = "completed"
	result.Score = 0.0
	result.TotalRules = 0
	result.PassedRules = 0
	result.FailedRules = 0

	return nil
}

// parseOscapResults parses oscap XML results
func (s *SCCService) parseOscapResults(result *SCCScanResult, resultsFile string) error {
	// This is a simplified parser - in a real implementation,
	// you would use an XML parser to extract detailed results
	data, err := ioutil.ReadFile(resultsFile)
	if err != nil {
		return fmt.Errorf("failed to read results file: %w", err)
	}

	// Basic parsing - count rule results
	content := string(data)
	result.TotalRules = strings.Count(content, "<rule-result")
	result.PassedRules = strings.Count(content, `result="pass"`)
	result.FailedRules = strings.Count(content, `result="fail"`)
	result.NotApplicable = strings.Count(content, `result="notapplicable"`)
	result.NotChecked = strings.Count(content, `result="notchecked"`)

	if result.TotalRules > 0 {
		result.Score = float64(result.PassedRules) / float64(result.TotalRules) * 100
	}

	return nil
}

// parseSCCResults parses SCC JSON results
func (s *SCCService) parseSCCResults(result *SCCScanResult, jsonFile string) error {
	// This would parse the JSON output from SCC
	// For now, we'll create a placeholder
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to read SCC results file: %w", err)
	}

	// Basic parsing - you would use json.Unmarshal in a real implementation
	content := string(data)
	result.TotalRules = strings.Count(content, "\"rule_id\"")
	result.PassedRules = strings.Count(content, "\"result\":\"pass\"")
	result.FailedRules = strings.Count(content, "\"result\":\"fail\"")

	if result.TotalRules > 0 {
		result.Score = float64(result.PassedRules) / float64(result.TotalRules) * 100
	}

	return nil
}

// CheckSCCAvailability checks if SCC or oscap is available on the system
func (s *SCCService) CheckSCCAvailability() (bool, string, error) {
	// Check for SCC first
	if s.sccBinaryPath != "" {
		cmd := exec.Command(s.sccBinaryPath, "--version")
		if output, err := cmd.CombinedOutput(); err == nil {
			return true, fmt.Sprintf("SCC: %s", string(output)), nil
		}
	}

	// Check for oscap
	cmd := exec.Command("oscap", "--version")
	if output, err := cmd.CombinedOutput(); err == nil {
		return true, fmt.Sprintf("oscap: %s", string(output)), nil
	}

	return false, "", fmt.Errorf("neither SCC nor oscap is available")
}

// GetSupportedOS returns the list of operating systems with available SCAP benchmarks
func (s *SCCService) GetSupportedOS() []string {
	return []string{
		"RHEL7", "RHEL8", "RHEL9",
		"SLES15",
		"macOS13", "macOS14", "macOS15",
		"Oracle7", "Oracle8",
		"Solaris11_SPARC", "Solaris11_X86",
		"TOSS4",
	}
}

// GetBenchmarksByOS returns available benchmarks for a specific OS
func (s *SCCService) GetBenchmarksByOS(os string) ([]SCCBenchmark, error) {
	allBenchmarks, err := s.GetAvailableBenchmarks()
	if err != nil {
		return nil, err
	}

	var filtered []SCCBenchmark
	for _, benchmark := range allBenchmarks {
		if benchmark.OS == os {
			filtered = append(filtered, benchmark)
		}
	}

	return filtered, nil
}
