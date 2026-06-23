package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

// runnerRegistrationTokenTTL is how long a one-time registration token issued for
// an unregistered runner stays valid.
const runnerRegistrationTokenTTL = time.Hour

// RunnerRegistrationTokenPrefix prefixes every one-time registration token so it
// is easy to recognize (e.g. in cloud-init scripts or logs).
const RunnerRegistrationTokenPrefix = "smrs_"

// generateRunnerRegistrationToken creates a new one-time registration token and
// returns the plaintext token (handed to the caller once) together with its hash
// (stored in the database, never the plaintext).
func generateRunnerRegistrationToken() (token string, hash string) {
	token = RunnerRegistrationTokenPrefix + base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	hash = HashRunnerRegistrationToken(token)
	return
}

// HashRunnerRegistrationToken hashes a registration token for storage and lookup.
// It is deterministic (SHA-256) so a runner can be found by the token it presents.
func HashRunnerRegistrationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// RunnerService owns the creation of runners for both global and project scopes:
// it decides and generates a runner's credentials (auth token or one-time
// registration token) before persisting it.
type RunnerService interface {
	// CreateRunner generates the runner's credentials and persists it.
	CreateRunner(runner db.Runner) (newRunner db.Runner, err error)

	// RegenerateRegistrationToken issues a fresh one-time registration token and
	// returns its plaintext (handed to the caller once). If the runner was already
	// registered, it is reset to the unregistered state (auth token cleared,
	// deactivated) so it can be registered again.
	RegenerateRegistrationToken(runner db.Runner) (registrationToken string, err error)
}

type RunnerServiceImpl struct {
	runnerRepo db.RunnerManager
}

func NewRunnerService(runnerRepo db.RunnerManager) RunnerService {
	return &RunnerServiceImpl{
		runnerRepo: runnerRepo,
	}
}

func (s *RunnerServiceImpl) CreateRunner(runner db.Runner) (newRunner db.Runner, err error) {
	if runner.Registered {
		runner.Token = db.GenerateRunnerToken()
	} else {
		// An unregistered runner is created with no credentials at all: no auth token
		// and no registration token. It is inactive. A one-time registration token is
		// issued later on demand via RegenerateRegistrationToken.
		runner.Token = ""
		runner.RegistrationTokenHash = nil
		runner.RegistrationTokenExpiresAt = nil
	}

	newRunner, err = s.runnerRepo.CreateRunner(runner)
	return
}

func (s *RunnerServiceImpl) RegenerateRegistrationToken(runner db.Runner) (registrationToken string, err error) {
	token, hash := generateRunnerRegistrationToken()
	expiresAt := tz.Now().Add(runnerRegistrationTokenTTL)

	// This works for both unregistered and already-registered runners: a registered
	// runner is reset to the unregistered state (its auth token is cleared and it is
	// deactivated) and gets a fresh one-time registration token.
	if err = s.runnerRepo.ResetRunnerRegistration(runner.ID, hash, expiresAt); err != nil {
		return
	}

	registrationToken = token
	return
}
