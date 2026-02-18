package sql

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetProjectInvites(projectID int, params db.RetrieveQueryParams) (invites []db.ProjectInviteWithUser, err error) {
	pp, err := params.Validate(db.ProjectInviteProps)
	if err != nil {
		return
	}

	invites = make([]db.ProjectInviteWithUser, 0)

	q := squirrel.Select("pi.*").
		Column("ib.name as inviter_user_id_name").
		Column("ib.username as inviter_username").
		Column("ib.email as inviter_user_id_email").
		Column("u.name as user_name").
		Column("u.username as user_username").
		Column("u.email as user_email").
		From("project__invite as pi").
		LeftJoin("`user` as ib on pi.inviter_user_id=ib.id").
		LeftJoin("`user` as u on pi.user_id=u.id").
		Where("pi.project_id=?", projectID)

	sortDirection := "ASC"
	if pp.SortInverted {
		sortDirection = "DESC"
	}

	switch pp.SortBy {
	case "created", "status", "role":
		q = q.OrderBy("pi." + pp.SortBy + " " + sortDirection)
	default:
		q = q.OrderBy("pi.created " + sortDirection)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	rows, err := d.Sql().Query(d.PrepareQuery(query), args...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var invite db.ProjectInviteWithUser
		var invitedByName, invitedByUsername, invitedByEmail sql.NullString
		var userName, userUsername, userEmail sql.NullString

		err = rows.Scan(
			&invite.ID,
			&invite.ProjectID,
			&invite.UserID,
			&invite.Email,
			&invite.Role,
			&invite.Status,
			&invite.Token,
			&invite.InviterUserID,
			&invite.Created,
			&invite.ExpiresAt,
			&invite.AcceptedAt,
			&invitedByName,
			&invitedByUsername,
			&invitedByEmail,
			&userName,
			&userUsername,
			&userEmail,
		)
		if err != nil {
			return
		}

		// Set invited by user info
		invite.InvitedByUser = &db.User{
			ID:       invite.InviterUserID,
			Name:     invitedByName.String,
			Username: invitedByUsername.String,
			Email:    invitedByEmail.String,
		}

		// Set user info if user exists
		if invite.UserID != nil {
			invite.User = &db.User{
				ID:       *invite.UserID,
				Name:     userName.String,
				Username: userUsername.String,
				Email:    userEmail.String,
			}
		}

		invites = append(invites, invite)
	}

	return
}

func (d *SqlDb) CreateProjectInvite(invite db.ProjectInvite) (newInvite db.ProjectInvite, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__invite (project_id, user_id, email, role, status, token, inviter_user_id, created, expires_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		invite.ProjectID,
		invite.UserID,
		invite.Email,
		invite.Role,
		invite.Status,
		invite.Token,
		invite.InviterUserID,
		invite.Created,
		invite.ExpiresAt)

	if err != nil {
		return
	}

	newInvite = invite
	newInvite.ID = insertID
	return
}

func (d *SqlDb) GetProjectInvite(projectID int, inviteID int) (invite db.ProjectInvite, err error) {
	err = d.selectOne(&invite,
		"select * from project__invite where project_id=? and id=?",
		projectID,
		inviteID)
	return
}

func (d *SqlDb) GetProjectInviteByToken(token string) (invite db.ProjectInvite, err error) {
	err = d.selectOne(&invite,
		"select * from project__invite where token=?",
		token)
	return
}

func (d *SqlDb) UpdateProjectInvite(invite db.ProjectInvite) error {
	_, err := d.exec(
		"update project__invite set status=?, accepted_at=? where id=?",
		invite.Status,
		invite.AcceptedAt,
		invite.ID)
	return err
}

func (d *SqlDb) DeleteProjectInvite(projectID int, inviteID int) error {
	_, err := d.exec(
		"delete from project__invite where project_id=? and id=?",
		projectID,
		inviteID)
	return err
}
