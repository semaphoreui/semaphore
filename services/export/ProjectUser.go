package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type ProjectUserExporter struct {
	ValueMap[db.ProjectUser]
}

func (e *ProjectUserExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		users, err := store.GetProjectUsers(projId, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = e.appendValuesAndCheck(getUsers(users, projId), strconv.Itoa(projId), false)
		if err != nil {
			return err
		}
	}

	return nil
}

func getUsers(vals []db.UserWithProjectRole, projId int) []db.ProjectUser {
	values := make([]db.ProjectUser, 0)

	for _, val := range vals {
		values = append(values, db.ProjectUser{
			UserID:    val.User.ID,
			Role:      val.Role,
			ProjectID: projId,
		})
	}

	return values
}

func (e *ProjectUserExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	for _, val := range e.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		old.UserID, err = exporter.getNewKeyInt(User, GlobalScope, old.UserID, e)
		if err != nil {
			return err
		}

		obj, err := store.CreateProjectUser(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(e.getName(), val.scope, old.ID, obj.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *ProjectUserExporter) exportDependsOn() []string {
	return []string{User, Project}
}

func (e *ProjectUserExporter) importDependsOn() []string {
	return []string{User, Project}
}

func (e *ProjectUserExporter) getName() string {
	return ProjectUser
}
