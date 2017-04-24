package api

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	sq "github.com/masterminds/squirrel"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/ldap.v2"
)

func ldapAuthentication(auth, password string) (error, db.User) {
	if util.Config.LdapEnable != true {
		return fmt.Errorf("LDAP not configured"), db.User{}
	}

	l, err := ldap.Dial("tcp", util.Config.LdapServer)
	if err != nil {
		return err, db.User{}
	}
	defer l.Close()

	// Reconnect with TLS if needed
	if util.Config.LdapNeedTLS == true {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err, db.User{}
		}
	}

	// First bind with a read only user
	err = l.Bind(util.Config.LdapBindDN, util.Config.LdapBindPassword)
	if err != nil {
		return err, db.User{}
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		util.Config.LdapSearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(util.Config.LdapSearchFilter, auth),
		[]string{util.Config.LdapMappings.DN},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return err, db.User{}
	}

	if len(sr.Entries) != 1 {
		return fmt.Errorf("User does not exist or too many entries returned"), db.User{}
	}

	// Bind as the user to verify their password
	userdn := sr.Entries[0].DN
	err = l.Bind(userdn, password)
	if err != nil {
		return err, db.User{}
	}

	// Get user info and ensure authentication in case LDAP supports unauthenticated bind
	searchRequest = ldap.NewSearchRequest(
		util.Config.LdapSearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(util.Config.LdapSearchFilter, auth),
		[]string{util.Config.LdapMappings.DN, util.Config.LdapMappings.Mail, util.Config.LdapMappings.Uid, util.Config.LdapMappings.CN},
		nil,
	)

	sr, err = l.Search(searchRequest)
	if err != nil {
		return err, db.User{}
	}

	ldapUser := db.User{
		Username: sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.Uid),
		Created:  time.Now(),
		Name:     sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.CN),
		Email:    sr.Entries[0].GetAttributeValue(util.Config.LdapMappings.Mail),
		External: true,
		Alert:    false,
	}

	log.Info("User " + ldapUser.Name + " with email " + ldapUser.Email + " authorized via LDAP correctly")
	return nil, ldapUser
}

func login(w http.ResponseWriter, r *http.Request) {
	var login struct {
		Auth     string `json:"auth" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var ldapErr error
	var ldapUser db.User

	if err := mulekick.Bind(w, r, &login); err != nil {
		return
	}

	login.Auth = strings.ToLower(login.Auth)

	var user db.User
	q := sq.Select("*").
		From("user")

	if util.Config.LdapEnable {

		// Try to perform LDAP authentication
		ldapErr, ldapUser = ldapAuthentication(login.Auth, login.Password)

		// If LDAP completed successully - proceed user
		if ldapErr == nil {
			// Check if that user already exist in database
			q = q.Where("username=? and external=true", ldapUser.Username)

			query, args, _ := q.ToSql()
			if err := db.Mysql.SelectOne(&user, query, args...); err != nil {
				if err == sql.ErrNoRows {
					// Create new user
					user = ldapUser
					if err := db.Mysql.Insert(&user); err != nil {
						panic(err)
					}
				} else if err != nil {
					panic(err)
				}
			}
		} else {
			log.Info(ldapErr.Error())
		}
	}

	// If LDAP not enabled, or LDAP auth finished not successfully (wrong login/pass, unreachable server etc)
	// - perform normal authorization
	if util.Config.LdapEnable != true || ldapErr != nil {

		// Perform normal authorization
		println("Perform normal authorization")
		_, err := mail.ParseAddress(login.Auth)
		if err == nil {
			q = q.Where("email=?", login.Auth)
		} else {
			q = q.Where("username=?", login.Auth)
		}

		query, args, _ := q.ToSql()
		if err := db.Mysql.SelectOne(&user, query, args...); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			panic(err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	session := db.Session{
		UserID:     user.ID,
		Created:    time.Now(),
		LastActive: time.Now(),
		IP:         r.Header.Get("X-Real-IP"),
		UserAgent:  r.Header.Get("user-agent"),
		Expired:    false,
	}
	if err := db.Mysql.Insert(&session); err != nil {
		panic(err)
	}

	encoded, err := util.Cookie.Encode("semaphore", map[string]interface{}{
		"user":    user.ID,
		"session": session.ID,
	})
	if err != nil {
		panic(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "semaphore",
		Value: encoded,
		Path:  "/",
	})

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
