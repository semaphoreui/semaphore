package api

import (
	"context"
	"encoding/json"
	"testing"
	
	"github.com/semaphoreui/semaphore/util"
	"golang.org/x/oauth2"
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

func TestTokenEndpointAuthMethod_ClientSecretPost(t *testing.T) {
	// Setup test provider configuration using JSON to handle unexported fields
	providerJSON := `{
		"test": {
			"client_id": "test-client-id",
			"client_secret": "test-client-secret",
			"token_endpoint_auth_method": "client_secret_post",
			"endpoint": {
				"issuer": "https://example.com",
				"auth": "https://example.com/auth",
				"token": "https://example.com/token"
			}
		}
	}`
	
	var providers map[string]util.OidcProvider
	if err := json.Unmarshal([]byte(providerJSON), &providers); err != nil {
		t.Fatalf("Failed to parse provider JSON: %v", err)
	}
	
	util.Config = &util.ConfigType{
		OidcProviders: providers,
	}

	ctx := context.Background()
	_, oauthConfig, err := getOidcProvider("test", ctx, "")
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if oauthConfig.Endpoint.AuthStyle != oauth2.AuthStyleInParams {
		t.Fatalf("Expected AuthStyle to be AuthStyleInParams (1), got: %d", oauthConfig.Endpoint.AuthStyle)
	}
}

func TestTokenEndpointAuthMethod_ClientSecretBasic(t *testing.T) {
	// Setup test provider configuration using JSON to handle unexported fields
	providerJSON := `{
		"test": {
			"client_id": "test-client-id",
			"client_secret": "test-client-secret",
			"token_endpoint_auth_method": "client_secret_basic",
			"endpoint": {
				"issuer": "https://example.com",
				"auth": "https://example.com/auth",
				"token": "https://example.com/token"
			}
		}
	}`
	
	var providers map[string]util.OidcProvider
	if err := json.Unmarshal([]byte(providerJSON), &providers); err != nil {
		t.Fatalf("Failed to parse provider JSON: %v", err)
	}
	
	util.Config = &util.ConfigType{
		OidcProviders: providers,
	}

	ctx := context.Background()
	_, oauthConfig, err := getOidcProvider("test", ctx, "")
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if oauthConfig.Endpoint.AuthStyle != oauth2.AuthStyleInHeader {
		t.Fatalf("Expected AuthStyle to be AuthStyleInHeader (2), got: %d", oauthConfig.Endpoint.AuthStyle)
	}
}

func TestTokenEndpointAuthMethod_Default(t *testing.T) {
	// Setup test provider configuration with no auth method specified using JSON
	providerJSON := `{
		"test": {
			"client_id": "test-client-id",
			"client_secret": "test-client-secret",
			"endpoint": {
				"issuer": "https://example.com",
				"auth": "https://example.com/auth",
				"token": "https://example.com/token"
			}
		}
	}`
	
	var providers map[string]util.OidcProvider
	if err := json.Unmarshal([]byte(providerJSON), &providers); err != nil {
		t.Fatalf("Failed to parse provider JSON: %v", err)
	}
	
	util.Config = &util.ConfigType{
		OidcProviders: providers,
	}

	ctx := context.Background()
	_, oauthConfig, err := getOidcProvider("test", ctx, "")
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if oauthConfig.Endpoint.AuthStyle != oauth2.AuthStyleAutoDetect {
		t.Fatalf("Expected AuthStyle to be AuthStyleAutoDetect (0), got: %d", oauthConfig.Endpoint.AuthStyle)
	}
}

