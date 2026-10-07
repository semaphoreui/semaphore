package projects

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDescriptionStore struct {
	db.Store
	calls       int
	projectID   int
	templateID  int
	description string
}

func (f *fakeDescriptionStore) SetTemplateDescription(projectID int, templateID int, description string) error {
	f.calls++
	f.projectID = projectID
	f.templateID = templateID
	f.description = description
	return nil
}

func (f *fakeDescriptionStore) CreateEvent(evt db.Event) (db.Event, error) {
	return evt, nil
}

func descriptionRequest(t *testing.T, store db.Store, description string) *http.Request {
	body, err := json.Marshal(map[string]string{"description": description})
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, "/api/project/12/templates/3/description", bytes.NewReader(body))
	r = helpers.SetContextValue(r, "template", db.Template{ID: 3, ProjectID: 12})
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &db.User{ID: 1})
	r = helpers.SetContextValue(r, "log_writer", nopLogWriter{})
	return r
}

func TestUpdateTemplateDescription_StoresRawMarkdown(t *testing.T) {
	store := &fakeDescriptionStore{}
	markdown := "# Deploy\n\n**bold** [link](https://example.com)\n\n<img src=x onerror=alert(1)>"

	w := httptest.NewRecorder()
	UpdateTemplateDescription(w, descriptionRequest(t, store, markdown))

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 12, store.projectID)
	assert.Equal(t, 3, store.templateID)
	assert.Equal(t, markdown, store.description, "the server stores the Markdown as-is")
}
