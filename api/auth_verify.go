package api

import (
	"bytes"
	"embed"
	"net/http"
	"text/template"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/semaphoreui/semaphore/util/mailer"
	log "github.com/sirupsen/logrus"
)

//go:embed templates/*.tmpl
var templates embed.FS

type emailOtpRequestBody struct {
	Passcode string `json:"passcode"`
}

func verifySessionByEmail(session *db.Session, w http.ResponseWriter, r *http.Request) {
	if !util.Config.Auth.Email.Enabled {
		helpers.WriteErrorStatus(w, "EMAIL_OTP_DISABLED", http.StatusForbidden)
		return
	}

	var body emailOtpRequestBody
	if !helpers.Bind(w, r, &body) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := helpers.Store(r).GetUser(session.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.EmailOtp == nil {
		helpers.WriteErrorStatus(w, "Cannot retrieve verification code from the server.", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.EmailOtp.Code != body.Passcode {
		helpers.WriteErrorStatus(w, "Invalid verification code.", http.StatusUnauthorized)
		return
	}

	if user.EmailOtp.IsExpired() {
		helpers.WriteErrorStatus(w, "The verification code has expired.", http.StatusUnauthorized)
		return
	}

	err = helpers.Store(r).VerifySession(session.UserID, session.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
}

func sendEmailVerificationCode(code string, email string) error {
	body := bytes.NewBufferString("")
	var alert struct {
		Code string
	}

	alert.Code = code

	tpl, err := template.ParseFS(templates, "templates/email_otp_code.tmpl")

	if err != nil {
		return err
	}

	err = tpl.Execute(body, alert)

	if err != nil {
		return err
	}

	content := body.String()

	log.WithFields(log.Fields{
		"email":   email,
		"context": "send_email_verification_code",
	}).Info("send email otp code")

	err = mailer.Send(
		util.Config.EmailSecure,
		util.Config.EmailTls,
		util.Config.EmailHost,
		util.Config.EmailPort,
		util.Config.EmailUsername,
		util.Config.EmailPassword,
		util.Config.EmailSender,
		email,
		"Email verification code",
		content,
	)

	if err != nil {

		log.WithError(err).WithFields(log.Fields{
			"email":   email,
			"context": "send_email_verification_code",
		}).Error("failed to send email verification code")

	}

	return err
}

func startEmailVerification(w http.ResponseWriter, r *http.Request) {
	if !util.Config.Auth.Email.Enabled {
		helpers.WriteErrorStatus(w, "EMAIL_VERIFICATION_DISABLED", http.StatusForbidden)
		return
	}

	session, ok := getSession(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	store := helpers.Store(r)

	user, err := store.GetUser(session.UserID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	code := util.RandString(16)

	err = sendEmailVerificationCode(code, user.Email)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
}
