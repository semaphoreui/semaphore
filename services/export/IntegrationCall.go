package export

import (
	"math"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type IntegrationCallExporter struct {
	ValueMap[db.IntegrationCall]
}

func (e *IntegrationCallExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		calls, err := store.GetIntegrationCalls(proj, db.RetrieveQueryParams{Count: math.MaxInt}, "", nil)
		if err != nil {
			return err
		}

		if err = e.appendValues(calls, strconv.Itoa(proj)); err != nil {
			return err
		}
	}

	return nil
}

func (e *IntegrationCallExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *IntegrationCallExporter) restoreValue(val EntityObject[db.IntegrationCall], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	// A call outlives what it refers to, and an id that no longer maps would
	// point at an unrelated object in the destination database. getNewKeyIntRef
	// answers nil for a key it cannot find, which is what the history expects:
	// the alias and the integration name carry it either way.
	old.IntegrationID, err = exporter.getNewKeyIntRef(Integration, val.scope, old.IntegrationID, e)
	if err != nil {
		return err
	}

	old.TaskID, err = exporter.getNewKeyIntRef(Task, val.scope, old.TaskID, e)
	if err != nil {
		return err
	}

	// Workflow runs are not part of the export, so their ids cannot be remapped
	// and are dropped rather than carried over as references to nothing.
	old.RunID = nil
	old.WorkflowID = nil

	return store.CreateIntegrationCall(old)
}

func (e *IntegrationCallExporter) getName() string {
	return IntegrationCall
}

func (e *IntegrationCallExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *IntegrationCallExporter) importDependsOn() []string {
	return []string{Project, Integration, Task}
}
