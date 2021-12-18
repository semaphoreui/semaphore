package api

import (
	"crypto/tls"
	"fmt"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/util"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/ldap.v2"
)

func findLDAPUser(username, password string) (*db.User, error) {
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

	// Reconnect with TLS if needed
	if util.Config.LdapNeedTLS {
		// TODO: InsecureSkipVerify should be configurable
		tlsConf := tls.Config{
			InsecureSkipVerify: true, //nolint: gas
		}
		if err = l.StartTLS(&tlsConf); err != nil {
			return nil, err
		}
	}

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

	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("user does not exist or too many entries returned")
	}

	// Bind as the user to verify their password
	userdn := sr.Entries[0].DN
	if err = l.Bind(userdn, password); err != nil {
		return nil, err
	}

	// Get user info and ensure authentication in case LDAP supports unauthenticated bind
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

	ldapUser := db.User{
		Username: sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.UID),
		Created:  time.Now(),
		Name:     sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.CN),
		Email:    sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.Mail),
		External: true,
		Alert:    false,
	}

	log.Info("User " + ldapUser.Name + " with email " + ldapUser.Email + " authorized via LDAP correctly")
	return &ldapUser, nil
}

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

func info(w http.ResponseWriter, r *http.Request) {
	var info struct {
		NewUserRequired bool `json:"newUserRequired"`
	}

	if util.Config.RegisterFirstUser {
		hasPlaceholderUser, err := db.HasPlaceholderUser(helpers.Store(r))
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		info.NewUserRequired = hasPlaceholderUser
	}

	helpers.WriteJSON(w, http.StatusOK, info)
}

func register(w http.ResponseWriter, r *http.Request) {
	var user db.UserWithPwd
	if !helpers.Bind(w, r, &user) {
		return
	}

	if !util.Config.RegisterFirstUser {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hasPlaceholderUser, err := db.HasPlaceholderUser(helpers.Store(r))
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !hasPlaceholderUser {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newUser, err := db.ReplacePlaceholderUser(helpers.Store(r), user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createSession(w, r, newUser)

	w.WriteHeader(http.StatusNoContent)
}

//nolint: gocyclo
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

	var ldapUser *db.User
	if util.Config.LdapEnable {
		// search LDAP for users
		if lu, err := findLDAPUser(login.Auth, login.Password); err == nil {
			ldapUser = lu
		} else {
			log.Info(err.Error())
		}
	}

	user, err := helpers.Store(r).GetUserByLoginOrEmail(login.Auth, login.Auth)

	if err == db.ErrNotFound {
		if ldapUser != nil {
			// create new LDAP user
			user, err = helpers.Store(r).CreateUserWithoutPassword(*ldapUser)
			if err != nil {
				panic(err)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	} else if err != nil {
		panic(err)
	}

	// check if ldap user & no ldap user found
	if user.External && ldapUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// non-ldap login
	if !user.External {
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// authenticated.
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
