package export

import "github.com/semaphoreui/semaphore/db"

type ProjectExporter struct {
	ValueMap[db.Project]
}

func (a *ProjectExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	allKeys := make([]db.Project, 0)

	users, err := exporter.getLoadedKeysInt(User, GlobalScope)
	if err != nil {
		return err
	}

	ids := make(map[int]bool)

	for _, userId := range users {
		projects, err := store.GetProjects(userId)
		if err != nil {
			return err
		}

		for _, proj := range projects {
			if ids[proj.ID] {
				continue
			}
			ids[proj.ID] = true
			allKeys = append(allKeys, proj)
		}
	}

	return a.appendValues(allKeys, GlobalScope)
}

func (a *ProjectExporter) restore(store db.Store, exporter DataExporter, progress Progress) error {
	for _, val := range a.values {
		old := val.value

		obj, err := store.CreateProject(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(a.getName(), GlobalScope, old.ID, obj.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ProjectExporter) exportDependsOn() []string {
	return []string{User}
}

func (a *ProjectExporter) getName() string {
	return Project
}
