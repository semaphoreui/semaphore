package db

// SetupDBLink is called by main after initialization of the Sql object to create or return an existing table map
func SetupDBLink() {
	Sql.AddTableWithName(APIToken{}, "user__token").SetKeys(false, "id")
	Sql.AddTableWithName(AccessKey{}, "access_key").SetKeys(true, "id")
	Sql.AddTableWithName(Environment{}, "project__environment").SetKeys(true, "id")
	Sql.AddTableWithName(Inventory{}, "project__inventory").SetKeys(true, "id")
	Sql.AddTableWithName(Project{}, "project").SetKeys(true, "id")
	Sql.AddTableWithName(Repository{}, "project__repository").SetKeys(true, "id")
	Sql.AddTableWithName(Task{}, "task").SetKeys(true, "id")
	Sql.AddTableWithName(TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
	Sql.AddTableWithName(Template{}, "project__template").SetKeys(true, "id")
	Sql.AddTableWithName(User{}, "user").SetKeys(true, "id")
	Sql.AddTableWithName(Session{}, "session").SetKeys(true, "id")
}
