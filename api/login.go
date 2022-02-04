package api

import (
	"crypto/tls"
	"fmt"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/go-ldap/ldap/v3"
	"net/http"
	"strings"
	"time"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/util"

	"golang.org/x/crypto/bcrypt"
)

func tryFindLDAPUser(username, password string) (*db.User, error) {
	if !util.Config.LdapEnable {
		return nil, fmt.Errorf("LDAP not configured")
	}

	var l *ldap.Conn
	var err error

	l, err = ldap.DialURL( util.Config.LdapServer, ldap.DialWithTLSConfig( &tls.Config{ InsecureSkipVerify: util.Config.LdapSkipVerifyCerts }))
	if err != nil {
		return nil, err
	}

	// Reconnect using StartTLS 
	if util.Config.LdapStartTLS {
		log.Info("Reconnecting with StartTLS")
		defer l.Close()
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: util.Config.LdapSkipVerifyCerts})
		if err != nil {
			return nil, err
		}
	}

	// Initial Bind
	if util.Config.LdapBindDN != "" ||  util.Config.LdapBindPassword != "" {
		log.Info("Binding to LDAP Server using provided bind user: " + util.Config.LdapBindDN)
	        err = l.Bind(util.Config.LdapBindDN, util.Config.LdapBindPassword)
        } else {
		err = l.UnauthenticatedBind("")
		log.Info("Binding to LDAP Server using unauthenticated bind.")
	}
	if err != nil {
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
		log.Info("No results found: " + err.Error() )
		return nil, err
	}

	if len(sr.Entries) < 1 {
		log.Info("User not found: " + strconv.Itoa(len(sr.Entries)) )
		return nil, nil
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("too many entries returned")
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
