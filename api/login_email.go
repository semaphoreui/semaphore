package api

import (
	"errors"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func newEmailOtp(userID int, userEmail string, store db.Store) error {
	code := random.Number(6)
	_, err := store.AddEmailOtpVerification(userID, code)

	if err != nil {
		return err
	}
	err = sendEmailVerificationCode(code, userEmail)
	return err
}

func resendEmailOtp(w http.ResponseWriter, r *http.Request) {

	session, ok := getSession(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := helpers.Store(r).GetUser(session.UserID)

	if err != nil {
		helpers.WriteErrorStatus(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = newEmailOtp(session.UserID, user.Email, helpers.Store(r))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user_id": session.UserID,
			"context": "resend_email_otp",
		}).Error("Failed to add email otp verification")
		helpers.WriteErrorStatus(w, "Failed to create email OTP verification", http.StatusInternalServerError)
		return
	}
}

func loginEmail(w http.ResponseWriter, r *http.Request) {

	var email struct {
		Email string `json:"email" binding:"required"`
	}
	if !helpers.Bind(w, r, &email) {
		return
	}

	store := helpers.Store(r)

	user, err := store.GetUserByLoginOrEmail("", email.Email)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			user, err = store.CreateUserWithoutPassword(db.User{
				Email:    email.Email,
				External: true,
				Alert:    true,
				Pro:      true,
				Name:     getRandomProfileName(),
				Username: getRandomUsername(),
			})
		} else {
			var validationError *db.ValidationError
			switch {
			case errors.As(err, &validationError):
				// TODO: Return more informative error code.
			}

			log.WithError(err).WithFields(log.Fields{
				"email":   email.Email,
				"context": "loginEmail",
			}).Error("Failed to get or create user by email")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	createSession(w, r, user, false)

	w.WriteHeader(http.StatusNoContent)
}
