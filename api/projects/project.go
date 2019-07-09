package projects

import (
	"database/sql"
	"net/http"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/mulekick"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

// ProjectMiddleware ensures a project exists and loads it to the context
func ProjectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.Get(r, "user").(*db.User)

		projectID, err := util.GetIntParam("project_id", w, r)
		if err != nil {
			return
		}

		query, args, err := squirrel.Select("p.*").
			From("project as p").
			Join("project__user as pu on pu.project_id=p.id").
			Where("p.id=?", projectID).
			Where("pu.user_id=?", user.ID).
			ToSql()
		util.LogWarning(err)

		var project db.Project
		if err := db.Mysql.SelectOne(&project, query, args...); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			panic(err)
		}

		context.Set(r, "project", project)
		if (next != nil) {
      next.ServeHTTP(w, r)
    }
	})
}

//GetProject returns a project details
func GetProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulekick.WriteJSON(w, http.StatusOK, context.Get(r, "project"))

		if (next != nil) {
      next.ServeHTTP(w, r)
    }
	})
}

// MustBeAdmin ensures that the user has administrator rights
func MustBeAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		user := context.Get(r, "user").(*db.User)

		userC, err := db.Mysql.SelectInt("select count(1) from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=? and pu.admin=1", user.ID, project.ID)
		if err != nil {
			panic(err)
		}

		if userC == 0 {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if (next != nil) {
      next.ServeHTTP(w, r)
    }
	})
}

// UpdateProject saves updated project details to the database
func UpdateProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		var body struct {
			Name      string `json:"name"`
			Alert     bool   `json:"alert"`
			AlertChat string `json:"alert_chat"`
		}

		if err := mulekick.Bind(w, r, &body); err != nil {
			return
		}

		if _, err := db.Mysql.Exec("update project set name=?, alert=?, alert_chat=? where id=?", body.Name, body.Alert, body.AlertChat, project.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)

		if (next != nil) {
      next.ServeHTTP(w, r)
    }
	})
}

// DeleteProject removes a project from the database
func DeleteProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)

		tx, err := db.Mysql.Begin()
		if err != nil {
			panic(err)
		}

		statements := []string{
			"delete tao from task__output as tao join task as t on t.id=tao.task_id join project__template as pt on pt.id=t.template_id where pt.project_id=?",
			"delete t from task as t join project__template as pt on pt.id=t.template_id where pt.project_id=?",
			"delete from project__template where project_id=?",
			"delete from project__user where project_id=?",
			"delete from project__repository where project_id=?",
			"delete from project__inventory where project_id=?",
			"delete from access_key where project_id=?",
			"delete from project where id=?",
		}

		for _, statement := range statements {
			_, err := tx.Exec(statement, project.ID)

			if err != nil {
				err = tx.Rollback()
				util.LogWarning(err)
				panic(err)
			}
		}

		if err := tx.Commit(); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)

		if (next != nil) {
      next.ServeHTTP(w, r)
    }
	})
}
