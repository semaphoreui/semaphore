package util

import "testing"

func TestIsAdminMappingEnable(t *testing.T) {
	p := OidcProvider{
		AdminGroup: "admin",
	}
	if !p.IsAdminMappingEnable() {
		t.Errorf("OIDC admin mapping is not enabled even though \"AdminGroup\" claim is not empty")
	}

	p = OidcProvider{}
	if p.IsAdminMappingEnable() {
		t.Errorf("OIDC admin mapping is enabled even though \"AdminGroup\" claim is empty")
	}
}

func TestIsAdminUserClaims(t *testing.T) {
	// Test IsAdminUserClaims correctly identifies admin users when they're in the admin group
	p := OidcProvider{
		GroupsClaim: "groups",
		AdminGroup:  "admin",
	}
	claims := map[string]any{
		"groups": []string{"group1", "admin", "group2"},
	}

	if p.IsAdminUserClaims(claims) {
		t.Errorf("IsAdminUserClaims does not correctly identify admin users when they are in the admin group")
	}

	// Test IsAdminUserClaims returns false when groups claim is missing
	p = OidcProvider{
		GroupsClaim: "groups",
		AdminGroup:  "admin",
	}
	claims = map[string]any{}

	if p.IsAdminUserClaims(claims) {
		t.Errorf("IsAdminUserClaims does not returns false when groups claim is missing")
	}

	// Test IsAdminUserClaims handles non-array group claims gracefully
	p = OidcProvider{
		GroupsClaim: "groups",
		AdminGroup:  "admin",
	}
	claims = map[string]any{
		"groups": "admin",
	}

	if p.IsAdminUserClaims(claims) {
		t.Errorf("IsAdminUserClaims does not handles non-array group claims gracefully")
	}

	// Test IsAdminUserClaims handles non-string group values within the array
	p = OidcProvider{
		GroupsClaim: "groups",
		AdminGroup:  "admin",
	}
	claims = map[string]any{
		"groups": []int{1, 2},
	}

	if p.IsAdminUserClaims(claims) {
		t.Errorf("IsAdminUserClaims does not handles non-string group values within the array")
	}
}
