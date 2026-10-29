package sql

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

var readBit = strconv.Itoa(int(db.CanReadProject))

// memberRoles lists the roles of a member, assigned directly or through a group mapping.
func memberRoles(projectID int, userID int) (string, []any) {
	roles := `
	select pur.role_id from project__user_role pur where pur.project_id = ? and pur.user_id = ?
	union
	select m.role_id
	from project__user_mapped_role mr
	join group_role_mapping m on m.id = mr.mapping_id
	where mr.project_id = ? and mr.user_id = ?`
	return roles, []any{projectID, userID, projectID, userID}
}

// readsProject is true when a role of the member carries read_project.
func readsProject(projectID int, userID int) (string, []any) {
	roles, args := memberRoles(projectID, userID)
	return "exists (select 1 from `role` r where r.permissions & " + readBit + " <> 0 and r.id in (" + roles + "))", args
}

// readGrant is true when a role of the member holds a read grant in table on the row matched by on.
func readGrant(table string, on string, projectID int, userID int) (string, []any) {
	roles, args := memberRoles(projectID, userID)
	return "exists (select 1 from " + table + " g where " + on + " and g.permissions & " + readBit + " <> 0 and g.role_id in (" + roles + "))", args
}

// VisibleWorkflowsWhere is the SQL that keeps the workflows of alias the member may see: read_project on the
// project, or a grant with read on the workflow. Admins are not filtered, callers skip the fragment for them.
func VisibleWorkflowsWhere(projectID int, userID int, alias string) (string, []any) {
	project, args := readsProject(projectID, userID)
	grant, grantArgs := readGrant("project__workflow_role", "g.workflow_id = "+alias+".id", projectID, userID)
	return "(" + project + " or " + grant + ")", append(args, grantArgs...)
}

func visibleTemplatesWhere(projectID int, userID int, alias string) (string, []any) {
	project, args := readsProject(projectID, userID)
	grant, grantArgs := readGrant("project__template_role", "g.template_id = "+alias+".id", projectID, userID)
	return "(" + project + " or " + grant + ")", append(args, grantArgs...)
}

// A task is visible with its template or with the workflow of its run.
func visibleTasksWhere(projectID int, userID int) (string, []any) {
	project, args := readsProject(projectID, userID)
	template, templateArgs := readGrant("project__template_role", "g.template_id = task.template_id", projectID, userID)
	workflow, workflowArgs := readGrant("project__workflow_role",
		"g.workflow_id = (select wr.workflow_template_id from project__workflow_run wr where wr.id = task.workflow_run_id)",
		projectID, userID)
	args = append(args, templateArgs...)
	args = append(args, workflowArgs...)
	return "(" + project + " or " + template + " or " + workflow + ")", args
}

func filtered(viewer *db.Viewer) bool { return viewer != nil && !viewer.Admin }
