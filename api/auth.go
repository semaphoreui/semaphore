package api

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"net/http"
	"strings"
	"time"
)

func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	var userID int

	authHeader := strings.ToLower(r.Header.Get("authorization"))

	if len(authHeader) > 0 && strings.Contains(authHeader, "bearer") {
		token, err := helpers.Store(r).GetAPIToken(strings.Replace(authHeader, "bearer ", "", 1))

		if err != nil {
			if err != db.ErrNotFound {
				log.Error(err)
			}

			w.WriteHeader(http.StatusUnauthorized)
			return
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
		session, err := helpers.Store(r).GetSession(userID, sessionID)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if time.Since(session.LastActive).Hours() > 7*24 {
			// more than week old unused session
			// destroy.
			if err := helpers.Store(r).ExpireSession(userID, sessionID); err != nil {
				// it is internal error, it doesn't concern the user
				log.Error(err)
			}

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := helpers.Store(r).TouchSession(userID, sessionID); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	user, err := helpers.Store(r).GetUser(userID)
	if err != nil {
		if err != db.ErrNotFound {
			// internal error
			log.Error(err)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if util.Config.DemoMode {
		if !user.Admin && r.Method != "GET" &&
			!strings.HasSuffix(r.URL.Path, "/tasks") &&
			!strings.HasSuffix(r.URL.Path, "/stop") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	context.Set(r, "user", &user)
}

// nolint: gocyclo
func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticationHandler(w, r)
		next.ServeHTTP(w, r)
	})
}

// nolint: gocyclo
func authenticationWithStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := helpers.Store(r)
		var url = r.URL.String()

		if !store.KeepConnection() {
			err := store.Connect(url)
			if err != nil {
				panic(err)
			}
		}

		authenticationHandler(w, r)

		if !store.KeepConnection() {
			_ = store.Close(url)
		}

		next.ServeHTTP(w, r)
	})
}
