package models

import "github.com/ansible-semaphore/semaphore/db"

func SetupDBLink() {
	db.Mysql.AddTableWithName(APIToken{}, "user__token").SetKeys(false, "id")
	db.Mysql.AddTableWithName(AccessKey{}, "access_key").SetKeys(true, "id")
	db.Mysql.AddTableWithName(Environment{}, "project__environment").SetKeys(true, "id")
	db.Mysql.AddTableWithName(Inventory{}, "project__inventory").SetKeys(true, "id")
	db.Mysql.AddTableWithName(Project{}, "project").SetKeys(true, "id")
	db.Mysql.AddTableWithName(Repository{}, "project__repository").SetKeys(true, "id")
	db.Mysql.AddTableWithName(Task{}, "task").SetKeys(true, "id")
	db.Mysql.AddTableWithName(TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
	db.Mysql.AddTableWithName(Template{}, "project__template").SetKeys(true, "id")
	db.Mysql.AddTableWithName(User{}, "user").SetKeys(true, "id")
	db.Mysql.AddTableWithName(Session{}, "session").SetKeys(true, "id")
}
