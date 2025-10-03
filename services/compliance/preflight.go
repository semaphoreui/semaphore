package compliance

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// PreflightService handles preflight checks for OpenSCAP installation
type PreflightService struct{}

// NewPreflightService creates a new preflight service
func NewPreflightService() *PreflightService {
	return &PreflightService{}
}

// PreflightCheck represents the result of a preflight check
type PreflightCheck struct {
	OscapAvailable bool     `json:"oscap_available"`
	OscapVersion   string   `json:"oscap_version"`
	Errors         []string `json:"errors"`
	Warnings       []string `json:"warnings"`
	Info           []string `json:"info"`
}

// CheckOpenScapInstallation performs a comprehensive check of OpenSCAP installation
func (p *PreflightService) CheckOpenScapInstallation() *PreflightCheck {
	check := &PreflightCheck{
		Errors:   []string{},
		Warnings: []string{},
		Info:     []string{},
	}

	// Check if oscap command is available
	if err := p.checkOscapCommand(); err != nil {
		check.Errors = append(check.Errors, fmt.Sprintf("OpenSCAP not found: %v", err))
		return check
	}

	check.OscapAvailable = true

	// Get oscap version
	version, err := p.getOscapVersion()
	if err != nil {
		check.Warnings = append(check.Warnings, fmt.Sprintf("Could not determine oscap version: %v", err))
	} else {
		check.OscapVersion = version
		check.Info = append(check.Info, fmt.Sprintf("OpenSCAP version: %s", version))
	}

	// Check basic functionality
	if err := p.checkOscapBasicFunctionality(); err != nil {
		check.Errors = append(check.Errors, fmt.Sprintf("OpenSCAP basic functionality check failed: %v", err))
	}

	// Check for common SCAP content
	p.checkScapContent(check)

	// Check system compatibility
	p.checkSystemCompatibility(check)

	return check
}

// checkOscapCommand checks if the oscap command is available in PATH
func (p *PreflightService) checkOscapCommand() error {
	cmd := exec.Command("which", "oscap")
	if err := cmd.Run(); err != nil {
		// Try alternative command on Windows
		if runtime.GOOS == "windows" {
			cmd = exec.Command("where", "oscap")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("oscap command not found in PATH")
			}
		} else {
			return fmt.Errorf("oscap command not found in PATH")
		}
	}
	return nil
}

// getOscapVersion gets the OpenSCAP version
func (p *PreflightService) getOscapVersion() (string, error) {
	cmd := exec.Command("oscap", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// checkOscapBasicFunctionality checks if oscap can perform basic operations
func (p *PreflightService) checkOscapBasicFunctionality() error {
	// Test oscap info command
	cmd := exec.Command("oscap", "info", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oscap info command failed: %v", err)
	}

	// Test oscap xccdf command
	cmd = exec.Command("oscap", "xccdf", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oscap xccdf command failed: %v", err)
	}

	return nil
}

// checkScapContent checks for common SCAP content availability
func (p *PreflightService) checkScapContent(check *PreflightCheck) {
	// Check for SCAP Security Guide
	locations := []string{
		"/usr/share/xml/scap/ssg",
		"/usr/share/scap-security-guide",
		"/usr/local/share/scap-security-guide",
	}

	found := false
	for _, location := range locations {
		cmd := exec.Command("test", "-d", location)
		if err := cmd.Run(); err == nil {
			found = true
			check.Info = append(check.Info, fmt.Sprintf("SCAP Security Guide found at: %s", location))
			break
		}
	}

	if !found {
		check.Warnings = append(check.Warnings, "SCAP Security Guide not found in common locations")
		check.Info = append(check.Info, "Consider installing scap-security-guide package")
	}

	// Check for common SCAP content files
	commonFiles := []string{
		"/usr/share/xml/scap/ssg/content/ssg-rhel8-ds.xml",
		"/usr/share/xml/scap/ssg/content/ssg-ubuntu1804-ds.xml",
		"/usr/share/xml/scap/ssg/content/ssg-debian9-ds.xml",
	}

	for _, file := range commonFiles {
		cmd := exec.Command("test", "-f", file)
		if err := cmd.Run(); err == nil {
			check.Info = append(check.Info, fmt.Sprintf("SCAP content file found: %s", file))
		}
	}
}

// checkSystemCompatibility checks system compatibility with OpenSCAP
func (p *PreflightService) checkSystemCompatibility(check *PreflightCheck) {
	// Check operating system
	os := runtime.GOOS
	check.Info = append(check.Info, fmt.Sprintf("Operating system: %s", os))

	// Check architecture
	arch := runtime.GOARCH
	check.Info = append(check.Info, fmt.Sprintf("Architecture: %s", arch))

	// OS-specific checks
	switch os {
	case "linux":
		// Check for common Linux distributions
		p.checkLinuxDistribution(check)
	case "windows":
		check.Warnings = append(check.Warnings, "Windows support for OpenSCAP may be limited")
	case "darwin":
		check.Warnings = append(check.Warnings, "macOS support for OpenSCAP may be limited")
	default:
		check.Warnings = append(check.Warnings, fmt.Sprintf("Unsupported operating system: %s", os))
	}
}

// checkLinuxDistribution checks the Linux distribution
func (p *PreflightService) checkLinuxDistribution(check *PreflightCheck) {
	// Check /etc/os-release
	cmd := exec.Command("cat", "/etc/os-release")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				distro := strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				check.Info = append(check.Info, fmt.Sprintf("Linux distribution: %s", distro))
				break
			}
		}
	}

	// Check for package managers
	packageManagers := []string{"apt", "yum", "dnf", "zypper", "pacman"}
	for _, pm := range packageManagers {
		cmd := exec.Command("which", pm)
		if err := cmd.Run(); err == nil {
			check.Info = append(check.Info, fmt.Sprintf("Package manager found: %s", pm))
		}
	}
}

// GetInstallationInstructions returns installation instructions for the current system
func (p *PreflightService) GetInstallationInstructions() map[string]interface{} {
	instructions := map[string]interface{}{
		"general": map[string]interface{}{
			"description": "Install OpenSCAP scanner and SCAP Security Guide",
			"packages":    []string{"openscap-scanner", "scap-security-guide"},
		},
		"ubuntu": map[string]interface{}{
			"commands": []string{
				"sudo apt-get update",
				"sudo apt-get install openscap-scanner scap-security-guide",
			},
		},
		"debian": map[string]interface{}{
			"commands": []string{
				"sudo apt-get update",
				"sudo apt-get install openscap-scanner scap-security-guide",
			},
		},
		"centos": map[string]interface{}{
			"commands": []string{
				"sudo yum install openscap-scanner scap-security-guide",
			},
		},
		"rhel": map[string]interface{}{
			"commands": []string{
				"sudo yum install openscap-scanner scap-security-guide",
			},
		},
		"rocky": map[string]interface{}{
			"commands": []string{
				"sudo dnf install openscap-scanner scap-security-guide",
			},
		},
		"fedora": map[string]interface{}{
			"commands": []string{
				"sudo dnf install openscap-scanner scap-security-guide",
			},
		},
		"sles": map[string]interface{}{
			"commands": []string{
				"sudo zypper install openscap-scanner scap-security-guide",
			},
		},
		"opensuse": map[string]interface{}{
			"commands": []string{
				"sudo zypper install openscap-scanner scap-security-guide",
			},
		},
	}

	return instructions
}

// ValidateScapContent validates a SCAP DataStream file
func (p *PreflightService) ValidateScapContent(filePath string) error {
	// Check if file exists
	cmd := exec.Command("test", "-f", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Try to get info from the file
	cmd = exec.Command("oscap", "info", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("invalid SCAP content: %v, output: %s", err, string(output))
	}

	return nil
}

// GetScapContentInfo extracts information from a SCAP DataStream file
func (p *PreflightService) GetScapContentInfo(filePath string) (map[string]interface{}, error) {
	// Get basic info
	cmd := exec.Command("oscap", "info", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get content info: %v", err)
	}

	info := map[string]interface{}{
		"raw_output": string(output),
		"file_path":  filePath,
	}

	// Try to extract structured information
	// This would require parsing the XML output
	// For now, we'll return the raw output

	return info, nil
}

// CheckRunnerCompatibility checks if a runner host is compatible with OpenSCAP
func (p *PreflightService) CheckRunnerCompatibility(host string) (*PreflightCheck, error) {
	// This would need to be implemented to check remote hosts
	// For now, we'll return a basic check
	check := &PreflightCheck{
		Errors:   []string{},
		Warnings: []string{},
		Info:     []string{},
	}

	check.Info = append(check.Info, fmt.Sprintf("Checking compatibility for host: %s", host))
	check.Warnings = append(check.Warnings, "Remote compatibility check not implemented")

	return check, nil
}
