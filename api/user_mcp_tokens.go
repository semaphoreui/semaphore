package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	proFeatures "github.com/semaphoreui/semaphore/pro/pkg/features"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	log "github.com/sirupsen/logrus"
)

// mcpTokenController serves the routes where a user manages their own MCP
// tokens, with the same rights and optional expiry as REST API tokens. The
// REST API accepts an MCP token whatever the license, so listing and revoking
// work in every build and state: a user who lost MCP — the license, the Pro
// seat or the switch — still revokes their tokens. Only issuing needs MCP
// switched on and allowed by the license.
type mcpTokenController struct {
	enabled      bool
	subscription pro_interfaces.SubscriptionService
	// getFeatures is the build's feature table (pro/pkg/features); tests substitute their own.
	getFeatures func(user *db.User, plan string) pro_interfaces.Features
}

func newMCPTokenController(enabled bool, subscription pro_interfaces.SubscriptionService) *mcpTokenController {
	return &mcpTokenController{enabled: enabled, subscription: subscription, getFeatures: proFeatures.GetFeatures}
}

func (c *mcpTokenController) GetTokens(w http.ResponseWriter, r *http.Request) {
	listUserTokens(w, r, db.APITokenTypeMCP)
}

func (c *mcpTokenController) CreateToken(w http.ResponseWriter, r *http.Request) {
	// As in a build without MCP: there is nothing to issue a token for.
	if !c.enabled {
		helpers.WriteErrorStatus(w, "MCP is switched off on this server", http.StatusNotFound)
		return
	}

	// The plan is read on every request, so a license or a seat that changes
	// takes effect at once and on every node alike.
	plan, err := proFeatures.CurrentPlan(c.subscription)
	if err != nil {
		log.WithError(err).WithField("context", "mcp").Error("Failed to read the subscription plan")
		helpers.WriteErrorStatus(w, "Failed to read the subscription plan", http.StatusInternalServerError)
		return
	}
	if !c.getFeatures(helpers.UserFromContext(r), plan).MCP {
		helpers.RecordDenied(r, pro_interfaces.MCPLicenseFeature, 0)
		helpers.WriteErrorStatus(w, pro_interfaces.MCPNeedsLicense, http.StatusForbidden)
		return
	}

	issueUserToken(w, r, db.APITokenTypeMCP)
}

func (c *mcpTokenController) DeleteToken(w http.ResponseWriter, r *http.Request) {
	revoked, ok := revokeUserTokens(w, r, db.APITokenTypeMCP)
	if !ok {
		return
	}
	// A prefix that matches none of the user's MCP tokens is a mistake the
	// client must see, not a revoke that happened.
	if revoked == 0 {
		helpers.WriteErrorStatus(w, "No MCP token of yours starts with this ID", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
