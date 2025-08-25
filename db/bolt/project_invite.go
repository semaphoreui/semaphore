package bolt

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *BoltDb) GetProjectInvites(projectID int, params db.RetrieveQueryParams) (invites []db.ProjectInviteWithUser, err error) {
	var projectInvites []db.ProjectInvite
	err = d.getObjects(projectID, db.ProjectInviteProps, params, nil, &projectInvites)
	if err != nil {
		return
	}

	for _, invite := range projectInvites {
		var inviteWithUser = db.ProjectInviteWithUser{
			ProjectInvite: invite,
		}

		// Get invited by user info
		invitedByUser, err := d.GetUser(invite.InviterUserID)
		if err == nil {
			inviteWithUser.InvitedByUser = &invitedByUser
		}

		// Get user info if user exists
		if invite.UserID != nil {
			user, err := d.GetUser(*invite.UserID)
			if err == nil {
				inviteWithUser.User = &user
			}
		}

		invites = append(invites, inviteWithUser)
	}

	return
}

func (d *BoltDb) CreateProjectInvite(invite db.ProjectInvite) (db.ProjectInvite, error) {
	newInvite, err := d.createObject(invite.ProjectID, db.ProjectInviteProps, invite)
	if err != nil {
		return db.ProjectInvite{}, err
	}
	return newInvite.(db.ProjectInvite), nil
}

func (d *BoltDb) GetProjectInvite(projectID int, inviteID int) (invite db.ProjectInvite, err error) {
	err = d.getObject(projectID, db.ProjectInviteProps, intObjectID(inviteID), &invite)
	return
}

func (d *BoltDb) GetProjectInviteByToken(token string) (invite db.ProjectInvite, err error) {
	var allInvites []db.ProjectInvite

	// Get all projects to search across all invites
	projects, err := d.GetAllProjects()
	if err != nil {
		return
	}

	for _, project := range projects {
		var projectInvites []db.ProjectInvite
		err = d.getObjects(project.ID, db.ProjectInviteProps, db.RetrieveQueryParams{}, nil, &projectInvites)
		if err != nil {
			continue
		}
		allInvites = append(allInvites, projectInvites...)
	}

	for _, inv := range allInvites {
		if inv.Token == token {
			invite = inv
			return
		}
	}

	err = db.ErrNotFound
	return
}

func (d *BoltDb) UpdateProjectInvite(invite db.ProjectInvite) error {
	return d.updateObject(invite.ProjectID, db.ProjectInviteProps, invite)
}

func (d *BoltDb) DeleteProjectInvite(projectID int, inviteID int) error {
	return d.deleteObject(projectID, db.ProjectInviteProps, intObjectID(inviteID), nil)
}
