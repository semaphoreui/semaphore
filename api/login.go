package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

func convertEntryToMap(entity *ldap.Entry) map[string]any {
	res := map[string]any{}
	for _, attr := range entity.Attributes {
		if len(attr.Values) == 0 {
			continue
		}
		res[attr.Name] = attr.Values[0]
	}

	return res
}

func tryFindLDAPUser(provider util.LdapProvider, username, password string) (*db.User, string, error) {
	var l *ldap.Conn
	var err error
	if provider.NeedTLS {
		// SECURITY: InsecureSkipVerify=true is pre-existing behavior carried
		// over from the flat config (api/login.go:54); it allows MITM on the
		// LDAP connection. Do not extend it further; see the out-of-scope
		// note about a per-provider tls_skip_verify option (default false).
		l, err = ldap.DialTLS("tcp", provider.Server, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		l, err = ldap.Dial("tcp", provider.Server)
	}

	if err != nil {
		return nil, "", err
	}
	defer l.Close() //nolint:errcheck

	// First bind with a read only user
	if err = l.Bind(provider.BindDN, provider.BindPassword); err != nil {
		return nil, "", err
	}

	mappings := provider.GetMappings()

	// Filter for the given username
	searchRequest := ldap.NewSearchRequest(
		provider.SearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(provider.SearchFilter, ldap.EscapeFilter(username)),
		[]string{mappings.DN},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, "", err
	}

	if len(sr.Entries) < 1 {
		return nil, "", nil
	}

	if len(sr.Entries) > 1 {
		return nil, "", fmt.Errorf("too many entries returned")
	}

	// Bind as the user
	userDN := sr.Entries[0].DN
	if err = l.Bind(userDN, password); err != nil {
		return nil, "", err
	}

	// Second time bind as read only user
	if err = l.Bind(provider.BindDN, provider.BindPassword); err != nil {
		return nil, "", err
	}

	// Get user info
	searchRequest = ldap.NewSearchRequest(
		provider.SearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(provider.SearchFilter, ldap.EscapeFilter(username)),
		[]string{mappings.DN, mappings.Mail, mappings.UID, mappings.CN},
		nil,
	)

	sr, err = l.Search(searchRequest)
	if err != nil {
		return nil, "", err
	}

	if len(sr.Entries) <= 0 {
		return nil, "", fmt.Errorf("ldap search returned no entries")
	}

	entry := convertEntryToMap(sr.Entries[0])

	prepareClaims(entry)

	claims, err := parseClaims(entry, mappings)
	if err != nil {
		return nil, "", err
	}

	ldapUser := db.User{
		Username: strings.ToLower(claims.username),
		Created:  tz.Now(),
		Name:     claims.name,
		Email:    claims.email,
		External: true,
		Alert:    false,
	}

	err = db.ValidateUser(ldapUser)
	if err != nil {
		jsonBytes, _ := json.Marshal(ldapUser)
		log.Error("LDAP returned incorrect user data: " + string(jsonBytes))
		return nil, "", err
	}

	log.Info("User " + ldapUser.Name + " with email " + ldapUser.Email + " authorized via LDAP correctly")
	return &ldapUser, userDN, nil
}

// createSession creates session for passed user and stores session details
// in cookies.
// logAuthEvent records an authentication event (login/logout/failure) in the
// event log. userID may be 0 when the user is unknown (failed attempt).
func logAuthEvent(r *http.Request, action helpers.EventLogType, userID int, description string) {
	helpers.EventLog(r, action, helpers.EventLogItem{
		UserID:      userID,
		ObjectType:  db.EventSession,
		ObjectID:    userID,
		Description: description,
	})
}

func createSession(w http.ResponseWriter, r *http.Request, user db.User, oidc bool) {
	var err error
	var verificationMethod db.SessionVerificationMethod
	verified := false

	switch {
	case user.Totp != nil && util.Config.Mfa.Totp.Enabled:
		verificationMethod = db.SessionVerificationTotp
	default:
		verificationMethod = db.SessionVerificationNone
		verified = true
	}

	newSession, err := helpers.Store(r).CreateSession(db.Session{
		UserID:             user.ID,
		Created:            tz.Now(),
		LastActive:         tz.Now(),
		IP:                 r.Header.Get("X-Real-IP"),
		UserAgent:          r.Header.Get("user-agent"),
		Expired:            false,
		VerificationMethod: verificationMethod,
		Verified:           verified,
	})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user_id": user.ID,
			"context": "session",
		}).Error("Failed to create session")
		helpers.WriteErrorStatus(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	encoded, err := util.Cookie.Encode("semaphore", map[string]any{
		"user":    user.ID,
		"session": newSession.ID,
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user_id": user.ID,
			"context": "session",
		}).Error("Failed to encode session cookie")
		helpers.WriteErrorStatus(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "semaphore",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		// SameSite=Lax prevents the session cookie from being attached to
		// cross-site POST requests, which mitigates CSRF (e.g. the password
		// change endpoint). Top-level GET navigations still carry the cookie
		// so following a link into Semaphore keeps the user logged in.
		SameSite: http.SameSiteLaxMode,
		// Secure is only enforced when Semaphore is served over HTTPS, so that
		// it can still be used without TLS inside private networks.
		Secure: isSecureWebHost(),
	})

	if verified {
		logAuthEvent(r, helpers.EventLogLoginSuccess, user.ID, fmt.Sprintf("User %s logged in", user.Username))
	}
}

// isSecureWebHost reports whether Semaphore's public web host uses HTTPS, in
// which case cookies should carry the Secure attribute.
func isSecureWebHost() bool {
	return util.WebHostURL != nil && util.WebHostURL.Scheme == "https"
}

func loginByPassword(store db.Store, login string, password string) (user db.User, err error) {
	user, err = store.GetUserByLoginOrEmail(login, login)
	if err != nil {
		return
	}

	if user.External {
		err = db.ErrNotFound
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err = db.ErrNotFound
		return
	}

	return
}

func loginByLDAP(store db.Store, ldapUser db.User, userDN string, providerID string) (db.User, error) {
	return resolveExternalUser(store, externalUserProfile{
		Type:            db.IdentityTypeLdap,
		Provider:        providerID,
		ExternalUID:     userDN,
		Username:        ldapUser.Username,
		Name:            ldapUser.Name,
		Email:           ldapUser.Email,
		MatchByUsername: true,
		// The email comes from the directory, not the user - authoritative.
		EmailVerified: true,
	})
}

type loginMetadataOidcProvider struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type LoginTotpAuthMethod struct {
	AllowRecovery bool `json:"allow_recovery"`
}

type LoginEmailAuthMethod struct {
}

type LoginAuthMethods struct {
	Totp  *LoginTotpAuthMethod  `json:"totp,omitempty"`
	Email *LoginEmailAuthMethod `json:"email,omitempty"`
}

type loginMetadataLdapProvider struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

type loginMetadata struct {
	OidcProviders     []loginMetadataOidcProvider `json:"oidc_providers"`
	LdapProviders     []loginMetadataLdapProvider `json:"ldap_providers"`
	LoginWithPassword bool                        `json:"login_with_password"`
	LoginWithLdap     bool                        `json:"login_with_ldap"`
	AuthMethods       LoginAuthMethods            `json:"auth_methods"`
}

// nolint: gocyclo
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		config := &loginMetadata{
			OidcProviders:     make([]loginMetadataOidcProvider, len(util.Config.OidcProviders)),
			LoginWithPassword: !util.Config.PasswordLoginDisable,
		}

		ldapProviders := util.Config.ActiveLdapProviders()
		config.LdapProviders = make([]loginMetadataLdapProvider, 0, len(ldapProviders))
		for _, entry := range ldapProviders {
			name := entry.Provider.DisplayName
			if name == "" {
				name = entry.ID
			}
			config.LdapProviders = append(config.LdapProviders, loginMetadataLdapProvider{
				ID:    entry.ID,
				Name:  name,
				Color: entry.Provider.Color,
				Icon:  entry.Provider.Icon,
			})
		}
		config.LoginWithLdap = len(ldapProviders) > 0

		i := 0

		for k, v := range util.Config.OidcProviders {
			config.OidcProviders[i] = loginMetadataOidcProvider{
				ID:    k,
				Name:  v.DisplayName,
				Color: v.Color,
				Icon:  v.Icon,
			}
			i++
		}

		sort.Slice(config.OidcProviders, func(i, j int) bool {
			a := util.Config.OidcProviders[config.OidcProviders[i].ID]
			b := util.Config.OidcProviders[config.OidcProviders[j].ID]
			return a.Order < b.Order
		})

		if util.Config.Mfa.Totp.Enabled {
			config.AuthMethods.Totp = &LoginTotpAuthMethod{
				AllowRecovery: util.Config.Mfa.Totp.AllowRecovery,
			}
		}

		helpers.WriteJSON(w, http.StatusOK, config)
		return
	}

	var login struct {
		Auth     string `json:"auth" binding:"required"`
		Password string `json:"password" binding:"required"`
		Method   string `json:"method"`   // "", "password" or "ldap"
		Provider string `json:"provider"` // LDAP provider ID when method == "ldap"
	}
	if !helpers.Bind(w, r, &login) {
		return
	}

	login.Auth = strings.ToLower(login.Auth)

	var err error
	var user db.User

	switch login.Method {
	case "password":
		if util.Config.PasswordLoginDisable {
			logAuthEvent(r, helpers.EventLogLoginFail, 0, fmt.Sprintf("Failed login attempt for %s", login.Auth))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user, err = loginByPassword(helpers.Store(r), login.Auth, login.Password)

	case "ldap":
		providerID := login.Provider
		if providerID == "" {
			providerID = "ldap"
		}
		provider, ok := util.Config.GetLdapProvider(providerID)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var ldapUser *db.User
		var ldapUserDN string
		ldapUser, ldapUserDN, err = tryFindLDAPUser(provider, login.Auth, login.Password)
		if err != nil || ldapUser == nil {
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"context":  "ldap",
					"provider": providerID,
					"auth":     login.Auth,
				}).Warn("Failed to find user in LDAP")
			}
			logAuthEvent(r, helpers.EventLogLoginFail, 0, fmt.Sprintf("Failed LDAP login attempt for %s", login.Auth))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user, err = loginByLDAP(helpers.Store(r), *ldapUser, ldapUserDN, providerID)

	default:
		// Legacy clients without the method field: previous behavior —
		// try the legacy flat LDAP first, fall back to password.
		var ldapUser *db.User
		var ldapUserDN string

		if legacy, ok := util.Config.GetLdapProvider("ldap"); ok {
			ldapUser, ldapUserDN, err = tryFindLDAPUser(legacy, login.Auth, login.Password)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"context": "ldap",
					"auth":    login.Auth,
				}).Warn("Failed to find user in LDAP")
				logAuthEvent(r, helpers.EventLogLoginFail, 0, fmt.Sprintf("Failed LDAP login attempt for %s", login.Auth))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		if ldapUser == nil {
			if util.Config.PasswordLoginDisable {
				logAuthEvent(r, helpers.EventLogLoginFail, 0, fmt.Sprintf("Failed login attempt for %s", login.Auth))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			user, err = loginByPassword(helpers.Store(r), login.Auth, login.Password)
		} else {
			user, err = loginByLDAP(helpers.Store(r), *ldapUser, ldapUserDN, "ldap")
		}
	}

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			logAuthEvent(r, helpers.EventLogLoginFail, 0, fmt.Sprintf("Failed login attempt for %s", login.Auth))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var validationError *db.ValidationError
		switch {
		case errors.As(err, &validationError):
			// TODO: Return more informative error code.
		}

		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	createSession(w, r, user, false)

	w.WriteHeader(http.StatusNoContent)
}

// logout handles the user logout process by expiring the current session
// and clearing the session cookie.
//
// Behavior:
//   - If a valid session exists, it is expired in the database.
//   - The session cookie is cleared by setting its value to an empty string
//     and its expiration date to a past time.
//
// Responses:
// - 204 No Content: Logout successful.
// - 500 Internal Server Error: An error occurred while expiring the session.
func logout(w http.ResponseWriter, r *http.Request) {
	if session, ok := getSession(r); ok {
		err := helpers.Store(r).ExpireSession(session.UserID, session.ID)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logAuthEvent(r, helpers.EventLogLogout, session.UserID, "User logged out")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "semaphore",
		Value:    "",
		Expires:  tz.Now().Add(24 * 7 * time.Hour * -1),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureWebHost(),
	})

	w.WriteHeader(http.StatusNoContent)
}

func getOidcProvider(id string, ctx context.Context, redirectPath string) (*oidc.Provider, *oauth2.Config, error) {
	provider, ok := util.Config.OidcProviders[id]
	if !ok {
		return nil, nil, fmt.Errorf("no such provider: %s", id)
	}
	config := oidc.ProviderConfig{
		IssuerURL:   provider.Endpoint.IssuerURL,
		AuthURL:     provider.Endpoint.AuthURL,
		TokenURL:    provider.Endpoint.TokenURL,
		UserInfoURL: provider.Endpoint.UserInfoURL,
		JWKSURL:     provider.Endpoint.JWKSURL,
		Algorithms:  provider.Endpoint.Algorithms,
	}
	oidcProvider := config.NewProvider(ctx)
	var err error
	if provider.AutoDiscovery != "" {
		oidcProvider, err = oidc.NewProvider(ctx, provider.AutoDiscovery)
		if err != nil {
			return nil, nil, err
		}
	}

	clientID := provider.ClientID
	if provider.ClientIDFile != "" {
		if clientID, err = getSecretFromFile(provider.ClientIDFile); err != nil {
			return nil, nil, err
		}
	}

	clientSecret := provider.ClientSecret
	if provider.ClientSecretFile != "" {
		if clientSecret, err = getSecretFromFile(provider.ClientSecretFile); err != nil {
			return nil, nil, err
		}
	}

	if redirectPath != "" {
		redirectPath = strings.TrimRight(redirectPath, "/")

		providerUrl, err2 := url.Parse(provider.RedirectURL)

		if err2 != nil {
			return nil, nil, err2
		}

		providerPath := strings.TrimRight(providerUrl.Path, "/")

		if redirectPath == providerPath {
			redirectPath = ""
		} else if strings.HasPrefix(redirectPath, providerPath+"/") {
			redirectPath = redirectPath[len(providerPath):]
		}
	}

	oauthConfig := oauth2.Config{
		Endpoint:     oidcProvider.Endpoint(),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  provider.RedirectURL + redirectPath,
		Scopes:       provider.Scopes,
	}
	if len(oauthConfig.RedirectURL) == 0 {
		redirectURL, err := url.JoinPath(util.Config.WebHost, "api/auth/oidc", id, "redirect")
		if err != nil {
			return nil, nil, err
		}

		oauthConfig.RedirectURL = redirectURL

		if redirectURL != redirectPath {
			oauthConfig.RedirectURL += redirectPath
		}
	}
	if len(oauthConfig.Scopes) == 0 {
		oauthConfig.Scopes = []string{"openid", "profile", "email"}
	}
	return oidcProvider, &oauthConfig, nil
}

func oidcLogin(w http.ResponseWriter, r *http.Request) {
	pid := mux.Vars(r)["provider"]
	ctx := context.Background()

	returnPath := ""
	redirectPath := ""

	config, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("no such provider: %s", pid))
		http.Error(w, "Unknown OIDC provider.", http.StatusNotFound)
		return
	}

	linkMode := r.URL.Query().Get("link") != ""

	if linkMode {
		// POST-only: SameSite=Lax attaches the session cookie to top-level
		// cross-site GET navigations, so a GET here would let an attacker
		// initiate linking (CSRF) and attach their IdP identity to the
		// victim's account. Lax never sends the cookie on cross-site POST.
		if r.Method != http.MethodPost {
			http.Error(w, "Account linking must be initiated with a POST request.", http.StatusMethodNotAllowed)
			return
		}
		session, ok := getSession(r)
		if !ok || !session.IsVerified() {
			http.Error(w, "You must be signed in to link an external account.", http.StatusUnauthorized)
			return
		}
	}

	returnValue := r.URL.Query().Get("return")
	if returnValue != "" {
		if config.ReturnViaState {
			returnPath = returnValue
		} else {
			redirectPath = returnValue
		}
	}

	_, oauth, err := getOidcProvider(pid, ctx, redirectPath)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "Failed to initialize OIDC provider. Contact your administrator.", http.StatusInternalServerError)
		return
	}
	state := generateStateOauthCookie(w, returnPath, linkMode)
	u := oauth.AuthCodeURL(state)
	status := http.StatusTemporaryRedirect
	if r.Method == http.MethodPost {
		// 303 turns the form POST into a GET on the IdP authorize URL.
		status = http.StatusSeeOther
	}
	http.Redirect(w, r, u, status)
}

type oAuthState struct {
	Csrf   string `json:"csrf"`
	Return string `json:"return"`
	Link   bool   `json:"link,omitempty"`
}

func generateStateOauthCookie(w http.ResponseWriter, returnPath string, link bool) string {

	expiration := tz.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	state := oAuthState{
		Csrf:   base64.URLEncoding.EncodeToString(b),
		Return: returnPath,
		Link:   link,
	}

	// Secure flag is not set to allow Semaphore to be used without HTTPS inside private networks
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state.Csrf,
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(stateBytes)
}

type claimResult struct {
	sub           string
	username      string
	name          string
	email         string
	emailVerified bool
}

// emailVerifiedClaim reads the standard OIDC email_verified claim. Only
// consulted when the provider has require_verified_email enabled, so an
// absent claim counts as NOT verified. Some providers (e.g. AWS Cognito)
// send it as the string "true"/"false".
func emailVerifiedClaim(claims map[string]any) bool {
	switch v := claims["email_verified"].(type) {
	case bool:
		return v
	case string:
		return v == "true"
	default:
		return false
	}
}

func parseClaim(str string, claims map[string]any) (string, bool) {
	for _, s := range strings.Split(str, "|") {
		s = strings.TrimSpace(s)

		if s == "" {
			continue
		}

		if strings.Contains(s, "{{") {
			tpl, err := template.New("").Parse(s)
			if err != nil {
				return "", false
			}

			buff := bytes.NewBufferString("")

			if err = tpl.Execute(buff, claims); err != nil {
				return "", false
			}

			res := buff.String()

			return res, res != ""
		}

		res, ok := claims[s].(string)
		if res != "" && ok {
			return res, ok
		}
	}

	return "", false
}

func prepareClaims(claims map[string]any) {
	for k, v := range claims {
		switch v := v.(type) {
		case float64:
			f := v
			i := int64(f)
			if float64(i) == f {
				claims[k] = i
			}
		case float32:
			f := v
			i := int64(f)
			if float32(i) == f {
				claims[k] = i
			}
		}
	}
}

func parseClaims(claims map[string]any, provider util.ClaimsProvider) (res claimResult, err error) {
	var ok bool
	res.email, ok = parseClaim(provider.GetEmailClaim(), claims)

	if !ok {
		err = fmt.Errorf("claim '%s' missing or has bad format", provider.GetEmailClaim())
		return
	}

	res.username, ok = parseClaim(provider.GetUsernameClaim(), claims)
	if !ok {
		res.username = getRandomUsername()
	}

	res.name, ok = parseClaim(provider.GetNameClaim(), claims)
	if !ok {
		res.name = getRandomProfileName()
	}

	return
}

// oidcEmailVerified reports whether the userinfo email may be used to match
// existing accounts. Always true unless the provider opted into
// require_verified_email; then the email_verified claim must be true.
func oidcEmailVerified(userInfo *oidc.UserInfo, provider util.OidcProvider) bool {
	if !provider.RequireVerifiedEmail {
		return true
	}

	rawClaims := make(map[string]any)
	if err := userInfo.Claims(&rawClaims); err != nil {
		return false
	}

	return emailVerifiedClaim(rawClaims)
}

func claimOidcUserInfo(userInfo *oidc.UserInfo, provider util.OidcProvider) (res claimResult, err error) {
	claims := make(map[string]any)
	if err = userInfo.Claims(&claims); err != nil {
		return
	}

	prepareClaims(claims)

	res, err = parseClaims(claims, &provider)
	res.sub = userInfo.Subject
	res.emailVerified = !provider.RequireVerifiedEmail || emailVerifiedClaim(claims)
	return
}

func claimOidcToken(idToken *oidc.IDToken, provider util.OidcProvider) (res claimResult, err error) {
	claims := make(map[string]any)
	if err = idToken.Claims(&claims); err != nil {
		return
	}

	prepareClaims(claims)

	res, err = parseClaims(claims, &provider)
	res.sub = idToken.Subject
	res.emailVerified = !provider.RequireVerifiedEmail || emailVerifiedClaim(claims)
	return
}

func getRandomUsername() string {
	return random.String(16)
}

func getRandomProfileName() string {
	return "Anonymous"
}

func getSecretFromFile(source string) (string, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func oidcRedirect(w http.ResponseWriter, r *http.Request) {
	pid := mux.Vars(r)["provider"]
	oauthState, err := r.Cookie("oauthstate")

	// Errors are shown as plain text at the current URL instead of a silent
	// redirect to the login page, so the user can see what went wrong.
	// Details stay in server logs.

	if err != nil {
		log.Error(err.Error())
		http.Error(w, "OIDC sign-in failed: state cookie is missing. Try signing in again.", http.StatusBadRequest)
		return
	}

	s := r.FormValue("state")
	b, err := base64.URLEncoding.DecodeString(s)

	if err != nil {
		log.Error(err.Error())
		http.Error(w, "OIDC sign-in failed: invalid state. Try signing in again.", http.StatusBadRequest)
		return
	}

	var stateData oAuthState
	err = json.Unmarshal(b, &stateData)

	if err != nil {
		log.Error(err.Error())
		http.Error(w, "OIDC sign-in failed: invalid state. Try signing in again.", http.StatusBadRequest)
		return
	}

	if stateData.Csrf != oauthState.Value {
		http.Error(w, "OIDC sign-in failed: state mismatch. Try signing in again.", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	_oidc, oauth, err := getOidcProvider(pid, ctx, r.URL.Path)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "Failed to initialize OIDC provider. Contact your administrator.", http.StatusInternalServerError)
		return
	}

	provider, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("no such provider: %s", pid))
		http.Error(w, "Unknown OIDC provider.", http.StatusNotFound)
		return
	}

	verifier := _oidc.Verifier(&oidc.Config{ClientID: oauth.ClientID})

	code := r.URL.Query().Get("code")

	oauth2Token, err := oauth.Exchange(ctx, code)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "OIDC sign-in failed: could not exchange authorization code. Contact your administrator.", http.StatusUnauthorized)
		return
	}

	var claims claimResult

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)

	if ok && rawIDToken != "" {
		var idToken *oidc.IDToken
		// Parse and verify ID Token payload.
		idToken, err = verifier.Verify(ctx, rawIDToken)

		if err == nil {
			claims, err = claimOidcToken(idToken, provider)
		}
	} else {
		var userInfo *oidc.UserInfo
		userInfo, err = _oidc.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))

		if err == nil {
			if userInfo.Email == "" {
				claims, err = claimOidcUserInfo(userInfo, provider)
			} else {
				claims.email = userInfo.Email
				claims.name = userInfo.Profile
				claims.sub = userInfo.Subject
				claims.emailVerified = oidcEmailVerified(userInfo, provider)
			}
		}

		claims.username = getRandomUsername()
		if userInfo.Profile == "" {
			claims.name = getRandomProfileName()
		}
	}

	if err != nil {
		log.Error(err.Error())
		http.Error(w, "OIDC sign-in failed: could not read user info from the provider. Contact your administrator.", http.StatusBadGateway)
		return
	}

	if claims.sub == "" {
		log.Error(fmt.Errorf("oidc provider %s returned no sub claim", pid))
		http.Error(w, "OIDC sign-in failed: the provider returned no user ID (sub claim). Contact your administrator.", http.StatusBadGateway)
		return
	}

	if stateData.Link {
		session, ok := getSession(r)
		if !ok || !session.IsVerified() {
			http.Error(w, "You must be signed in to link an external account.", http.StatusUnauthorized)
			return
		}

		sessionUser, uErr := helpers.Store(r).GetUser(session.UserID)
		if uErr != nil {
			log.Error(uErr.Error())
			http.Error(w, "Failed to link external account.", http.StatusInternalServerError)
			return
		}

		if lErr := linkExternalIdentity(helpers.Store(r), sessionUser, db.IdentityTypeOidc, pid, claims.sub); lErr != nil {
			log.WithError(lErr).WithFields(log.Fields{
				"user_id":  sessionUser.ID,
				"provider": pid,
				"context":  "oidc_link",
			}).Error("Failed to link external identity")

			switch {
			case errors.Is(lErr, errIdentityLinkedToAnother):
				http.Error(w, "This external account is already linked to another user.", http.StatusConflict)
			case errors.Is(lErr, errProviderAlreadyLinked):
				http.Error(w, "Your account already has a linked identity for this provider. Unlink it first.", http.StatusConflict)
			default:
				http.Error(w, "Failed to link external account.", http.StatusInternalServerError)
			}
			return
		}

		redirectURL, _ := url.JoinPath(util.Config.WebHost, "/")
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	user, err := resolveExternalUser(helpers.Store(r), externalUserProfile{
		Type:          db.IdentityTypeOidc,
		Provider:      pid,
		ExternalUID:   claims.sub,
		Username:      claims.username,
		Name:          claims.name,
		Email:         claims.email,
		EmailVerified: claims.emailVerified,
		// MatchByUsername stays false: OIDC matches by email only
		// (username matching "creates a lot of problems" - see old comment).
	})
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "OIDC sign-in failed: could not find or create the user account. Contact your administrator.", http.StatusInternalServerError)
		return
	}

	createSession(w, r, user, true)

	config, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("no such provider: %s", pid))
		http.Error(w, "Unknown OIDC provider.", http.StatusNotFound)
		return
	}

	redirectPath := ""
	if config.ReturnViaState {
		redirectPath = stateData.Return
	} else {
		redirectPath = mux.Vars(r)["redirect_path"]
	}

	if !strings.HasPrefix(redirectPath, "/") {
		redirectPath = "/" + redirectPath
	}

	redirectURL, err := url.JoinPath(util.Config.WebHost, redirectPath)
	if err != nil {
		log.Error(err)
		http.Error(w, "OIDC sign-in failed: invalid redirect URL.", http.StatusInternalServerError)
		return
	}

	if redirectURL == "" {
		redirectURL = "/"
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
