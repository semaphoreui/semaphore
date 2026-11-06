package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	log "github.com/sirupsen/logrus"
)

// mcpSettingsMaxBodyBytes bounds the PUT body: five booleans fit many times over.
const mcpSettingsMaxBodyBytes = 1024

// mcpSettingsController serves the server-wide MCP settings at /api/mcp/settings.
type mcpSettingsController struct {
	store db.OptionsManager
}

func newMCPSettingsController(store db.OptionsManager) *mcpSettingsController {
	return &mcpSettingsController{store: store}
}

// Get answers any signed-in user: the values hold no secret, and the
// connection tab of every user needs them.
func (c *mcpSettingsController) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := db.ReadMCPSettings(c.store)
	if err != nil {
		log.WithError(err).WithField("context", "mcp").Error("Failed to read the MCP settings")
		helpers.WriteErrorStatus(w, "Can not read the MCP settings", http.StatusInternalServerError)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, settings)
}

// Set sits under systemAPI: globalPermissionMiddleware(db.CanManageGlobalSystem)
// has already admitted the caller (an admin or a holder of a global role with that permission).
func (c *mcpSettingsController) Set(w http.ResponseWriter, r *http.Request) {
	settings, ok := bindMCPSettings(w, r)
	if !ok {
		return
	}

	if err := db.WriteMCPSettings(c.store, settings); err != nil {
		log.WithError(err).WithField("context", "mcp").Error("Failed to store the MCP settings")
		helpers.WriteErrorStatus(w, "Can not store the MCP settings", http.StatusInternalServerError)
		return
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:     audit.SystemSettingsUpdate,
		Metadata: audit.SettingsMetadata{Keys: db.MCPSettingsOptions()},
	})

	// The answer is what the store now holds, as every node reads it.
	stored, err := db.ReadMCPSettings(c.store)
	if err != nil {
		log.WithError(err).WithField("context", "mcp").Error("Failed to read the MCP settings")
		helpers.WriteErrorStatus(w, "Can not read the MCP settings", http.StatusInternalServerError)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, stored)
}

// bindMCPSettings accepts one JSON object with exactly the five boolean fields.
// Field names are matched exactly: encoding/json alone would also take
// "Allow_Run", and would read null as false.
func bindMCPSettings(w http.ResponseWriter, r *http.Request) (db.MCPSettings, bool) {
	invalid := func(message string) (db.MCPSettings, bool) {
		helpers.WriteErrorStatus(w, message, http.StatusBadRequest)
		return db.MCPSettings{}, false
	}

	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, mcpSettingsMaxBodyBytes))
	if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
		return invalid("The request body is too large")
	}
	if err != nil {
		return invalid("Can not read the request body")
	}

	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err = decoder.Decode(&fields); err != nil || fields == nil || decoder.More() {
		return invalid(mcpSettingsBodyError)
	}
	var settings db.MCPSettings
	targets := map[string]*bool{
		"allow_write":         &settings.AllowWrite,
		"allow_run":           &settings.AllowRun,
		"allow_destroy":       &settings.AllowDestroy,
		"admin_allow_write":   &settings.AdminAllowWrite,
		"admin_allow_destroy": &settings.AdminAllowDestroy,
	}
	if len(fields) != len(targets) {
		return invalid(mcpSettingsBodyError)
	}
	for key, target := range targets {
		value, ok := fields[key]
		if !ok || !parseJSONBool(value, target) {
			return invalid(mcpSettingsBodyError)
		}
	}
	return settings, true
}

const mcpSettingsBodyError = "The body must be a JSON object with exactly five boolean fields: " +
	"allow_write, allow_run, allow_destroy, admin_allow_write and admin_allow_destroy"

// parseJSONBool takes the JSON literals true and false only.
func parseJSONBool(value json.RawMessage, target *bool) bool {
	switch string(bytes.TrimSpace(value)) {
	case "true":
		*target = true
	case "false":
		*target = false
	default:
		return false
	}
	return true
}
