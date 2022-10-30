package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/go-ldap/ldap/v3"

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

// nolint: gocyclo
func login(w http.ResponseWriter, r *http.Request) {
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
			log.Info(err.Error())
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
		panic(err)
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
