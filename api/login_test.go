package api

import (
	"testing"
)

func TestParseClaim(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("email | {{ .id }}@test.com", claims)

	if !ok {
		t.Fail()
	}

	if res != "1234567@test.com" {
		t.Fatalf("%s must be %d@test.com", res, claims["id"])
	}
}

func TestParseClaim2(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("username", claims)

	if !ok {
		t.Fail()
	}

	if res != claims["username"] {
		t.Fail()
	}
}

func TestParseClaim3(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("email", claims)

	if ok {
		t.Fail()
	}
}

func TestParseClaim4(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("|", claims)

	if ok {
		t.Fail()
	}
}

func TestParseClaim5(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       123456757343.0,
	}

	prepareClaims(claims)

	res, ok := parseClaim("{{ .id }}", claims)

	if !ok || res != "123456757343" {
		t.Fatalf("Expected: %v, Got: %v", "123456757343", res)
	}
}

func TestExtractGroupsFromClaims(t *testing.T) {
	// Test with string array
	claims := map[string]any{
		"memberOf": []string{"group1", "group2", "group3"},
	}
	groups := extractGroupsFromClaims(claims, "memberOf")
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0] != "group1" || groups[1] != "group2" || groups[2] != "group3" {
		t.Fatalf("unexpected groups: %v", groups)
	}

	// Test with single string
	claims2 := map[string]any{
		"memberOf": "single-group",
	}
	groups2 := extractGroupsFromClaims(claims2, "memberOf")
	if len(groups2) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups2))
	}
	if groups2[0] != "single-group" {
		t.Fatalf("unexpected group: %v", groups2)
	}

	// Test with interface{} array
	claims3 := map[string]any{
		"memberOf": []any{"group1", "group2"},
	}
	groups3 := extractGroupsFromClaims(claims3, "memberOf")
	if len(groups3) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups3))
	}

	// Test with missing attribute
	claims4 := map[string]any{}
	groups4 := extractGroupsFromClaims(claims4, "memberOf")
	if len(groups4) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups4))
	}
}

func TestCheckIfUserIsAdmin(t *testing.T) {
	adminGroups := []string{"admin-group", "super-admin"}

	// Test user in admin group
	userGroups1 := []string{"user-group", "admin-group", "other-group"}
	if !checkIfUserIsAdmin(userGroups1, adminGroups) {
		t.Fatal("expected user to be admin")
	}

	// Test user not in admin group
	userGroups2 := []string{"user-group", "other-group"}
	if checkIfUserIsAdmin(userGroups2, adminGroups) {
		t.Fatal("expected user to not be admin")
	}

	// Test with empty admin groups
	userGroups3 := []string{"admin-group"}
	if checkIfUserIsAdmin(userGroups3, []string{}) {
		t.Fatal("expected user to not be admin when no admin groups configured")
	}
}
