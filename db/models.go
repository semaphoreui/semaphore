package db

func SetupDBLink() {
	Mysql.AddTableWithName(APIToken{}, "user__token").SetKeys(false, "id")
	Mysql.AddTableWithName(AccessKey{}, "access_key").SetKeys(true, "id")
	Mysql.AddTableWithName(Environment{}, "project__environment").SetKeys(true, "id")
	Mysql.AddTableWithName(Inventory{}, "project__inventory").SetKeys(true, "id")
	Mysql.AddTableWithName(Project{}, "project").SetKeys(true, "id")
	Mysql.AddTableWithName(Repository{}, "project__repository").SetKeys(true, "id")
	Mysql.AddTableWithName(Task{}, "task").SetKeys(true, "id")
	Mysql.AddTableWithName(TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
	Mysql.AddTableWithName(Template{}, "project__template").SetKeys(true, "id")
	Mysql.AddTableWithName(User{}, "user").SetKeys(true, "id")
	Mysql.AddTableWithName(Session{}, "session").SetKeys(true, "id")
}
