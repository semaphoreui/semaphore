package api

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
	"io"
	"net/http"
	"strings"
)

type UserController struct {
	subscriptionService pro_interfaces.SubscriptionService
}

func NewUserController(subscriptionService pro_interfaces.SubscriptionService) *UserController {
	return &UserController{
		subscriptionService: subscriptionService,
	}
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	if u, exists := helpers.GetOkFromContext(r, "_user"); exists {
		helpers.WriteJSON(w, http.StatusOK, u)
		return
	}

	var user struct {
		db.User
		CanCreateProject      bool `json:"can_create_project"`
		HasActiveSubscription bool `json:"has_active_subscription"`
	}

	user.User = *helpers.GetFromContext(r, "user").(*db.User)
	user.CanCreateProject = user.Admin || util.Config.NonAdminCanCreateProject
	user.HasActiveSubscription = c.subscriptionService.HasActiveSubscription()

	helpers.WriteJSON(w, http.StatusOK, user)
}

func getAPITokens(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	tokens, err := helpers.Store(r).GetAPITokens(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i := range tokens {
		if len(tokens[i].ID) >= 8 {
			tokens[i].ID = tokens[i].ID[:8]
		}
		// If ID is shorter than 8 chars, leave it as-is
	}

	helpers.WriteJSON(w, http.StatusOK, tokens)
}

func createAPIToken(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)
	tokenID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
		panic(err)
	}

	token, err := helpers.Store(r).CreateAPIToken(db.APIToken{
		ID:      strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
		UserID:  user.ID,
		Expired: false,
	})
	if err != nil {
		panic(err)
	}

	helpers.WriteJSON(w, http.StatusCreated, token)
}

func deleteAPIToken(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	tokenID := mux.Vars(r)["token_id"]

	err := helpers.Store(r).DeleteAPIToken(user.ID, tokenID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
