package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
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
	if !user.HasActiveSubscription {
		user.Pro = false
	}
	helpers.WriteJSON(w, http.StatusOK, user)
}

// linkLdapIdentity attaches an LDAP identity to the current account.
// Proof of ownership is a successful bind with the user's own LDAP credentials.
func linkLdapIdentity(w http.ResponseWriter, r *http.Request) {
	currentUser := helpers.GetFromContext(r, "user").(*db.User)

	var creds struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Provider string `json:"provider"` // LDAP provider ID, default "ldap"
	}
	if !helpers.Bind(w, r, &creds) {
		return
	}

	providerID := creds.Provider
	if providerID == "" {
		providerID = "ldap"
	}

	provider, ok := util.Config.GetLdapProvider(providerID)
	if !ok {
		helpers.WriteErrorStatus(w, "LDAP provider not found", http.StatusBadRequest)
		return
	}

	ldapUser, userDN, err := tryFindLDAPUser(provider, creds.Username, creds.Password)
	if err != nil || ldapUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !ldapProfileMatchesSemaphoreUser(*ldapUser, *currentUser) {
		helpers.WriteErrorStatus(w, "LDAP directory profile does not match your account", http.StatusForbidden)
		return
	}

	err = linkExternalIdentity(helpers.Store(r), *currentUser, db.IdentityTypeLdap, providerID, userDN)
	if err != nil {
		switch {
		case errors.Is(err, errIdentityLinkedToAnother):
			helpers.WriteErrorStatus(w, "This LDAP account is already linked to another user.", http.StatusConflict)
		case errors.Is(err, errProviderAlreadyLinked):
			helpers.WriteErrorStatus(w, "Your account already has a linked LDAP identity. Unlink it first.", http.StatusConflict)
		default:
			log.WithError(err).WithFields(log.Fields{
				"provider": providerID,
				"user_dn":  userDN,
				"context":  "ldap",
			}).Warn("Failed to link LDAP identity")
			helpers.WriteErrorStatus(w, "Failed to link LDAP account", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	var body struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if body.ExpiresAt != nil && !body.ExpiresAt.After(tz.Now()) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
		panic(err)
	}

	token, err := helpers.Store(r).CreateAPIToken(db.APIToken{
		ID:        strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
		UserID:    user.ID,
		Expired:   false,
		ExpiresAt: body.ExpiresAt,
		Name:      body.Name,
	})
	if err != nil {
		panic(err)
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      user.ID,
		ObjectType:  db.EventAPIToken,
		ObjectID:    user.ID,
		Description: fmt.Sprintf("API token %s created", shortTokenID(token.ID)),
	})

	helpers.WriteJSON(w, http.StatusCreated, token)
}

// shortTokenID truncates a token to the same 8-char prefix the API exposes,
// so the full secret never reaches the event log.
func shortTokenID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func deleteAPIToken(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	tokenID := mux.Vars(r)["token_id"]

	err := helpers.Store(r).DeleteAPIToken(user.ID, tokenID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      user.ID,
		ObjectType:  db.EventAPIToken,
		ObjectID:    user.ID,
		Description: fmt.Sprintf("API token %s deleted", shortTokenID(tokenID)),
	})

	w.WriteHeader(http.StatusNoContent)
}
