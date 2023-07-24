package api

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/util"
)

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
	defer l.Close()

	// First bind with a read only user
	if err = l.Bind(util.Config.LdapBindDN, util.Config.LdapBindPassword); err != nil {
		return nil, err
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		util.Config.LdapSearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(util.Config.LdapSearchFilter, username),
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
	userdn := sr.Entries[0].DN
	if err = l.Bind(userdn, password); err != nil {
		return nil, err
	}

	// Ensure authentication and verify itself with whoami operation
	var res *ldap.WhoAmIResult
	if res, err = l.WhoAmI(nil); err != nil {
		return nil, err
	}
	if len(res.AuthzID) <= 0 {
		return nil, fmt.Errorf("error while doing whoami operation")
	}

	// Second time bind as read only user
	if err = l.Bind(util.Config.LdapBindDN, util.Config.LdapBindPassword); err != nil {
		return nil, err
	}

	// Get user info
	searchRequest = ldap.NewSearchRequest(
		util.Config.LdapSearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(util.Config.LdapSearchFilter, username),
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

	ldapUser := db.User{
		Username: strings.ToLower(sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.UID)),
		Created:  time.Now(),
		Name:     sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.CN),
		Email:    sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.Mail),
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
func createSession(w http.ResponseWriter, r *http.Request, user db.User) {
	newSession, err := helpers.Store(r).CreateSession(db.Session{
		UserID:     user.ID,
		Created:    time.Now(),
		LastActive: time.Now(),
		IP:         r.Header.Get("X-Real-IP"),
		UserAgent:  r.Header.Get("user-agent"),
		Expired:    false,
	})

	if err != nil {
		panic(err)
	}

	encoded, err := util.Cookie.Encode("semaphore", map[string]interface{}{
		"user":    user.ID,
		"session": newSession.ID,
	})
	if err != nil {
		panic(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "semaphore",
		Value: encoded,
		Path:  "/",
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

	if err == db.ErrNotFound {
		user, err = store.CreateUserWithoutPassword(ldapUser)
	}

	if !user.External {
		err = db.ErrNotFound
		return
	}

	return
}

type loginMetadataOidcProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type loginMetadata struct {
	OidcProviders     []loginMetadataOidcProvider `json:"oidc_providers"`
	LoginWithPassword bool                        `json:"login_with_password"`
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
				ID:   k,
				Name: v.DisplayName,
			}
			i++
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
			log.Warn(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
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
		if err == db.ErrNotFound {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch err.(type) {
		case *db.ValidationError:
			// TODO: Return more informative error code.
		}

		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	createSession(w, r, user)

	w.WriteHeader(http.StatusNoContent)
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "semaphore",
		Value:   "",
		Expires: time.Now().Add(24 * 7 * time.Hour * -1),
		Path:    "/",
	})

	w.WriteHeader(http.StatusNoContent)
}

func getOidcProvider(id string, ctx context.Context) (*oidc.Provider, *oauth2.Config, error) {
	provider, ok := util.Config.OidcProviders[id]
	if !ok {
		return nil, nil, fmt.Errorf("No such provider: %s", id)
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
	if len(provider.AutoDiscovery) > 0 {
		oidcProvider, err = oidc.NewProvider(ctx, provider.AutoDiscovery)
		if err != nil {
			return nil, nil, err
		}
	}

	oauthConfig := oauth2.Config{
		Endpoint:     oidcProvider.Endpoint(),
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
	}
	if len(oauthConfig.RedirectURL) == 0 {
		rurl, err := url.JoinPath(util.Config.WebHost, "api/auth/oidc", id, "redirect")
		if err != nil {
			return nil, nil, err
		}
		oauthConfig.RedirectURL = rurl
	}
	if len(oauthConfig.Scopes) == 0 {
		oauthConfig.Scopes = []string{"openid", "profile", "email"}
	}
	return oidcProvider, &oauthConfig, nil
}

func oidcLogin(w http.ResponseWriter, r *http.Request) {
	pid := mux.Vars(r)["provider"]
	ctx := context.Background()
	_, oauth, err := getOidcProvider(pid, ctx)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}
	state := generateStateOauthCookie(w)
	u := oauth.AuthCodeURL(state)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	oauthState := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: oauthState, Expires: expiration}
	http.SetCookie(w, &cookie)

	return oauthState
}

func oidcRedirect(w http.ResponseWriter, r *http.Request) {
	pid := mux.Vars(r)["provider"]
	oauthState, err := r.Cookie("oauthstate")
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	ctx := context.Background()
	_oidc, oauth, err := getOidcProvider(pid, ctx)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}
	provider, ok := util.Config.OidcProviders[pid]
	if !ok {
		log.Error(fmt.Errorf("No such provider: %s", pid))
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}
	verifier := _oidc.Verifier(&oidc.Config{ClientID: oauth.ClientID})

	if r.FormValue("state") != oauthState.Value {
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	oauth2Token, err := oauth.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Error(fmt.Errorf("id_token is missing in token response"))
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	// Extract custom claims
	claims := make(map[string]interface{})
	if err := idToken.Claims(&claims); err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	if len(provider.UsernameClaim) == 0 {
		provider.UsernameClaim = "preferred_username"
	}
	usernameClaim, ok := claims[provider.UsernameClaim].(string)
	if !ok {
		log.Error(fmt.Errorf("Claim '%s' missing from id_token or not a string", provider.UsernameClaim))
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}
	if len(provider.NameClaim) == 0 {
		provider.NameClaim = "preferred_username"
	}
	nameClaim, ok := claims[provider.NameClaim].(string)
	if !ok {
		log.Error(fmt.Errorf("Claim '%s' missing from id_token or not a string", provider.NameClaim))
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}
	if len(provider.EmailClaim) == 0 {
		provider.EmailClaim = "email"
	}
	emailClaim, ok := claims[provider.EmailClaim].(string)
	if !ok {
		log.Error(fmt.Errorf("Claim '%s' missing from id_token or not a string", provider.EmailClaim))
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	user, err := helpers.Store(r).GetUserByLoginOrEmail(usernameClaim, emailClaim)
	if err != nil {
		user = db.User{
			Username: usernameClaim,
			Name:     nameClaim,
			Email:    emailClaim,
			External: true,
		}
		user, err = helpers.Store(r).CreateUserWithoutPassword(user)
		if err != nil {
			log.Error(err.Error())
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}
	}

	if !user.External {
		log.Error(fmt.Errorf("OIDC user '%s' conflicts with local user", user.Username))
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	createSession(w, r, user)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
