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
	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	sq "github.com/masterminds/squirrel"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/ldap.v2"
)

func ldapAuthentication(auth, password string) (error, models.User) {

	if util.Config.LdapEnable != true {
		return fmt.Errorf("LDAP not configured"), models.User{}
	}

	bindusername := util.Config.LdapBindDN
	bindpassword := util.Config.LdapBindPassword

	l, err := ldap.Dial("tcp", util.Config.LdapServer)
	if err != nil {
		return err, models.User{}
	}
	defer l.Close()

	// Reconnect with TLS if needed
	if util.Config.LdapNeedTLS == true {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err, models.User{}
		}
	}

	// First bind with a read only user
	err = l.Bind(bindusername, bindpassword)
	if err != nil {
		return err, models.User{}
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
		return err, models.User{}
	}

	if len(sr.Entries) != 1 {
		return fmt.Errorf("User does not exist or too many entries returned"), models.User{}
	}

	// Bind as the user to verify their password
	userdn := sr.Entries[0].DN
	err = l.Bind(userdn, password)
	if err != nil {
		return err, models.User{}
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
		return err, models.User{}
	}

	ldapUser := models.User{
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

func login(c *gin.Context) {
	var login struct {
		Auth     string `json:"auth" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&login); err != nil {
		return
	}

	login.Auth = strings.ToLower(login.Auth)

	ldapErr, ldapUser := ldapAuthentication(login.Auth, login.Password)

	if util.Config.LdapEnable == true && ldapErr != nil {
		log.Info(ldapErr.Error())
	}

	q := sq.Select("*").
		From("user")

	var user models.User
	if ldapErr != nil {
		// Perform normal authorization
		_, err := mail.ParseAddress(login.Auth)
		if err == nil {
			q = q.Where("email=?", login.Auth)
		} else {
			q = q.Where("username=?", login.Auth)
		}

		query, args, _ := q.ToSql()

		if err := database.Mysql.SelectOne(&user, query, args...); err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatus(400)
				return
			}

			panic(err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			c.AbortWithStatus(400)
			return
		}
	} else {
		// Check if that user already exist in database
		q = q.Where("username=? and external=true", ldapUser.Username)

		query, args, _ := q.ToSql()

		if err := database.Mysql.SelectOne(&user, query, args...); err != nil {
			if err == sql.ErrNoRows {
				//Create new user
				user = ldapUser
				if err := database.Mysql.Insert(&user); err != nil {
					panic(err)
				}
			} else if err != nil {
				panic(err)
			}

		}
	}

	session := models.Session{
		UserID:     user.ID,
		Created:    time.Now(),
		LastActive: time.Now(),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.Header.Get("user-agent"),
		Expired:    false,
	}
	if err := database.Mysql.Insert(&session); err != nil {
		panic(err)
	}

	encoded, err := util.Cookie.Encode("semaphore", map[string]interface{}{
		"user":    user.ID,
		"session": session.ID,
	})
	if err != nil {
		panic(err)
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:  "semaphore",
		Value: encoded,
		Path:  "/",
	})

	c.AbortWithStatus(204)
}

func logout(c *gin.Context) {
	c.SetCookie("semaphore", "", -1, "/", "", false, true)
	c.AbortWithStatus(204)
}
