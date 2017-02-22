package projects

import (
	"database/sql"
	"strconv"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func UserMiddleware(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	userID, err := util.GetIntParam("user_id", c)
	if err != nil {
		return
	}

	var user models.User
	if err := database.Mysql.SelectOne(&user, "select u.* from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=?", userID, project.ID); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	c.Set("projectUser", user)
	c.Next()
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var users []struct {
		models.User
		Admin bool `db:"admin" json:"admin"`
	}

	query, args, _ := squirrel.Select("u.*").Column("pu.admin").
		From("project__user as pu").
		Join("user as u on pu.user_id=u.id").
		Where("pu.project_id=?", project.ID).
		ToSql()

	if _, err := database.Mysql.Select(&users, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, users)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var user struct {
		UserID int  `json:"user_id" binding:"required"`
		Admin  bool `json:"admin"`
	}

	if err := mulekick.Bind(w, r, &user); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("insert into project__user set user_id=?, project_id=?, admin=?", user.UserID, project.ID, user.Admin); err != nil {
		panic(err)
	}

	objType := "user"
	desc := "User ID " + strconv.Itoa(user.UserID) + " added to team"
	if err := (models.Event{
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
	project := context.Get(r, "project").(models.Project)
	user := context.Get(r, "projectUser").(models.User)

	if _, err := database.Mysql.Exec("delete from project__user where user_id=? and project_id=?", user.ID, project.ID); err != nil {
		panic(err)
	}

	objType := "user"
	desc := "User ID " + strconv.Itoa(user.ID) + " removed from team"
	if err := (models.Event{
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
	project := context.Get(r, "project").(models.Project)
	user := context.Get(r, "projectUser").(models.User)
	admin := 1

	if r.Method == "DELETE" {
		// strip admin
		admin = 0
	}

	if _, err := database.Mysql.Exec("update project__user set admin=? where user_id=? and project_id=?", admin, user.ID, project.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
