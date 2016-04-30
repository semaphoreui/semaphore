package models

import "github.com/ansible-semaphore/semaphore/database"

func SetupDBLink() {
	database.Mysql.AddTableWithName(APIToken{}, "user__token").SetKeys(false, "id")
	database.Mysql.AddTableWithName(AccessKey{}, "access_key").SetKeys(true, "id")
	database.Mysql.AddTableWithName(Environment{}, "project__environment").SetKeys(true, "id")
	database.Mysql.AddTableWithName(Inventory{}, "project__inventory").SetKeys(true, "id")
	database.Mysql.AddTableWithName(Project{}, "project").SetKeys(true, "id")
	database.Mysql.AddTableWithName(Repository{}, "project__repository").SetKeys(true, "id")
	database.Mysql.AddTableWithName(Task{}, "task").SetKeys(true, "id")
	database.Mysql.AddTableWithName(TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
	database.Mysql.AddTableWithName(Template{}, "project__template").SetKeys(true, "id")
	database.Mysql.AddTableWithName(User{}, "user").SetKeys(true, "id")
	database.Mysql.AddTableWithName(Session{}, "session").SetKeys(true, "id")
}
