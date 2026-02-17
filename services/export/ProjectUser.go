package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type ProjectUserExporter struct {
	ValueMap[db.ProjectUser]
}

func (a *ProjectUserExporter) load(store db.Store, exporter DataExporter) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {
		users, err := store.GetProjectUsers(projId, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		err = a.appendValuesAndCheck(getUsers(users, projId), strconv.Itoa(projId), false)
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

func (a *ProjectUserExporter) restore(store db.Store, exporter DataExporter) (err error) {
	for _, val := range a.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		old.UserID, err = exporter.getNewKeyInt(User, GlobalScope, old.UserID)
		if err != nil {
			return err
		}

		obj, err := store.CreateProjectUser(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(a.getName(), val.scope, old.ID, obj.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ProjectUserExporter) exportDependsOn() []string {
	return []string{User, Project}
}

func (a *ProjectUserExporter) importDependsOn() []string {
	return []string{User, Project}
}

func (a *ProjectUserExporter) getName() string {
	return ProjectUser
}
