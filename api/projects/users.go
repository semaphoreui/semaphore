package projects

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func UserMiddleware(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	userID, err := util.GetIntParam("user_id", w, r)
	if err != nil {
		return
	}

	var user db.User
	if err := db.Mysql.SelectOne(&user, "select u.* from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=?", userID, project.ID); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	context.Set(r, "projectUser", user)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var users []struct {
		db.User
		Admin bool `db:"admin" json:"admin"`
	}

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != "asc" && order != "desc" {
		order = "asc"
	}

	q := squirrel.Select("u.*").Column("pu.admin").
			From("project__user as pu").
			LeftJoin("user as u on pu.user_id=u.id").
			Where("pu.project_id=?", project.ID)

	switch sort {
	case "name", "username", "email":
		q = q.OrderBy("u." + sort + " " + order)
	case "admin":
		q = q.OrderBy("pu." + sort + " " + order)
	default:
		q = q.OrderBy("u.name " + order)
	}

	query, args, _ := q.ToSql()

	if _, err := db.Mysql.Select(&users, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, users)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var user struct {
		UserID int  `json:"user_id" binding:"required"`
		Admin  bool `json:"admin"`
	}

	if err := mulekick.Bind(w, r, &user); err != nil {
		return
	}

	if _, err := db.Mysql.Exec("insert into project__user set user_id=?, project_id=?, admin=?", user.UserID, project.ID, user.Admin); err != nil {
		panic(err)
	}

	objType := "user"
	desc := "User ID " + strconv.Itoa(user.UserID) + " added to team"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &user.UserID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "projectUser").(db.User)

	if _, err := db.Mysql.Exec("delete from project__user where user_id=? and project_id=?", user.ID, project.ID); err != nil {
		panic(err)
	}

	objType := "user"
	desc := "User ID " + strconv.Itoa(user.ID) + " removed from team"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &user.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func MakeUserAdmin(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "projectUser").(db.User)
	admin := 1

	if r.Method == "DELETE" {
		// strip admin
		admin = 0
	}

	if _, err := db.Mysql.Exec("update project__user set admin=? where user_id=? and project_id=?", admin, user.ID, project.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
