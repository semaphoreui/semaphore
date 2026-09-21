package alerting

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https public host", "https://hooks.slack.com/services/x", false},
		{"http private network", "http://10.0.0.5:8080/hook", false},
		{"empty", "", true},
		{"no scheme", "hooks.slack.com/x", true},
		{"ftp", "ftp://hooks.example/x", true},
		{"javascript", "javascript:alert(1)", true},
		{"localhost", "http://localhost:3000/hook", true},
		{"localhost subdomain", "http://api.localhost/hook", true},
		{"loopback ip", "http://127.0.0.1/hook", true},
		{"ipv6 loopback", "http://[::1]/hook", true},
		{"unspecified", "http://0.0.0.0/hook", true},
		{"link local metadata", "http://169.254.169.254/latest/meta-data", true},
		{"gcp metadata host", "http://metadata.google.internal/computeMetadata/v1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostJSON_TrustedDestinationReachesLoopback(t *testing.T) {
	var got string
	var contentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = string(body)
		contentType = r.Header.Get("Content-Type")
		assert.Equal(t, "v", r.Header.Get("X-Test"))
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	err := postJSON(context.Background(), Destination{Trusted: true}, server.URL, `{"a":1}`, map[string]string{"X-Test": "v"}, 200, 202)
	require.NoError(t, err)
	assert.Equal(t, `{"a":1}`, got)
	assert.Equal(t, "application/json", contentType)
}

func TestPostJSON_UnexpectedStatusIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	err := postJSON(context.Background(), Destination{Trusted: true}, server.URL, "{}", nil, 200)
	assert.ErrorContains(t, err, "500")
}

func TestPostJSON_UntrustedDestinationRejectsLoopback(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
	}))
	defer server.Close()

	err := postJSON(context.Background(), Destination{}, server.URL, "{}", nil, 200)
	assert.Error(t, err)
	assert.False(t, called, "a project webhook must never reach the loopback interface")
}

func TestPostJSON_RedirectsAreNotFollowed(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer redirecting.Close()

	err := postJSON(context.Background(), Destination{Trusted: true}, redirecting.URL, "{}", nil, 200)
	assert.ErrorContains(t, err, "302")
}
