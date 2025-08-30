package sql

import (
	"os"
	"regexp"
	"testing"
)

// TestGetIntegrationsByAlias_IncludesTaskParams verifies that TaskParams are loaded
// when retrieving integrations via project-level aliases by checking the source code
func TestGetIntegrationsByAlias_IncludesTaskParams(t *testing.T) {
	// Read the integration_alias.go file to verify the fix is in place
	sourceCode, err := os.ReadFile("integration_alias.go")
	if err != nil {
		t.Fatalf("Could not read integration_alias.go: %v", err)
	}

	// Look for the corrected line that calls GetIntegrations with includeTaskParams=true
	// The fix should have: d.GetIntegrations(aliasObj.ProjectID, db.RetrieveQueryParams{}, true)
	correctPattern := regexp.MustCompile(`d\.GetIntegrations\([^,]+,\s*db\.RetrieveQueryParams\{\},\s*true\)`)
	
	// Also check that the old incorrect pattern (with false) is NOT present
	incorrectPattern := regexp.MustCompile(`d\.GetIntegrations\([^,]+,\s*db\.RetrieveQueryParams\{\},\s*false\)`)

	sourceStr := string(sourceCode)

	// Verify the correct pattern exists
	if !correctPattern.MatchString(sourceStr) {
		t.Error("FAIL: GetIntegrations is not called with includeTaskParams=true for project-level aliases")
		t.Error("Expected to find: d.GetIntegrations(aliasObj.ProjectID, db.RetrieveQueryParams{}, true)")
		return
	}

	// Verify the incorrect pattern does NOT exist (in the context of project-level aliases)
	if incorrectPattern.MatchString(sourceStr) {
		t.Error("FAIL: Found GetIntegrations still being called with includeTaskParams=false")
		t.Error("This suggests the fix was not properly applied")
		return
	}

	// Additional verification: check that the change is in the right context
	// Look for the project-level alias handling code block
	projectLevelPattern := regexp.MustCompile(`if\s+aliasObj\.IntegrationID\s*==\s*nil\s*\{[^}]*d\.GetIntegrations\([^,]+,\s*db\.RetrieveQueryParams\{\},\s*true\)`)
	
	if !projectLevelPattern.MatchString(sourceStr) {
		t.Error("FAIL: The fix is not properly placed within the project-level alias handling block")
		return
	}

	t.Log("SUCCESS: Verified that GetIntegrationsByAlias loads TaskParams (includeTaskParams=true) for project-level aliases")
	t.Log("This ensures survey variable values are available during webhook execution")
}