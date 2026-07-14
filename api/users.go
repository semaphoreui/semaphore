package api

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/semaphoreui/semaphore/util"
)

type UsersController struct {
	subscriptionService pro_interfaces.SubscriptionService
	log                 *log.Entry
}

func NewUsersController(subscriptionService pro_interfaces.SubscriptionService) *UsersController {
	return &UsersController{
		subscriptionService: subscriptionService,
		log:                 log.WithField("context", "api.users"),
	}
}

type minimalUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func (c *UsersController) GetUsers(w http.ResponseWriter, r *http.Request) {
	currentUser := helpers.GetFromContext(r, "user").(*db.User)
	users, err := helpers.Store(r).GetUsers(db.RetrieveQueryParams{
		Filter: r.URL.Query().Get("s"),
	})

	if err != nil {
		panic(err)
	}

	if currentUser.Admin {
		helpers.WriteJSON(w, http.StatusOK, users)
	} else {
		var result = make([]minimalUser, 0)

		for _, user := range users {
			result = append(result, minimalUser{
				ID:       user.ID,
				Name:     user.Name,
				Username: user.Username,
			})
		}

		helpers.WriteJSON(w, http.StatusOK, result)
	}
}

func (c *UsersController) AddUser(w http.ResponseWriter, r *http.Request) {
	var user db.UserWithPwd
	if !helpers.Bind(w, r, &user) {
		return
	}

	editor := helpers.GetFromContext(r, "user").(*db.User)
	if !editor.Admin {
		c.log.WithField("editor", editor.Username).Debug("Not permitted to create users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Pro {
		ok, err := c.subscriptionService.CanAddProUser()

		if err != nil {
			c.log.WithError(err).Error("Failed to check Pro user limit")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !ok {
			helpers.WriteErrorStatus(
				w,
				"You have reached the limit of Pro users for your subscription.",
				http.StatusForbidden,
			)
			return
		}
	}

	var err error
	var newUser db.User

	if user.External {
		newUser, err = helpers.Store(r).CreateUserWithoutPassword(user.User)
	} else {
		newUser, err = helpers.Store(r).CreateUser(user)
	}

	if err != nil {
		c.log.WithError(err).WithField("username", user.Username).Error("Failed to create user")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newUser)
}
func (c *UsersController) ReadonlyUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := helpers.GetIntParam("user_id", w, r)

		if err != nil {
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		editor := helpers.GetFromContext(r, "user").(*db.User)

		if !editor.Admin && editor.ID != user.ID {
			user = db.User{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
			}
		}

		r = helpers.SetContextValue(r, "_user", user)
		next.ServeHTTP(w, r)
	})
}

func (c *UsersController) GetUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := helpers.GetIntParam("user_id", w, r)

		if err != nil {
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		editor := helpers.GetFromContext(r, "user").(*db.User)

		if !editor.Admin && editor.ID != user.ID {
			c.log.WithFields(log.Fields{
				"editor":  editor.Username,
				"user_id": user.ID,
			}).Debug("Not permitted to access another user")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r = helpers.SetContextValue(r, "_user", user)
		next.ServeHTTP(w, r)
	})
}

func (c *UsersController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	targetUser := helpers.GetFromContext(r, "_user").(db.User)
	editor := helpers.GetFromContext(r, "user").(*db.User)

	var user db.UserWithPwd
	if !helpers.Bind(w, r, &user) {
		return
	}

	if !editor.Admin && (user.Pro && !targetUser.Pro) {
		c.log.WithFields(log.Fields{
			"editor":  editor.Username,
			"user_id": targetUser.ID,
		}).Debug("Not permitted to mark users as Pro")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Pro {
		ok, err := c.subscriptionService.CanAddProUser()

		if err != nil {
			c.log.WithError(err).Error("Failed to check Pro user limit")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !ok {
			helpers.WriteErrorStatus(
				w,
				"You have reached the limit of Pro users for your subscription.",
				http.StatusForbidden,
			)
			return
		}
	}

	if !editor.Admin && editor.ID != targetUser.ID {
		c.log.WithFields(log.Fields{
			"editor":  editor.Username,
			"user_id": targetUser.ID,
		}).Debug("Not permitted to update another user")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if editor.ID == targetUser.ID && targetUser.Admin != user.Admin {
		c.log.WithField("editor", editor.Username).Debug("Not permitted to change own admin status")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if targetUser.External && targetUser.Username != user.Username {
		c.log.WithField("user_id", targetUser.ID).Debug("Username is not editable for external users")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user.ID = targetUser.ID
	if err := helpers.Store(r).UpdateUser(user); err != nil {
		c.log.WithError(err).WithField("user_id", targetUser.ID).Error("Failed to update user")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *UsersController) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "_user").(db.User)
	editor := helpers.GetFromContext(r, "user").(*db.User)

	var pwd struct {
		Pwd        string `json:"password"`
		CurrentPwd string `json:"current_password"`
	}

	if !editor.Admin && editor.ID != user.ID {
		c.log.WithFields(log.Fields{
			"editor":  editor.Username,
			"user_id": user.ID,
		}).Debug("Not permitted to change another user's password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.External {
		c.log.WithField("user_id", user.ID).Debug("Password is not editable for external users")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !helpers.Bind(w, r, &pwd) {
		return
	}

	// CWE-620: require the current password when a user changes their own,
	// so a stolen session can't be used to take over the account. Admins
	// changing someone else's password are exempt — they can't know it.
	if editor.ID == user.ID {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pwd.CurrentPwd)); err != nil {
			c.log.WithField("user_id", user.ID).Debug("Current password does not match")
			helpers.WriteErrorStatus(w, "Current password is incorrect", http.StatusBadRequest)
			return
		}
	}

	if err := helpers.Store(r).SetUserPassword(user.ID, pwd.Pwd); err != nil {
		c.log.WithError(err).WithField("user_id", user.ID).Error("Failed to set user password")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *UsersController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "_user").(db.User)
	editor := helpers.GetFromContext(r, "user").(*db.User)

	if !editor.Admin && editor.ID != user.ID {
		c.log.WithFields(log.Fields{
			"editor":  editor.Username,
			"user_id": user.ID,
		}).Debug("Not permitted to delete another user")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := helpers.Store(r).DeleteUser(user.ID); err != nil {
		c.log.WithError(err).WithField("user_id", user.ID).Error("Failed to delete user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := helpers.Store(r).DeleteOptions(fmt.Sprintf("user%d", user.ID)); err != nil {
		c.log.WithError(err).WithField("user_id", user.ID).Error("Failed to delete options of removed user")
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *UsersController) GetUserIdentities(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "_user").(db.User)

	identities, err := helpers.Store(r).GetUserExternalIdentities(user.ID)
	if err != nil {
		c.log.WithError(err).WithField("user_id", user.ID).Error("Failed to get user identities")
		helpers.WriteErrorStatus(w, "Failed to get identities", http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, identities)
}

func (c *UsersController) DeleteUserIdentity(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "_user").(db.User)
	idType := mux.Vars(r)["type"]
	provider := mux.Vars(r)["provider"]

	if idType != db.IdentityTypeLdap && idType != db.IdentityTypeOidc {
		helpers.WriteErrorStatus(w, "Invalid identity type", http.StatusBadRequest)
		return
	}

	identities, err := helpers.Store(r).GetUserExternalIdentities(user.ID)
	if err != nil {
		c.log.WithError(err).WithField("user_id", user.ID).Error("Failed to get user identities")
		helpers.WriteErrorStatus(w, "Failed to delete identity", http.StatusInternalServerError)
		return
	}
	// Only block unlinking when the requested identity exists and it is the last one.
	targetExists := false
	for _, identity := range identities {
		if identity.Type == idType && identity.Provider == provider {
			targetExists = true
			break
		}
	}
	if user.External && targetExists && len(identities) <= 1 {
		helpers.WriteErrorStatus(w, errCannotUnlinkLastIdentity.Error(), http.StatusConflict)
		return
	}

	if err := helpers.Store(r).DeleteExternalIdentity(user.ID, idType, provider); err != nil {
		c.log.WithError(err).WithField("user_id", user.ID).Error("Failed to delete user identity")
		helpers.WriteErrorStatus(w, "Failed to delete identity", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *UsersController) TotpQr(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "_user").(db.User)

	if user.Totp == nil {
		helpers.WriteErrorStatus(w, "TOTP not enabled", http.StatusNotFound)
		return
	}

	key, err := otp.NewKeyFromURL(user.Totp.URL)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	image, err := key.Image(256, 256)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, image)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	pngBytes := buf.Bytes()

	w.Header().Add("Content-Type", "image/png")
	_, err = w.Write(pngBytes)
}

func (c *UsersController) EnableTotp(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "_user").(db.User)

	if !util.Config.Mfa.Totp.Enabled {
		helpers.WriteErrorStatus(w, "TOTP not enabled", http.StatusBadRequest)
		return
	}

	if user.Totp != nil {
		helpers.WriteErrorStatus(w, "TOTP already enabled", http.StatusBadRequest)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Semaphore",
		AccountName: user.Email,
	})

	if err != nil {
		c.log.WithError(err).WithFields(log.Fields{"user_id": user.ID}).Error("Failed to generate TOTP key")
		http.Error(w, "Error generating key", http.StatusInternalServerError)
		return
	}

	var code, hash string

	if util.Config.Mfa.Totp.AllowRecovery {
		code, hash, err = util.GenerateRecoveryCode()
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	newTotp, err := helpers.Store(r).AddTotpVerification(user.ID, key.URL(), hash)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	newTotp.RecoveryCode = code

	helpers.WriteJSON(w, http.StatusOK, newTotp)
}

func (c *UsersController) DisableTotp(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "_user").(db.User)
	if user.Totp == nil {
		helpers.WriteErrorStatus(w, "TOTP not enabled", http.StatusBadRequest)
		return
	}

	totpID, err := helpers.GetIntParam("totp_id", w, r)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).DeleteTotpVerification(user.ID, totpID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
