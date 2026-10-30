package sql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"golang.org/x/crypto/bcrypt"
)

func (d *SqlDb) CreateUserWithoutPassword(user db.User) (newUser db.User, err error) {

	err = db.ValidateUser(user)
	if err != nil {
		return
	}

	user.Password = ""
	user.Created = db.GetParsedTime(tz.Now())

	err = d.Sql().Insert(&user)

	if err != nil {
		return
	}

	newUser = user
	return
}

func (d *SqlDb) CreateUser(user db.UserWithPwd) (newUser db.User, err error) {

	err = db.ValidateUser(user.User)
	if err != nil {
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), 11)

	if err != nil {
		return
	}

	user.Password = string(pwdHash)
	user.Created = db.GetParsedTime(tz.Now())

	err = d.Sql().Insert(&user.User)

	if err != nil {
		return
	}

	newUser = user.User
	return
}

func (d *SqlDb) ImportUser(user db.UserWithPwd) (newUser db.User, err error) {
	err = db.ValidateUser(user.User)
	if err != nil {
		return
	}

	user.Created = db.GetParsedTime(tz.Now())

	err = d.Sql().Insert(&user.User)

	if err != nil {
		return
	}

	newUser = user.User
	return
}

// DeleteUser removes the user, unless it is the last admin.
func (d *SqlDb) DeleteUser(userID int) error {
	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = d.LockAdministratorState(tx); err != nil {
		return err
	}
	if err = d.requireAnotherAdmin(tx, userID); err != nil {
		return err
	}
	res, err := d.execTx(tx, "delete from `user` where id=?", userID)
	if err = validateMutationResult(res, err); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateUser saves the user. Clearing the admin flag of the last admin is refused.
func (d *SqlDb) UpdateUser(user db.UserWithPwd) error {
	var pwdHash []byte
	if user.Pwd != "" {
		var err error
		pwdHash, err = bcrypt.GenerateFromPassword([]byte(user.Pwd), 11)
		if err != nil {
			return err
		}
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = d.LockAdministratorState(tx); err != nil {
		return err
	}
	if !user.Admin {
		if err = d.requireAnotherAdmin(tx, user.ID); err != nil {
			return err
		}
	}

	if len(pwdHash) > 0 {
		_, err = d.execTx(tx,
			"update `user` set name=?, username=?, email=?, alert=?, admin=?, pro=?, password=? where id=?",
			user.Name, user.Username, user.Email, user.Alert, user.Admin, user.Pro, string(pwdHash), user.ID)
	} else {
		_, err = d.execTx(tx,
			"update `user` set name=?, username=?, email=?, alert=?, admin=?, pro=? where id=?",
			user.Name, user.Username, user.Email, user.Alert, user.Admin, user.Pro, user.ID)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (d *SqlDb) SetUserPassword(userID int, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return err
	}
	_, err = d.exec(
		"update `user` set password=? where id=?",
		string(hash), userID)
	return err
}

func (d *SqlDb) CreateProjectUser(projectID int, userID int, roleIDs []int) error {
	if len(roleIDs) == 0 {
		return &common_errors.ValidationError{Message: "A member needs at least one role"}
	}
	// A user with roles only from group mappings can get manual roles here.
	manual, err := d.Sql().SelectInt(d.PrepareQuery("select count(*) from project__user_role where project_id=? and user_id=?"), projectID, userID)
	if err != nil {
		return err
	}
	if manual > 0 {
		return fmt.Errorf("the user is already a project member: %w", db.ErrInvalidOperation)
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	if err = d.insertProjectUserRoles(tx, projectID, userID, roleIDs); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// SetProjectUserRoles replaces the manual roles of a member. The list may be empty while the member keeps roles from
// group mappings. It returns db.ErrNotFound if the user is not a member.
func (d *SqlDb) SetProjectUserRoles(projectID int, userID int, roleIDs []int) (err error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	res, err := d.execTx(tx, "delete from project__user_role where project_id=? and user_id=?", projectID, userID)
	if err != nil {
		return
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return
	}
	mapped, err := tx.SelectInt(d.PrepareQuery("select count(*) from project__user_mapped_role where project_id=? and user_id=?"), projectID, userID)
	if err != nil {
		return
	}
	if deleted == 0 && mapped == 0 {
		// A member removed meanwhile must not come back.
		return db.ErrNotFound
	}
	if len(roleIDs) == 0 && mapped == 0 {
		return &common_errors.ValidationError{Message: "A member needs at least one role"}
	}
	if err = d.insertProjectUserRoles(tx, projectID, userID, roleIDs); err != nil {
		return
	}
	return tx.Commit()
}

func (d *SqlDb) insertProjectUserRoles(tx *gorp.Transaction, projectID int, userID int, roleIDs []int) error {
	query, args, err := squirrel.Select("count(*)").From("`role`").
		Where(squirrel.Eq{"id": roleIDs}).
		Where("(project_id is null or project_id=?)", projectID).
		ToSql()
	if err != nil {
		return err
	}
	available, err := tx.SelectInt(d.PrepareQuery(query), args...)
	if err != nil {
		return err
	}
	if int(available) != len(roleIDs) {
		return &common_errors.ValidationError{Message: "Role is not available in this project"}
	}
	for _, roleID := range roleIDs {
		_, err = d.execTx(tx, "insert into project__user_role (project_id, user_id, role_id) values (?, ?, ?)", projectID, userID, roleID)
		if err != nil {
			return err
		}
	}
	return nil
}

// MappedRolesQuery selects db.MappedRole rows, each caller adds its filter and order.
func MappedRolesQuery() squirrel.SelectBuilder {
	return squirrel.Select("mr.project_id", "mr.user_id", "u.username", "mr.mapping_id", "m.provider", "m.group_name",
		"r.id as role_id", "r.slug as role_slug", "r.name as role_name", "r.permissions as role_permissions").
		From("project__user_mapped_role as mr").
		Join("group_role_mapping as m on m.id = mr.mapping_id").
		Join("`role` as r on r.id = m.role_id").
		Join("`user` as u on u.id = mr.user_id")
}

// GetProjectMember returns the manual and the mapped roles of a member, db.ErrNotFound if the user holds neither.
func (d *SqlDb) GetProjectMember(projectID int, userID int) (member db.ProjectMember, err error) {
	_, err = d.selectAll(&member.Roles, `
		select
		  r.*
		from role r
		join project__user_role pur on pur.role_id = r.id
		where pur.project_id = ? and pur.user_id = ?
		order by r.name
	`, projectID, userID)
	if err != nil {
		return
	}
	query, args, err := MappedRolesQuery().
		Where("mr.project_id = ? and mr.user_id = ?", projectID, userID).
		OrderBy("r.name", "m.provider", "m.group_name").
		ToSql()
	if err != nil {
		return
	}
	if _, err = d.selectAll(&member.MappedRoles, query, args...); err != nil {
		return
	}
	if len(member.Roles) == 0 && len(member.MappedRoles) == 0 {
		err = db.ErrNotFound
	}
	return
}

// CountProjectOwners counts manual Owner roles, mapped ones can vanish at the next login.
func (d *SqlDb) CountProjectOwners(projectID int) (int, error) {
	count, err := d.Sql().SelectInt(d.PrepareQuery(`
		select
		  count(*)
		from project__user_role pur
		join role r on r.id = pur.role_id
		where pur.project_id=? and r.builtin=true and r.slug=?
	`), projectID, string(db.ProjectOwner))
	return int(count), err
}

func (d *SqlDb) GetProjectUser(projectID, userID int) (user db.ProjectUser, err error) {
	err = d.selectOne(&user, `
		select project_id, user_id from project__user_role where project_id = ? and user_id = ?
		union
		select project_id, user_id from project__user_mapped_role where project_id = ? and user_id = ?
		limit 1
	`, projectID, userID, projectID, userID)
	return
}

func (d *SqlDb) GetProjectUsers(projectID int, params db.RetrieveQueryParams) ([]db.UserWithProjectRoles, error) {

	pp, err := params.Validate(db.UserProps)
	if err != nil {
		return nil, err
	}

	q := squirrel.Select("u.*").
		From("`user` as u").
		Where(`(
			exists (select 1 from project__user_role pur where pur.project_id = ? and pur.user_id = u.id)
			or exists (select 1 from project__user_mapped_role mr where mr.project_id = ? and mr.user_id = u.id)
		)`, projectID, projectID)

	sortDirection := "ASC"
	if pp.SortInverted {
		sortDirection = "DESC"
	}

	switch pp.SortBy {
	case "name", "username", "email":
		q = q.OrderBy("u." + pp.SortBy + " " + sortDirection)
	case "role":
		q = q.OrderByClause(`(
			select
			  min(r.name)
			from project__user_role pur
			join role r on r.id = pur.role_id
			where pur.project_id = ? and pur.user_id = u.id
		) `+sortDirection, projectID)
	default:
		q = q.OrderBy("u.name " + sortDirection)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	var users []db.User
	if _, err = d.selectAll(&users, query, args...); err != nil {
		return nil, err
	}

	var memberRoles []struct {
		UserID int `db:"user_id"`
		db.Role
	}
	_, err = d.selectAll(&memberRoles, `
		select
		  pur.user_id, r.*
		from project__user_role pur
		join role r on r.id = pur.role_id
		where pur.project_id=?
		order by r.name
	`, projectID)
	if err != nil {
		return nil, err
	}

	rolesByUser := make(map[int][]db.Role)
	for _, mr := range memberRoles {
		rolesByUser[mr.UserID] = append(rolesByUser[mr.UserID], mr.Role)
	}

	mappedQuery, mappedArgs, err := MappedRolesQuery().
		Where("mr.project_id = ?", projectID).
		OrderBy("r.name", "m.provider", "m.group_name").
		ToSql()
	if err != nil {
		return nil, err
	}
	var mapped []db.MappedRole
	if _, err = d.selectAll(&mapped, mappedQuery, mappedArgs...); err != nil {
		return nil, err
	}
	mappedByUser := make(map[int][]db.MappedRole)
	for _, m := range mapped {
		mappedByUser[m.UserID] = append(mappedByUser[m.UserID], m)
	}

	result := make([]db.UserWithProjectRoles, 0, len(users))
	for _, u := range users {
		result = append(result, db.UserWithProjectRoles{User: u, Roles: rolesByUser[u.ID], MappedRoles: mappedByUser[u.ID]})
	}

	return result, nil
}

func (d *SqlDb) DeleteProjectUser(projectID, userID int) error {
	res, err := d.exec("delete from project__user_role where project_id=? and user_id=?", projectID, userID)
	return requireDeletedRow(res, err)
}

// GetUser retrieves a user from the database by ID
func (d *SqlDb) GetUser(userID int) (user db.User, err error) {

	err = d.selectOne(&user, "select * from `user` where id=?", userID)

	if err != nil {
		return
	}

	var totp db.UserTotp
	err = d.selectOne(&totp, "select * from `user__totp` where user_id=?", user.ID)

	if err == nil {
		user.Totp = &totp
	}

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	if err != nil {
		return
	}

	var emailOtp db.UserEmailOtp
	err = d.selectOne(&emailOtp, "select * from `user__email_otp` where user_id=?", user.ID)

	if err == nil {
		user.EmailOtp = &emailOtp
	}

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	return
}

func (d *SqlDb) GetProUserCount() (count int, err error) {

	cnt, err := d.Sql().SelectInt(d.PrepareQuery("select count(*) from `user` where pro"))

	count = int(cnt)

	return
}

func (d *SqlDb) GetUserCount() (count int, err error) {

	cnt, err := d.Sql().SelectInt(d.PrepareQuery("select count(*) from `user`"))

	count = int(cnt)

	return
}

// escapeLike makes s match itself literally inside a LIKE pattern. The query
// must name the escape character with `ESCAPE '!'`: SQLite has no default one,
// so a backslash is taken literally there, and `!` is also written the same way
// in a MySQL and a PostgreSQL string literal, unlike a backslash.
func escapeLike(s string) string {
	// Order matters: the escape character itself goes first.
	s = strings.ReplaceAll(s, "!", "!!")
	s = strings.ReplaceAll(s, "%", "!%")
	s = strings.ReplaceAll(s, "_", "!_")
	return s
}

func (d *SqlDb) GetUsers(params db.RetrieveQueryParams) (users []db.User, err error) {
	q := squirrel.Select("*").From("`user`")

	q, err = getQueryForParams(q, "", db.UserProps, params)

	if err != nil {
		return
	}

	if params.Filter != "" {
		q = q.Where("username LIKE ? ESCAPE '!'", escapeLike(params.Filter)+"%")
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&users, query, args...)

	return
}

func (d *SqlDb) GetUserByLoginOrEmail(login string, email string) (user db.User, err error) {
	err = d.selectOne(
		&user,
		d.PrepareQuery("select * from `user` where email=? or username=?"),
		email, login)

	if err != nil {
		return
	}

	var totp db.UserTotp
	err = d.selectOne(&totp, "select * from `user__totp` where user_id=?", user.ID)

	if err == nil {
		user.Totp = &totp
	}

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	if err != nil {
		return
	}

	var emailOtp db.UserEmailOtp
	err = d.selectOne(&emailOtp, "select * from `user__email_otp` where user_id=?", user.ID)

	if err == nil && !emailOtp.IsExpired() {
		user.EmailOtp = &emailOtp
	}

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	return
}

func (d *SqlDb) GetAllAdmins() (users []db.User, err error) {
	_, err = d.selectAll(&users, "select * from `user` where `admin` = true")

	return
}

func (d *SqlDb) AddTotpVerification(userID int, url string, recoveryHash string) (totp db.UserTotp, err error) {

	totp.UserID = userID
	totp.URL = url
	totp.RecoveryHash = recoveryHash
	totp.Created = db.GetParsedTime(tz.Now())

	res, err := d.exec(
		"insert into user__totp (user_id, url, recovery_hash, created) values (?, ?, ?, ?)",
		totp.UserID,
		totp.URL,
		totp.RecoveryHash,
		totp.Created)

	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		return
	}

	totp.ID = int(id)

	return
}

func (d *SqlDb) DeleteTotpVerification(userID int, totpID int) error {
	res, err := d.exec("delete from user__totp where user_id=? and id = ?", userID, totpID)
	return requireDeletedRow(res, err)
}

func (d *SqlDb) insertEmailOtp(userID int, code string) (totp db.UserEmailOtp, err error) {

	totp.UserID = userID
	totp.Code = code
	totp.Created = db.GetParsedTime(tz.Now())

	res, err := d.exec(
		"insert into user__email_otp (user_id, code, created) values (?, ?, ?)",
		totp.UserID,
		totp.Code,
		totp.Created)

	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		return
	}

	totp.ID = int(id)

	return
}

func (d *SqlDb) AddEmailOtpVerification(userID int, code string) (res db.UserEmailOtp, err error) {

	var emailOtp db.UserEmailOtp
	err = d.selectOne(&emailOtp, "select * from `user__email_otp` where user_id=?", userID)

	if err == nil {
		now := db.GetParsedTime(tz.Now())
		_, err = d.exec("update user__email_otp set code=?, created=?, attempts=0 where user_id=?", code, now, userID)
	} else if errors.Is(err, db.ErrNotFound) {
		res, err = d.insertEmailOtp(userID, code)
	} else {
		return
	}

	return
}

func (d *SqlDb) IncrementEmailOtpAttempts(userID int) error {
	_, err := d.exec("update user__email_otp set attempts = attempts + 1 where user_id=?", userID)
	return err
}

func (d *SqlDb) DeleteEmailOtpVerification(userID int, totpID int) error {
	_, err := d.exec("delete from user__email_otp where user_id=? and id = ?", userID, totpID)
	return err
}
