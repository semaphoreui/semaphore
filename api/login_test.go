package api

import (
	"testing"
)

func TestParseClaim(t *testing.T) {
	claims := map[string]interface{}{
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
	claims := map[string]interface{}{
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
	claims := map[string]interface{}{
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
	claims := map[string]interface{}{
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
	claims := map[string]interface{}{
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
