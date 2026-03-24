package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

var (
	globalKeyfunc  keyfunc.Keyfunc
	globalJWTParser *jwt.Parser
)

// initJWKSCache creates the JWT parser and starts keyfunc's JWKS client.
// keyfunc.NewDefaultCtx performs an initial HTTP fetch (up to 1 min timeout)
// but with NoErrorReturnFirstHTTPReq=true it returns successfully even if the
// endpoint is unreachable. Its built-in refresh goroutine retries hourly.
func initJWKSCache(jwksURL string) {
	if !strings.HasPrefix(jwksURL, "https://") {
		log.Warn("JWT JWKS URL is not HTTPS: ", jwksURL)
	}

	globalJWTParser = newJWTParser(util.Config.Auth.JWT)

	kf, err := keyfunc.NewDefaultCtx(context.Background(), []string{jwksURL})
	if err != nil {
		log.Errorf("JWKS setup for %s failed: %v — JWT auth will not work", jwksURL, err)
		return
	}

	globalKeyfunc = kf
	log.Info("JWKS initialized from ", jwksURL)
}

func newJWTParser(config *util.JWTAuthConfig) *jwt.Parser {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"ES256", "ES384", "ES512", "RS256", "RS384", "RS512"}),
		jwt.WithExpirationRequired(),
	}

	if config.Audience != "" {
		opts = append(opts, jwt.WithAudience(config.Audience))
	}
	if config.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(config.Issuer))
	}

	return jwt.NewParser(opts...)
}

func validateProxyJWT(tokenString string) (map[string]any, error) {
	if globalKeyfunc == nil {
		return nil, fmt.Errorf("JWKS not available — JWT auth is not configured")
	}

	token, err := globalJWTParser.Parse(tokenString, globalKeyfunc.Keyfunc)
	if err != nil {
		// Parse without verification solely to extract iss/aud for operator-facing
		// log messages. The token has already been rejected above.
		unverified, _, parseErr := jwt.NewParser(jwt.WithoutClaimsValidation()).ParseUnverified(tokenString, jwt.MapClaims{})
		if parseErr == nil {
			if claims, ok := unverified.Claims.(jwt.MapClaims); ok {
				return nil, fmt.Errorf("JWT validation failed (iss=%v aud=%v): %w",
					claims["iss"], claims["aud"], err)
			}
		}
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}

	return claims, nil
}

func authenticateByJWT(r *http.Request) (int, error) {
	config := util.Config.Auth.JWT

	tokenString := r.Header.Get(config.GetHeader())
	if tokenString == "" {
		return 0, fmt.Errorf("no JWT in header %s", config.GetHeader())
	}

	claims, err := validateProxyJWT(tokenString)
	if err != nil {
		return 0, err
	}

	prepareClaims(claims)
	parsed, err := parseClaims(claims, config)
	if err != nil {
		return 0, fmt.Errorf("extract claims: %w", err)
	}

	store := helpers.Store(r)

	user, err := store.GetUserByLoginOrEmail("", parsed.email)

	if errors.Is(err, db.ErrNotFound) {
		user = db.User{
			Username: parsed.username,
			Name:     parsed.name,
			Email:    parsed.email,
			External: true,
		}
		user, err = store.CreateUserWithoutPassword(user)
	}

	if err != nil {
		return 0, fmt.Errorf("JWT user lookup/creation: %w", err)
	}

	if !user.External {
		return 0, fmt.Errorf("JWT user %q conflicts with local user", user.Email)
	}

	return user.ID, nil
}
