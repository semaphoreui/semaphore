package projects

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/audit"
)

// UserMiddleware ensures a user exists and loads it to the context
func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		userID, ok := helpers.GetIntParamOrAbort("user_id", w, r)
		if !ok {
			return
		}

		member, err := helpers.Store(r).GetProjectMember(project.ID, userID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "projectUser", user)
		r = helpers.SetContextValue(r, "projectMember", member)
		next.ServeHTTP(w, r)
	})
}

type mappedRole struct {
	Role     string `json:"role"`
	Provider string `json:"provider"`
	Group    string `json:"group"`
}

type projUser struct {
	ID          int          `json:"id"`
	Username    string       `json:"username"`
	Name        string       `json:"name"`
	Role        string       `json:"role"`
	Roles       []string     `json:"roles"`
	MappedRoles []mappedRole `json:"mapped_roles,omitempty"`
}

// takesLastOwner reports whether the change takes the owner role from the only owner of the project.
func takesLastOwner(store db.Store, projectID int, oldRoles []db.Role, newRoles []db.Role) (bool, error) {
	isOwner := func(role db.Role) bool { return role.Builtin && role.Slug == string(db.ProjectOwner) }
	if !slices.ContainsFunc(oldRoles, isOwner) || slices.ContainsFunc(newRoles, isOwner) {
		return false, nil
	}
	owners, err := store.CountProjectOwners(projectID)
	return owners == 1, err
}

// resolveRoles finds the slugs among the project and global roles, a role of another project counts as unknown.
func resolveRoles(store db.Store, projectID int, role string, roles []string) ([]db.Role, error) {
	slugs := roles
	if role != "" {
		if len(roles) > 0 {
			return nil, &common_errors.ValidationError{Message: "Send either role or roles, not both"}
		}
		slugs = []string{role}
	}
	// An empty list is a request to drop the manual roles, the store decides whether the member keeps access.
	if slugs == nil {
		return nil, &common_errors.ValidationError{Message: "Send role or roles"}
	}
	result := make([]db.Role, 0, len(slugs))
	for _, slug := range slugs {
		found, err := store.GetProjectOrGlobalRoleBySlug(projectID, slug)
		if errors.Is(err, db.ErrNotFound) {
			return nil, &common_errors.ValidationError{Message: "Role does not exist in this project"}
		}
		if err != nil {
			return nil, err
		}
		// Slugs in another case name the same role.
		if !slices.ContainsFunc(result, func(r db.Role) bool { return r.ID == found.ID }) {
			result = append(result, found)
		}
	}
	return result, nil
}

func roleIDs(roles []db.Role) []int {
	ids := make([]int, 0, len(roles))
	for _, role := range roles {
		ids = append(ids, role.ID)
	}
	return ids
}

// GetUsers returns all users in a project
func GetUsers(w http.ResponseWriter, r *http.Request) {
	// get single user if user ID specified in the request
	if user := helpers.GetFromContext(r, "projectUser"); user != nil {
		helpers.WriteJSON(w, http.StatusOK, user.(db.User))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)
	users, err := helpers.Store(r).GetProjectUsers(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	result := make([]projUser, 0)

	for _, user := range users {
		item := projUser{ID: user.ID, Name: user.Name, Username: user.Username, Roles: make([]string, 0, len(user.Roles))}
		for _, role := range user.Roles {
			item.Roles = append(item.Roles, role.Slug)
		}
		// One value would hide the other roles, so old clients see it only for a single role.
		if len(item.Roles) == 1 {
			item.Role = item.Roles[0]
		}
		for _, mapped := range user.MappedRoles {
			item.MappedRoles = append(item.MappedRoles, mappedRole{Role: mapped.RoleSlug, Provider: mapped.Provider, Group: mapped.Group})
		}
		result = append(result, item)
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

// AddUser adds a user to a projects team in the database
func AddUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	store := helpers.Store(r)
	var projectUser struct {
		UserID int      `json:"user_id" binding:"required"`
		Role   string   `json:"role"`
		Roles  []string `json:"roles"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	roles, err := resolveRoles(store, project.ID, projectUser.Role, projectUser.Roles)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	member, err := store.GetProjectMember(project.ID, projectUser.UserID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		helpers.WriteError(w, err)
		return
	}

	err = store.CreateProjectUser(project.ID, projectUser.UserID, roleIDs(roles))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventUser,
		ObjectID:    projectUser.UserID,
		Description: fmt.Sprintf("User ID %d added to team", projectUser.UserID),
	})

	target := audit.UserTarget(projectUser.UserID, "")
	if added, err := helpers.Store(r).GetUser(projectUser.UserID); err == nil {
		target.Name = added.Username
	}
	event := audit.Event{
		Kind:      audit.IAMMembershipAdd,
		Target:    target,
		ProjectID: project.ID,
		Metadata:  audit.MembershipMetadata{Roles: audit.NewRoleRefs(roles)},
	}
	// Mapped roles already made the user a member, so only the manual roles change.
	if len(member.MappedRoles) > 0 {
		event.Kind = audit.IAMProjectRoleChange
		event.Metadata = audit.ProjectRoleMetadata{OldRoles: []audit.RoleRef{}, NewRoles: audit.NewRoleRefs(roles)}
	}
	helpers.Audit(r).Record(r.Context(), event)

	w.WriteHeader(http.StatusNoContent)
}

// removeUser removes a user from a project team
func removeUser(targetUser db.User, member db.ProjectMember, w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user

	if len(member.Roles) == 0 {
		helpers.WriteError(w, &common_errors.ValidationError{Message: "Access comes from group mappings. An admin can change the mappings."})
		return
	}

	event := audit.Event{
		Kind:      audit.IAMMembershipRemove,
		Target:    audit.UserTarget(targetUser.ID, targetUser.Username),
		ProjectID: project.ID,
		Metadata:  audit.MembershipMetadata{Roles: audit.NewRoleRefs(member.Roles), SelfRemoval: targetUser.ID == me.ID},
	}
	// Mapped roles keep the user in the project, so only the manual roles change.
	if len(member.MappedRoles) > 0 {
		event.Kind = audit.IAMProjectRoleChange
		event.Metadata = audit.ProjectRoleMetadata{OldRoles: audit.NewRoleRefs(member.Roles), NewRoles: []audit.RoleRef{}}
	}

	if !me.Admin {
		last, err := takesLastOwner(helpers.Store(r), project.ID, member.Roles, nil)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
		if last {
			event.Outcome = audit.OutcomeFailure
			event.Reason = audit.ReasonLastOwner
			if len(member.MappedRoles) > 0 {
				event.Metadata = audit.ProjectRoleMetadata{OldRoles: audit.NewRoleRefs(member.Roles)}
			}
			helpers.Audit(r).Record(r.Context(), event)
			helpers.WriteError(w, &common_errors.ValidationError{Message: "The project needs at least one Owner"})
			return
		}
	}

	err := helpers.Store(r).DeleteProjectUser(project.ID, targetUser.ID)
	if errors.Is(err, db.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventUser,
		ObjectID:    targetUser.ID,
		Description: fmt.Sprintf("User ID %d removed from team", targetUser.ID),
	})
	helpers.Audit(r).Record(r.Context(), event)

	w.WriteHeader(http.StatusNoContent)
}

// LeftProject removes a user from a project team
func LeftProject(w http.ResponseWriter, r *http.Request) {
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	project := helpers.GetFromContext(r, "project").(db.Project)
	member, err := helpers.Store(r).GetProjectMember(project.ID, me.ID)
	if errors.Is(err, db.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	removeUser(*me, member, w, r)
}

// RemoveUser removes a user from a project team
func RemoveUser(w http.ResponseWriter, r *http.Request) {
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User) // target user
	removeUser(targetUser, helpers.GetFromContext(r, "projectMember").(db.ProjectMember), w, r)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	store := helpers.Store(r)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User)
	oldRoles := helpers.GetFromContext(r, "projectMember").(db.ProjectMember).Roles

	event := audit.Event{
		Kind:      audit.IAMProjectRoleChange,
		Target:    audit.UserTarget(targetUser.ID, targetUser.Username),
		ProjectID: project.ID,
		Metadata:  audit.ProjectRoleMetadata{OldRoles: audit.NewRoleRefs(oldRoles)},
	}

	var projectUser struct {
		Role  string   `json:"role"`
		Roles []string `json:"roles"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	newRoles, err := resolveRoles(store, project.ID, projectUser.Role, projectUser.Roles)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if !me.Admin {
		last, err := takesLastOwner(store, project.ID, oldRoles, newRoles)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
		if last {
			event.Outcome = audit.OutcomeFailure
			event.Reason = audit.ReasonLastOwner
			helpers.Audit(r).Record(r.Context(), event)
			helpers.WriteError(w, &common_errors.ValidationError{Message: "The project needs at least one Owner"})
			return
		}
	}

	err = store.SetProjectUserRoles(project.ID, targetUser.ID, roleIDs(newRoles))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventUser,
		ObjectID:    targetUser.ID,
		Description: fmt.Sprintf("Changed role for User ID %d", targetUser.ID),
	})
	event.Metadata = audit.ProjectRoleMetadata{
		OldRoles: audit.NewRoleRefs(oldRoles),
		NewRoles: audit.NewRoleRefs(newRoles),
	}
	helpers.Audit(r).Record(r.Context(), event)

	w.WriteHeader(http.StatusNoContent)
}
