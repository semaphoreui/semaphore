package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type projectMember struct {
	ProjectID int   `json:"project_id"`
	UserID    int   `json:"user_id"`
	RoleIDs   []int `json:"role_ids"`
}

func (m projectMember) GetDbKey() EntityKey {
	return NewKeyFromInt(m.ProjectID) + ":" + NewKeyFromInt(m.UserID)
}

type ProjectUserExporter struct {
	ValueMap[projectMember]
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

		values := make([]projectMember, 0)
		for _, user := range users {
			member := projectMember{ProjectID: projId, UserID: user.ID}
			for _, role := range user.Roles {
				member.RoleIDs = append(member.RoleIDs, role.ID)
			}
			values = append(values, member)
		}

		err = e.appendValues(values, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *ProjectUserExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *ProjectUserExporter) restoreValue(val EntityObject[projectMember], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	projectID, err := exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
	if err != nil {
		return err
	}

	userID, err := exporter.getNewKeyInt(User, GlobalScope, old.UserID)
	if err != nil {
		return err
	}

	roleIDs := make([]int, 0, len(old.RoleIDs))
	for _, oldRoleID := range old.RoleIDs {
		roleID, err := exporter.getNewKeyInt(Role, GlobalScope, oldRoleID)
		if err != nil {
			return err
		}
		roleIDs = append(roleIDs, roleID)
	}

	return store.CreateProjectUser(projectID, userID, roleIDs)
}

func (e *ProjectUserExporter) exportDependsOn() []string {
	return []string{User, Project}
}

func (e *ProjectUserExporter) importDependsOn() []string {
	return []string{User, Project, Role}
}

func (e *ProjectUserExporter) getName() string {
	return ProjectUser
}
