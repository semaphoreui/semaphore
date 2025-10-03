package compliance

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ARF (Asset Reporting Format) XML structures for parsing scan results

// ArfRoot represents the root element of an ARF file
type ArfRoot struct {
	XMLName xml.Name    `xml:"arf:asset-report-collection"`
	Reports []ArfReport `xml:"arf:reports>arf:report"`
}

// ArfReport represents a report within an ARF file
type ArfReport struct {
	XMLName xml.Name   `xml:"arf:report"`
	ID      string     `xml:"id,attr"`
	Title   string     `xml:"title"`
	Content ArfContent `xml:"arf:content"`
}

// ArfContent represents the content of an ARF report
type ArfContent struct {
	XMLName xml.Name `xml:"arf:content"`
	Data    ArfData  `xml:"sys:system-data"`
}

// ArfData represents system data in an ARF report
type ArfData struct {
	XMLName     xml.Name        `xml:"sys:system-data"`
	TestResults []ArfTestResult `xml:"sys:test-result"`
	Definitions []ArfDefinition `xml:"sys:definition"`
}

// ArfTestResult represents test results
type ArfTestResult struct {
	XMLName   xml.Name  `xml:"sys:test-result"`
	ID        string    `xml:"id,attr"`
	StartTime string    `xml:"start-time,attr"`
	EndTime   string    `xml:"end-time,attr"`
	Result    string    `xml:"result,attr"`
	Score     string    `xml:"score,attr"`
	Rules     []ArfRule `xml:"sys:rule-result"`
}

// ArfDefinition represents rule definitions
type ArfDefinition struct {
	XMLName xml.Name `xml:"sys:definition"`
	ID      string   `xml:"id,attr"`
	Title   string   `xml:"sys:title"`
}

// ArfRule represents individual rule results
type ArfRule struct {
	XMLName     xml.Name        `xml:"sys:rule-result"`
	ID          string          `xml:"idref,attr"`
	Result      string          `xml:"result,attr"`
	Severity    string          `xml:"severity,attr"`
	Weight      string          `xml:"weight,attr"`
	Identifiers []ArfIdentifier `xml:"sys:ident"`
	Messages    []ArfMessage    `xml:"sys:message"`
}

// ArfIdentifier represents rule identifiers
type ArfIdentifier struct {
	XMLName xml.Name `xml:"sys:ident"`
	System  string   `xml:"system,attr"`
	Text    string   `xml:",chardata"`
}

// ArfMessage represents rule messages
type ArfMessage struct {
	XMLName xml.Name `xml:"sys:message"`
	Level   string   `xml:"level,attr"`
	Text    string   `xml:",chardata"`
}

// ScanSummary represents a summary of scan results
type ScanSummary struct {
	OverallResult      string                 `json:"overall_result"`
	Score              float64                `json:"score"`
	RulesTotal         int                    `json:"rules_total"`
	RulesPassed        int                    `json:"rules_passed"`
	RulesFailed        int                    `json:"rules_failed"`
	RulesError         int                    `json:"rules_error"`
	RulesNotApplicable int                    `json:"rules_notapplicable"`
	RulesNotChecked    int                    `json:"rules_notchecked"`
	RulesNotSelected   int                    `json:"rules_notselected"`
	StartTime          string                 `json:"start_time"`
	EndTime            string                 `json:"end_time"`
	Duration           string                 `json:"duration"`
	Details            map[string]interface{} `json:"details"`
}

// RuleResult represents individual rule results
type RuleResult struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Result      string   `json:"result"`
	Severity    string   `json:"severity"`
	Weight      float64  `json:"weight"`
	Identifiers []string `json:"identifiers"`
	Messages    []string `json:"messages"`
}

// ArfParser handles parsing of ARF XML files
type ArfParser struct{}

// NewArfParser creates a new ARF parser
func NewArfParser() *ArfParser {
	return &ArfParser{}
}

// ParseArfFile parses an ARF file and returns scan summary and rule results
func (p *ArfParser) ParseArfFile(filePath string) (*ScanSummary, []RuleResult, error) {
	// Open and read the ARF file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open ARF file: %v", err)
	}
	defer file.Close()

	return p.ParseArfReader(file)
}

// ParseArfReader parses ARF content from a reader
func (p *ArfParser) ParseArfReader(reader io.Reader) (*ScanSummary, []RuleResult, error) {
	// Parse XML
	var arf ArfRoot
	if err := xml.NewDecoder(reader).Decode(&arf); err != nil {
		return nil, nil, fmt.Errorf("failed to parse ARF XML: %v", err)
	}

	if len(arf.Reports) == 0 {
		return nil, nil, fmt.Errorf("no reports found in ARF file")
	}

	// Get the first report
	report := arf.Reports[0]
	if len(report.Content.Data.TestResults) == 0 {
		return nil, nil, fmt.Errorf("no test results found in ARF file")
	}

	// Get the first test result
	testResult := report.Content.Data.TestResults[0]

	// Create scan summary
	summary := &ScanSummary{
		OverallResult: strings.ToLower(testResult.Result),
		StartTime:     testResult.StartTime,
		EndTime:       testResult.EndTime,
		Duration:      p.calculateDuration(testResult.StartTime, testResult.EndTime),
		Details:       make(map[string]interface{}),
	}

	// Parse score
	if score, err := strconv.ParseFloat(testResult.Score, 64); err == nil {
		summary.Score = score
	}

	// Create rule definitions map for lookup
	ruleDefs := make(map[string]string)
	for _, def := range report.Content.Data.Definitions {
		ruleDefs[def.ID] = def.Title
	}

	// Process rule results
	var ruleResults []RuleResult
	for _, rule := range testResult.Rules {
		ruleResult := RuleResult{
			ID:       rule.ID,
			Title:    ruleDefs[rule.ID],
			Result:   strings.ToLower(rule.Result),
			Severity: rule.Severity,
		}

		// Parse weight
		if weight, err := strconv.ParseFloat(rule.Weight, 64); err == nil {
			ruleResult.Weight = weight
		}

		// Extract identifiers
		for _, ident := range rule.Identifiers {
			ruleResult.Identifiers = append(ruleResult.Identifiers, ident.Text)
		}

		// Extract messages
		for _, msg := range rule.Messages {
			ruleResult.Messages = append(ruleResult.Messages, msg.Text)
		}

		ruleResults = append(ruleResults, ruleResult)

		// Update summary counts
		switch ruleResult.Result {
		case "pass":
			summary.RulesPassed++
		case "fail":
			summary.RulesFailed++
		case "error":
			summary.RulesError++
		case "notapplicable":
			summary.RulesNotApplicable++
		case "notchecked":
			summary.RulesNotChecked++
		case "notselected":
			summary.RulesNotSelected++
		}
		summary.RulesTotal++
	}

	return summary, ruleResults, nil
}

// calculateDuration calculates the duration between start and end times
func (p *ArfParser) calculateDuration(startTime, endTime string) string {
	// This is a simplified implementation
	// In reality, you'd parse the timestamps and calculate the actual duration
	if startTime != "" && endTime != "" {
		return fmt.Sprintf("%s to %s", startTime, endTime)
	}
	return ""
}

// ExtractRuleResults extracts only the rule results from an ARF file
func (p *ArfParser) ExtractRuleResults(filePath string) ([]RuleResult, error) {
	summary, results, err := p.ParseArfFile(filePath)
	if err != nil {
		return nil, err
	}

	// We don't need the summary for this method, but we need to call it to parse the file
	_ = summary

	return results, nil
}

// ExtractSummary extracts only the summary from an ARF file
func (p *ArfParser) ExtractSummary(filePath string) (*ScanSummary, error) {
	summary, _, err := p.ParseArfFile(filePath)
	return summary, err
}
