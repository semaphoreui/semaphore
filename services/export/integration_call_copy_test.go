package export

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Switching database backend copies project data only through this chain, so a
// history left out of it disappears without a word.
func TestCopyKeepsIntegrationCallHistory(t *testing.T) {
	src := sql.InitConfigCreateTestStore()

	proj, err := src.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	key, err := src.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := src.CreateRepository(db.Repository{
		ProjectID: proj.ID, SSHKeyID: key.ID, Name: "Repo",
		GitURL: "git@example.com:t/t", GitBranch: "master",
	})
	require.NoError(t, err)
	tpl, err := src.CreateTemplate(db.Template{
		Name: "Tpl", Playbook: "t.yml", ProjectID: proj.ID, RepositoryID: repo.ID, App: db.AppBash,
	})
	require.NoError(t, err)
	integration, err := src.CreateIntegration(db.Integration{
		Name: "live", ProjectID: proj.ID, TemplateID: &tpl.ID,
	})
	require.NoError(t, err)

	runID, workflowID := 77, 88
	body := `{"ref":"refs/heads/main"}`
	require.NoError(t, src.CreateIntegrationCall(db.IntegrationCall{
		ProjectID: proj.ID, IntegrationID: &integration.ID, IntegrationName: integration.Name,
		Alias: "live-alias", Created: time.Now(), SourceIP: "203.0.113.9", Method: "POST",
		Outcome: db.IntegrationCallOK, RequestSize: len(body), RequestBody: &body,
	}))
	require.NoError(t, src.CreateIntegrationCall(db.IntegrationCall{
		ProjectID: proj.ID, IntegrationName: "deleted integration",
		Alias: "gone-alias", Created: time.Now(), SourceIP: "203.0.113.10", Method: "GET",
		Outcome: db.IntegrationCallSkipped, Reason: "No integration matched the alias",
		RunID: &runID, WorkflowID: &workflowID,
	}))

	dst := sql.InitConfigCreateTestStore()
	chain := InitProjectExporters(NewKeyMapper(), true, false)
	require.NoError(t, chain.Load(src))
	require.NoError(t, chain.Restore(dst, 10))

	projects, err := dst.GetAllProjects()
	require.NoError(t, err)
	require.Len(t, projects, 1)

	calls, err := dst.GetIntegrationCalls(projects[0].ID, db.RetrieveQueryParams{}, "", nil, true)
	require.NoError(t, err)
	require.Len(t, calls, 2, "both calls are copied")

	byAlias := map[string]db.IntegrationCall{}
	for _, call := range calls {
		byAlias[call.Alias] = call
	}

	live := byAlias["live-alias"]
	require.NotNil(t, live.IntegrationID, "a live reference is remapped, not dropped")
	restored, err := dst.GetIntegration(projects[0].ID, *live.IntegrationID)
	require.NoError(t, err, "the remapped id must point at an integration in the destination")
	assert.Equal(t, "live", restored.Name, "and at the same one, not an unrelated object")
	assert.Equal(t, "203.0.113.9", live.SourceIP)
	require.NotNil(t, live.RequestBody, "the body survives a backend migration")
	assert.Equal(t, body, *live.RequestBody)
	assert.Equal(t, len(body), live.RequestSize)

	gone := byAlias["gone-alias"]
	assert.Nil(t, gone.IntegrationID)
	assert.Equal(t, "deleted integration", gone.IntegrationName, "the name carries the history")
	assert.Equal(t, "No integration matched the alias", gone.Reason)
	assert.Nil(t, gone.RunID, "workflow runs are not exported, so the reference is dropped")
	assert.Nil(t, gone.WorkflowID)
}
