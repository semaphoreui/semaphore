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

func (d *SqlDb) DeleteUser(userID int) error {
	res, err := d.exec("delete from `user` where id=?", userID)
	return validateMutationResult(res, err)
}

func (d *SqlDb) UpdateUser(user db.UserWithPwd) error {
	var err error

	if user.Pwd != "" {
		var pwdHash []byte
		pwdHash, err = bcrypt.GenerateFromPassword([]byte(user.Pwd), 11)
		if err != nil {
			return err
		}
		_, err = d.exec(
			"update `user` set name=?, username=?, email=?, alert=?, admin=?, pro=?, password=? where id=?",
			user.Name,
			user.Username,
			user.Email,
			user.Alert,
			user.Admin,
			user.Pro,
			string(pwdHash),
			user.ID)
	} else {
		_, err = d.exec(
			"update `user` set name=?, username=?, email=?, alert=?, admin=?, pro=? where id=?",
			user.Name,
			user.Username,
			user.Email,
			user.Alert,
			user.Admin,
			user.Pro,
			user.ID)
	}

	return err
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
	_, err := d.GetProjectUser(projectID, userID)
	if err == nil {
		return fmt.Errorf("the user is already a project member: %w", db.ErrInvalidOperation)
	}
	if !errors.Is(err, db.ErrNotFound) {
		return err
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

// SetProjectUserRoles replaces the roles of a member. It returns db.ErrNotFound if the user is not a member.
func (d *SqlDb) SetProjectUserRoles(projectID int, userID int, roleIDs []int) error {
	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	res, err := d.execTx(tx, "delete from project__user_role where project_id=? and user_id=?", projectID, userID)
	if err = requireDeletedRow(res, err); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = d.insertProjectUserRoles(tx, projectID, userID, roleIDs); err != nil {
		_ = tx.Rollback()
		return err
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

func (d *SqlDb) GetProjectUserRoles(projectID int, userID int) (roles []db.Role, err error) {
	_, err = d.selectAll(&roles, `
		select
		  r.*
		from role r
		join project__user_role pur on pur.role_id = r.id
		where pur.project_id=? and pur.user_id=?
		order by r.name
	`, projectID, userID)
	return
}

func (d *SqlDb) GetProjectUser(projectID, userID int) (user db.ProjectUser, err error) {
	err = d.selectOne(&user,
		"select project_id, user_id from project__user_role where project_id=? and user_id=? limit 1",
		projectID,
		userID)
	return
}

func (d *SqlDb) GetProjectUsers(projectID int, params db.RetrieveQueryParams) ([]db.UserWithProjectRoles, error) {

	pp, err := params.Validate(db.UserProps)
	if err != nil {
		return nil, err
	}

	q := squirrel.Select("u.*").
		From("`user` as u").
		Where("exists (select 1 from project__user_role pur where pur.project_id = ? and pur.user_id = u.id)", projectID)

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

	result := make([]db.UserWithProjectRoles, 0, len(users))
	for _, u := range users {
		result = append(result, db.UserWithProjectRoles{User: u, Roles: rolesByUser[u.ID]})
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

func escapeLike(s string) string {
	// Order matters: escape \ first
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (d *SqlDb) GetUsers(params db.RetrieveQueryParams) (users []db.User, err error) {
	q := squirrel.Select("*").From("`user`")

	q, err = getQueryForParams(q, "", db.UserProps, params)

	if err != nil {
		return
	}

	if params.Filter != "" {
		q = q.Where(squirrel.Like{"username": escapeLike(params.Filter) + "%"})
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
