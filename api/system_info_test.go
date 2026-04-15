package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"

	"github.com/stretchr/testify/assert"
)

type mockSubscriptionService struct {
	token pro_interfaces.SubscriptionToken
	err   error
}

func (m *mockSubscriptionService) HasActiveSubscription() bool          { return false }
func (m *mockSubscriptionService) CanAddProUser() (bool, error)        { return true, nil }
func (m *mockSubscriptionService) CanAddRunner() (bool, error)         { return true, nil }
func (m *mockSubscriptionService) CanAddTerraformHTTPBackend() (bool, error) {
	return true, nil
}
func (m *mockSubscriptionService) StartValidationCron() {}
func (m *mockSubscriptionService) GetToken() (pro_interfaces.SubscriptionToken, error) {
	return m.token, m.err
}

type mockStore struct {
	db.Store
}

func (m *mockStore) GetGlobalRoles() ([]db.Role, error) {
	return []db.Role{}, nil
}

func buildSystemInfoRequest(store db.Store) *http.Request {
	req, _ := http.NewRequest("GET", "/api/info", nil)
	ctx := context.WithValue(req.Context(), "user", &db.User{ID: 1})
	ctx = context.WithValue(ctx, "store", store)
	return req.WithContext(ctx)
}

func TestGetSystemInfo_SubscriptionError_ReturnsJSON(t *testing.T) {
	util.Config = &util.ConfigType{
		Auth: &util.AuthConfig{
			Totp:  &util.TotpConfig{},
			Email: &util.EmailAuthConfig{},
		},
		Schedule:  &util.ScheduleConfig{},
		Debugging: &util.DebuggingConfig{},
	}

	svc := &mockSubscriptionService{
		err: fmt.Errorf("connection refused"),
	}
	controller := NewSystemInfoController(svc)

	rr := httptest.NewRecorder()
	req := buildSystemInfoRequest(&mockStore{})
	controller.GetSystemInfo(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code,
		"GetSystemInfo must return 200 even when subscription service errors")

	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err, "Response body must be valid JSON")
	assert.Contains(t, body, "version")
	assert.Contains(t, body, "premium_features")
}

func TestGetSystemInfo_NotFound_ReturnsJSON(t *testing.T) {
	util.Config = &util.ConfigType{
		Auth: &util.AuthConfig{
			Totp:  &util.TotpConfig{},
			Email: &util.EmailAuthConfig{},
		},
		Schedule:  &util.ScheduleConfig{},
		Debugging: &util.DebuggingConfig{},
	}

	svc := &mockSubscriptionService{
		err: db.ErrNotFound,
	}
	controller := NewSystemInfoController(svc)

	rr := httptest.NewRecorder()
	req := buildSystemInfoRequest(&mockStore{})
	controller.GetSystemInfo(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err, "Response body must be valid JSON")
	assert.Contains(t, body, "version")
}

func TestGetSystemInfo_ExpiredSubscription_EmptyPlan(t *testing.T) {
	util.Config = &util.ConfigType{
		Auth: &util.AuthConfig{
			Totp:  &util.TotpConfig{},
			Email: &util.EmailAuthConfig{},
		},
		Schedule:  &util.ScheduleConfig{},
		Debugging: &util.DebuggingConfig{},
	}

	svc := &mockSubscriptionService{
		token: pro_interfaces.SubscriptionToken{
			Plan:  "pro",
			State: "expired",
		},
	}
	controller := NewSystemInfoController(svc)

	rr := httptest.NewRecorder()
	req := buildSystemInfoRequest(&mockStore{})
	controller.GetSystemInfo(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "expired", body["subscription_state"])
}
