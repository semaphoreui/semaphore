package api

import (
	"database/sql"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
)

//nolint: gocyclo
func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID int

		if authHeader := strings.ToLower(r.Header.Get("authorization")); len(authHeader) > 0 && strings.Contains(authHeader, "bearer") {
			var token models.APIToken
			if err := context.Get(r, "store").(db.Store).Sql().SelectOne(&token, "select * from user__token where id=? and expired=0", strings.Replace(authHeader, "bearer ", "", 1)); err != nil {
				if err == sql.ErrNoRows {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				panic(err)
			}

			userID = token.UserID
		} else {
			// fetch session from cookie
			cookie, err := r.Cookie("semaphore")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			value := make(map[string]interface{})
			if err = util.Cookie.Decode("semaphore", cookie.Value, &value); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, ok := value["user"]
			sessionVal, okSession := value["session"]
			if !ok || !okSession {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID = user.(int)
			sessionID := sessionVal.(int)

			// fetch session
			var session models.Session
			if err := context.Get(r, "store").(db.Store).Sql().SelectOne(&session, "select * from session where id=? and user_id=? and expired=0", sessionID, userID); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if time.Since(session.LastActive).Hours() > 7*24 {
				// more than week old unused session
				// destroy.
				if _, err := context.Get(r, "store").(db.Store).Sql().Exec("update session set expired=1 where id=?", sessionID); err != nil {
					panic(err)
				}

				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if _, err := context.Get(r, "store").(db.Store).Sql().Exec("update session set last_active=? where id=?", time.Now(), sessionID); err != nil {
				panic(err)
			}
		}

		user, err := context.Get(r, "store").(db.Store).GetUserById(userID)
		if err != nil {
			fmt.Println("Can't find user", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		context.Set(r, "user", &user)

		next.ServeHTTP(w, r)
	})
}
