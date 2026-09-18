package services

import (
	"net/http"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type SessionService interface {
	GetSession(cookie http.Cookie) (*db.Session, bool)
}

type sessionServiceImpl struct {
	sessionRepo db.SessionManager
}

func NewSessionService(sessionRepo db.SessionManager) SessionService {
	return &sessionServiceImpl{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionServiceImpl) GetSession(cookie http.Cookie) (*db.Session, bool) {
	var err error

	value := make(map[string]any)
	if err = util.Cookie.Decode("semaphore", cookie.Value, &value); err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	user, ok := value["user"]
	sessionVal, okSession := value["session"]
	if !ok || !okSession {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	userID := user.(int)
	sessionID := sessionVal.(int)

	// fetch session
	session, err := s.sessionRepo.GetSession(userID, sessionID)

	if err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	if session.IsExpiredAt(tz.Now(), util.Config.MaxSessionLife(), db.SessionInactivityTimeout) {
		// The session was unused for too long or outlived the configured
		// absolute lifetime. Destroy it so it cannot be reused.
		if err = s.sessionRepo.ExpireSession(userID, sessionID); err != nil {
			// it is internal error, it doesn't concern the user
			log.Error(err)
		}

		return nil, false
	}

	return &session, true
}
