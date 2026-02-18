package services

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
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

	if time.Since(session.LastActive).Hours() > 7*24 {
		// more than week old unused session
		// destroy.
		if err = s.sessionRepo.ExpireSession(userID, sessionID); err != nil {
			// it is internal error, it doesn't concern the user
			log.Error(err)
		}

		return nil, false
	}

	return &session, true
}
