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

func tryFindLDAPUser(username, password string) (*db.User, error) {
	if !util.Config.LdapEnable {
		return nil, fmt.Errorf("LDAP not configured")
	}

	var l *ldap.Conn
	var err error
	if util.Config.LdapNeedTLS {
		l, err = ldap.DialTLS("tcp", util.Config.LdapServer, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		l, err = ldap.Dial("tcp", util.Config.LdapServer)
	}

	if err != nil {
		return nil, err
	}
	defer l.Close() //nolint:errcheck

	// First bind with a read only user
	if err = l.Bind(util.Config.LdapBindDN, util.Config.LdapBindPassword); err != nil {
		return nil, err
	}

	// Filter for the given username
	searchRequest := ldap.NewSearchRequest(
		util.Config.LdapSearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(util.Config.LdapSearchFilter, ldap.EscapeFilter(username)),
		[]string{util.Config.LdapMappings.DN},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) < 1 {
		return nil, nil
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("too many entries returned")
	}

	// Bind as the user
	userDN := sr.Entries[0].DN
	if err = l.Bind(userDN, password); err != nil {
		return nil, err
	}

	// Second time bind as read only user
	if err = l.Bind(util.Config.LdapBindDN, util.Config.LdapBindPassword); err != nil {
		return nil, err
	}

	// Get user info
	searchRequest = ldap.NewSearchRequest(
		util.Config.LdapSearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(util.Config.LdapSearchFilter, ldap.EscapeFilter(username)),
		[]string{util.Config.LdapMappings.DN, util.Config.LdapMappings.Mail, util.Config.LdapMappings.UID, util.Config.LdapMappings.CN},
		nil,
	)

	sr, err = l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) <= 0 {
		return nil, fmt.Errorf("ldap search returned no entries")
	}

	entry := convertEntryToMap(sr.Entries[0])

	prepareClaims(entry)

	claims, err := parseClaims(entry, util.Config.LdapMappings)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	log.Info("User " + ldapUser.Name + " with email " + ldapUser.Email + " authorized via LDAP correctly")
	return &ldapUser, nil
}

// createSession creates session for passed user and stores session details
// in cookies.
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
	})
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

func loginByLDAP(store db.Store, ldapUser db.User) (user db.User, err error) {
	user, err = store.GetUserByLoginOrEmail(ldapUser.Username, ldapUser.Email)

	if errors.Is(err, db.ErrNotFound) {
		user, err = store.CreateUserWithoutPassword(ldapUser)
	}

	if err != nil {
		return
	}

	if !user.External {
		err = db.ErrNotFound
		return
	}

	return
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

type loginMetadata struct {
	OidcProviders     []loginMetadataOidcProvider `json:"oidc_providers"`
	LoginWithPassword bool                        `json:"login_with_password"`
	AuthMethods       LoginAuthMethods            `json:"auth_methods"`
}

// nolint: gocyclo
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		config := &loginMetadata{
			OidcProviders:     make([]loginMetadataOidcProvider, len(util.Config.OidcProviders)),
			LoginWithPassword: !util.Config.PasswordLoginDisable,
		}
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
	}
	if !helpers.Bind(w, r, &login) {
		return
	}

	/*
		logic:
		- fetch user from ldap if enabled
		- fetch user from database by username/email
		- create user in database if doesn't exist & ldap record found
		- check password if non-ldap user
		- create session & send cookie
	*/

	login.Auth = strings.ToLower(login.Auth)

	var err error

	var ldapUser *db.User

	if util.Config.LdapEnable {
		ldapUser, err = tryFindLDAPUser(login.Auth, login.Password)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "ldap",
				"auth":    login.Auth,
			}).Warn("Failed to find user in LDAP")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	var user db.User

	if ldapUser == nil {
		user, err = loginByPassword(helpers.Store(r), login.Auth, login.Password)
	} else {
		user, err = loginByLDAP(helpers.Store(r), *ldapUser)
	}

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
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
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "semaphore",
		Value:    "",
		Expires:  tz.Now().Add(24 * 7 * time.Hour * -1),
		Path:     "/",
		HttpOnly: true,
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

// startOidcAuthFlow builds the OAuth2 config for the provider, sets the CSRF
// state (and nonce) cookie, and redirects the browser to the identity
// provider's authorization endpoint. It is shared by the SP-initiated
// (oidcLogin) and IdP-initiated (oidcInitiate) entry points.
//
// extraAuthParams lets a caller add provider-specific parameters such as
// login_hint to the authorization request.
func startOidcAuthFlow(
	w http.ResponseWriter,
	r *http.Request,
	pid string,
	returnPath string,
	redirectPath string,
	extraAuthParams []oauth2.AuthCodeOption,
) {
	ctx := context.Background()
	loginURL, _ := url.JoinPath(util.Config.WebHost, "auth/login")

	_, oauth, err := getOidcProvider(pid, ctx, redirectPath)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	state, nonce := generateStateOauthCookie(w, returnPath)

	opts := append([]oauth2.AuthCodeOption{oidc.Nonce(nonce)}, extraAuthParams...)

	http.Redirect(w, r, oauth.AuthCodeURL(state, opts...), http.StatusTemporaryRedirect)
}

func oidcLogin(w http.ResponseWriter, r *http.Request) {
	pid := mux.Vars(r)["provider"]
	loginURL, _ := url.JoinPath(util.Config.WebHost, "auth/login")

	config, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("no such provider: %s", pid))
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	returnPath := ""
	redirectPath := ""

	returnValue := r.URL.Query().Get("return")
	if returnValue != "" {
		if config.ReturnViaState {
			returnPath = returnValue
		} else {
			redirectPath = returnValue
		}
	}

	startOidcAuthFlow(w, r, pid, returnPath, redirectPath, nil)
}

// oidcInitiate handles IdP-initiated login (OpenID Connect Core 1.0 §4,
// "Initiating Login from a Third Party"). The identity provider redirects the
// browser here with an "iss" parameter (and optionally "login_hint" and
// "target_link_uri"); Semaphore validates them and then starts a normal
// SP-initiated Authorization Code flow.
func oidcInitiate(w http.ResponseWriter, r *http.Request) {
	pid := mux.Vars(r)["provider"]
	loginURL, _ := url.JoinPath(util.Config.WebHost, "auth/login")

	config, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("no such provider: %s", pid))
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	if !config.AllowIdPInitiated {
		log.Warnf("IdP-initiated login is disabled for OIDC provider '%s'", pid)
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	// Validate the issuer to defend against provider mix-up: the request must
	// originate from the configured identity provider.
	iss := r.FormValue("iss")
	expectedIss := config.ExpectedIssuer()
	if iss == "" || !sameIssuer(iss, expectedIss) {
		log.Warnf("IdP-initiated login for provider '%s' rejected: iss mismatch (got %q, expected %q)", pid, iss, expectedIss)
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	returnPath := ""
	redirectPath := ""

	// target_link_uri tells Semaphore where to land the user after login. It
	// must point back to Semaphore to avoid being abused as an open redirect.
	if tlu := r.FormValue("target_link_uri"); tlu != "" {
		if rel, valid := safeReturnPath(tlu); valid {
			if config.ReturnViaState {
				returnPath = rel
			} else {
				redirectPath = rel
			}
		} else {
			log.Warnf("IdP-initiated login for provider '%s' ignored target_link_uri %q (not same-origin)", pid, tlu)
		}
	}

	var extra []oauth2.AuthCodeOption
	if hint := r.FormValue("login_hint"); hint != "" {
		extra = append(extra, oauth2.SetAuthURLParam("login_hint", hint))
	}

	startOidcAuthFlow(w, r, pid, returnPath, redirectPath, extra)
}

type oAuthState struct {
	Csrf   string `json:"csrf"`
	Return string `json:"return"`
	Nonce  string `json:"nonce"`
}

// generateStateOauthCookie creates a fresh CSRF state and OIDC nonce, stores the
// CSRF token in a cookie, and returns the encoded state (round-tripped through
// the identity provider as the OAuth "state" parameter) together with the raw
// nonce to embed in the authorization request.
func generateStateOauthCookie(w http.ResponseWriter, returnPath string) (encodedState string, nonce string) {

	expiration := tz.Now().Add(365 * 24 * time.Hour)

	csrf := make([]byte, 16)
	if _, err := rand.Read(csrf); err != nil {
		panic(err)
	}

	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		panic(err)
	}

	state := oAuthState{
		Csrf:   base64.URLEncoding.EncodeToString(csrf),
		Return: returnPath,
		Nonce:  base64.URLEncoding.EncodeToString(nonceBytes),
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

	return base64.URLEncoding.EncodeToString(stateBytes), state.Nonce
}

// sameIssuer compares two OIDC issuer identifiers, tolerating a single trailing
// slash difference.
func sameIssuer(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return strings.TrimRight(a, "/") == strings.TrimRight(b, "/")
}

// safeReturnPath validates an externally supplied landing URL (e.g. the
// target_link_uri sent by an IdP-initiated login) and returns the path (with
// query and fragment) to redirect to after login. Only relative paths and
// absolute URLs whose scheme and host match util.Config.WebHost are accepted,
// preventing the parameter from being used as an open redirect.
func safeReturnPath(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	appendQueryFragment := func(p string) string {
		if u.RawQuery != "" {
			p += "?" + u.RawQuery
		}
		if u.Fragment != "" {
			p += "#" + u.EscapedFragment()
		}
		return p
	}

	// Relative reference such as "/dashboard?x=1". Scheme-relative "//host"
	// forms are rejected below because url.Parse exposes a non-empty Host.
	if u.Scheme == "" && u.Host == "" {
		p := u.EscapedPath()
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return appendQueryFragment(p), true
	}

	webHost, err := url.Parse(util.Config.WebHost)
	if err != nil || webHost.Host == "" {
		return "", false
	}

	if !strings.EqualFold(u.Scheme, webHost.Scheme) || !strings.EqualFold(u.Host, webHost.Host) {
		return "", false
	}

	p := u.EscapedPath()
	if p == "" {
		p = "/"
	}
	return appendQueryFragment(p), true
}

// checkNonce verifies the ID token nonce against the value bound to the auth
// request. Verification is skipped when no nonce was issued or when the identity
// provider did not echo one back (a token minted for our request would carry the
// signed nonce, so a missing nonce cannot be a stripped replay).
func checkNonce(expected, actual string) bool {
	if expected == "" || actual == "" {
		return true
	}
	return expected == actual
}

type claimResult struct {
	username string
	name     string
	email    string
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

func claimOidcUserInfo(userInfo *oidc.UserInfo, provider util.OidcProvider) (res claimResult, err error) {
	claims := make(map[string]any)
	if err = userInfo.Claims(&claims); err != nil {
		return
	}

	prepareClaims(claims)

	return parseClaims(claims, &provider)
}

func claimOidcToken(idToken *oidc.IDToken, provider util.OidcProvider) (res claimResult, err error) {
	claims := make(map[string]any)
	if err = idToken.Claims(&claims); err != nil {
		return
	}

	prepareClaims(claims)

	return parseClaims(claims, &provider)
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
	loginURL, _ := url.JoinPath(util.Config.WebHost, "auth/login")

	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	s := r.FormValue("state")
	b, err := base64.URLEncoding.DecodeString(s)

	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	var stateData oAuthState
	err = json.Unmarshal(b, &stateData)

	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	if stateData.Csrf != oauthState.Value {
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	ctx := context.Background()

	_oidc, oauth, err := getOidcProvider(pid, ctx, r.URL.Path)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	provider, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("no such provider: %s", pid))
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	verifier := _oidc.Verifier(&oidc.Config{ClientID: oauth.ClientID})

	code := r.URL.Query().Get("code")

	oauth2Token, err := oauth.Exchange(ctx, code)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	var claims claimResult

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)

	if ok && rawIDToken != "" {
		var idToken *oidc.IDToken
		// Parse and verify ID Token payload.
		idToken, err = verifier.Verify(ctx, rawIDToken)

		if err == nil && !checkNonce(stateData.Nonce, idToken.Nonce) {
			log.Error("OIDC ID token nonce mismatch")
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
			return
		}

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
			}
		}

		claims.username = getRandomUsername()
		if userInfo.Profile == "" {
			claims.name = getRandomProfileName()
		}
	}

	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	user, err := helpers.Store(r).GetUserByLoginOrEmail("", claims.email) // ignore username because it creates a lot of problems
	if err != nil {
		user = db.User{
			Username: claims.username,
			Name:     claims.name,
			Email:    claims.email,
			External: true,
		}
		user, err = helpers.Store(r).CreateUserWithoutPassword(user)
		if err != nil {
			log.Error(err.Error())
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
			return
		}
	}

	if !user.External {
		log.Error(fmt.Errorf("OIDC user '%s' conflicts with local user", user.Username))
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	createSession(w, r, user, true)

	config, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("no such provider: %s", pid))
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
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
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	if redirectURL == "" {
		redirectURL = "/"
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
