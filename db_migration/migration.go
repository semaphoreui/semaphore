package db_migration

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

type Migrator struct {
	oldStore db.Store
	newStore db.Store

	userIDs        map[int]int
	projectIDs     map[int]int
	keyIDs         map[int]int
	repositoryIDs  map[int]int
	inventoryIDs   map[int]int
	templateIDs    map[int]int
	viewIDs        map[int]int
	environmentIDs map[int]int
}

func (m *Migrator) Migrate(oldStore, newStore db.Store) error {
	m.oldStore = oldStore
	m.newStore = newStore

	m.userIDs = make(map[int]int)
	m.projectIDs = make(map[int]int)
	m.keyIDs = make(map[int]int)
	m.repositoryIDs = make(map[int]int)
	m.inventoryIDs = make(map[int]int)
	m.templateIDs = make(map[int]int)
	m.viewIDs = make(map[int]int)
	m.environmentIDs = make(map[int]int)

	fmt.Println("Migrating users...")
	if err := m.migrateUsers(); err != nil {
		return err
	}

	fmt.Println("Migrating projects...")
	if err := m.migrateProjects(); err != nil {
		return err
	}

	fmt.Println("Migration completed successfully.")
	return nil
}

func (m *Migrator) migrateUsers() error {
	users, err := m.oldStore.GetUsers(db.RetrieveQueryParams{})
	if err != nil {
		return err
	}

	for _, user := range users {
		oldID := user.ID
		user.ID = 0
		newUser, err := m.newStore.ImportUser(db.UserWithPwd{Pwd: user.Password, User: user})
		if err != nil {
			return err
		}
		m.userIDs[oldID] = newUser.ID
	}
	return nil
}

func (m *Migrator) migrateProjects() error {
	allProjects, err := m.oldStore.GetAllProjects()
	if err != nil {
		// Fallback for older versions if GetAllProjects is not available or fails
		allProjects = make([]db.Project, 0)
		for oldUserID := range m.userIDs {
			projects, err := m.oldStore.GetProjects(oldUserID)
			if err != nil {
				return err
			}
			for _, project := range projects {
				found := false
				for _, p := range allProjects {
					if p.ID == project.ID {
						found = true
						break
					}
				}
				if !found {
					allProjects = append(allProjects, project)
				}
			}
		}
	}

	for _, project := range allProjects {
		oldID := project.ID
		project.ID = 0
		newProject, err := m.newStore.CreateProject(project)
		if err != nil {
			return err
		}
		m.projectIDs[oldID] = newProject.ID

		if err := m.migrateProjectData(oldID, newProject.ID); err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) migrateProjectData(oldProjectID, newProjectID int) error {
	if err := m.migrateProjectUsers(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateKeys(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateEnvironments(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateRepositories(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateInventories(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateViews(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateTemplates(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateTasks(oldProjectID, newProjectID); err != nil {
		return err
	}
	if err := m.migrateSchedules(oldProjectID, newProjectID); err != nil {
		return err
	}
	return nil
}

func (m *Migrator) migrateProjectUsers(oldProjectID, newProjectID int) error {
	users, err := m.oldStore.GetProjectUsers(oldProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}
	for _, user := range users {
		newUserID, ok := m.userIDs[user.ID]
		if !ok {
			continue
		}
		_, err = m.newStore.CreateProjectUser(db.ProjectUser{
			ProjectID: newProjectID,
			UserID:    newUserID,
			Role:      user.Role,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) migrateKeys(oldProjectID, newProjectID int) error {
	keys, err := m.oldStore.GetAccessKeys(oldProjectID, db.GetAccessKeyOptions{}, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}
	for _, key := range keys {
		oldID := key.ID
		key.ID = 0
		key.ProjectID = &newProjectID
		newKey, err := m.newStore.CreateAccessKey(key)
		if err != nil {
			return err
		}
		m.keyIDs[oldID] = newKey.ID
	}
	return nil
}

func (m *Migrator) migrateEnvironments(oldProjectID, newProjectID int) error {
	envs, err := m.oldStore.GetEnvironments(oldProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}
	for _, env := range envs {
		oldID := env.ID
		env.ID = 0
		env.ProjectID = newProjectID
		newEnv, err := m.newStore.CreateEnvironment(env)
		if err != nil {
			return err
		}
		m.environmentIDs[oldID] = newEnv.ID
	}
	return nil
}

func (m *Migrator) migrateRepositories(oldProjectID, newProjectID int) error {
	repos, err := m.oldStore.GetRepositories(oldProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}
	for _, repo := range repos {
		oldID := repo.ID
		repo.ID = 0
		repo.ProjectID = newProjectID
		if newKeyID, ok := m.keyIDs[repo.SSHKeyID]; ok {
			repo.SSHKeyID = newKeyID
		}
		newRepo, err := m.newStore.CreateRepository(repo)
		if err != nil {
			return err
		}
		m.repositoryIDs[oldID] = newRepo.ID
	}
	return nil
}

func (m *Migrator) migrateInventories(oldProjectID, newProjectID int) error {
	inventories, err := m.oldStore.GetInventories(oldProjectID, db.RetrieveQueryParams{}, nil)
	if err != nil {
		return err
	}
	for _, inv := range inventories {
		oldID := inv.ID
		inv.ID = 0
		inv.ProjectID = newProjectID
		if inv.SSHKeyID != nil {
			if newKeyID, ok := m.keyIDs[*inv.SSHKeyID]; ok {
				inv.SSHKeyID = &newKeyID
			}
		}
		if inv.BecomeKeyID != nil {
			if newKeyID, ok := m.keyIDs[*inv.BecomeKeyID]; ok {
				inv.BecomeKeyID = &newKeyID
			}
		}
		if inv.RepositoryID != nil {
			if newRepoID, ok := m.repositoryIDs[*inv.RepositoryID]; ok {
				inv.RepositoryID = &newRepoID
			}
		}
		newInv, err := m.newStore.CreateInventory(inv)
		if err != nil {
			return err
		}
		m.inventoryIDs[oldID] = newInv.ID
	}
	return nil
}

func (m *Migrator) migrateViews(oldProjectID, newProjectID int) error {
	views, err := m.oldStore.GetViews(oldProjectID)
	if err != nil {
		return err
	}
	for _, view := range views {
		oldID := view.ID
		view.ID = 0
		view.ProjectID = newProjectID
		newView, err := m.newStore.CreateView(view)
		if err != nil {
			return err
		}
		m.viewIDs[oldID] = newView.ID
	}
	return nil
}

func (m *Migrator) migrateTemplates(oldProjectID, newProjectID int) error {
	templates, err := m.oldStore.GetTemplates(oldProjectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}
	for _, tpl := range templates {
		oldID := tpl.ID
		tpl.ID = 0
		tpl.ProjectID = newProjectID

		if tpl.InventoryID != nil {
			if newInvID, ok := m.inventoryIDs[*tpl.InventoryID]; ok {
				tpl.InventoryID = &newInvID
			}
		}
		if newRepoID, ok := m.repositoryIDs[tpl.RepositoryID]; ok {
			tpl.RepositoryID = newRepoID
		}
		if tpl.ViewID != nil {
			if newViewID, ok := m.viewIDs[*tpl.ViewID]; ok {
				tpl.ViewID = &newViewID
			}
		}
		if tpl.EnvironmentID != nil {
			if newEnvID, ok := m.environmentIDs[*tpl.EnvironmentID]; ok {
				tpl.EnvironmentID = &newEnvID
			}
		}

		newTpl, err := m.newStore.CreateTemplate(tpl)
		if err != nil {
			return err
		}
		m.templateIDs[oldID] = newTpl.ID
	}
	return nil
}

func (m *Migrator) migrateTasks(oldProjectID, newProjectID int) error {
	tasks, err := m.oldStore.GetProjectTasks(oldProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}
	for _, taskWithTpl := range tasks {
		task := taskWithTpl.Task
		oldID := task.ID
		task.ID = 0
		task.ProjectID = newProjectID
		if newTplID, ok := m.templateIDs[task.TemplateID]; ok {
			task.TemplateID = newTplID
		}
		if task.InventoryID != nil {
			if newInvID, ok := m.inventoryIDs[*task.InventoryID]; ok {
				task.InventoryID = &newInvID
			}
		}

		if task.UserID != nil {
			if newUserID, ok := m.userIDs[*task.UserID]; ok {
				task.UserID = &newUserID
			}
		}

		newTask, err := m.newStore.CreateTask(task, 0)
		if err != nil {
			return err
		}

		// Migrate task outputs
		outputs, err := m.oldStore.GetTaskOutputs(oldProjectID, oldID, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}
		for _, output := range outputs {
			output.TaskID = newTask.ID
			_, err = m.newStore.CreateTaskOutput(output)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Migrator) migrateSchedules(oldProjectID, newProjectID int) error {
	schedules, err := m.oldStore.GetProjectSchedules(oldProjectID, true, true)
	if err != nil {
		return err
	}
	for _, scheduleWithTpl := range schedules {
		schedule := scheduleWithTpl.Schedule
		schedule.ID = 0
		schedule.ProjectID = newProjectID
		if newTplID, ok := m.templateIDs[schedule.TemplateID]; ok {
			schedule.TemplateID = newTplID
		}
		if schedule.RepositoryID != nil {
			if newRepoID, ok := m.repositoryIDs[*schedule.RepositoryID]; ok {
				schedule.RepositoryID = &newRepoID
			}
		}

		_, err = m.newStore.CreateSchedule(schedule)
		if err != nil {
			return err
		}
	}
	return nil
}

func Migrate(oldStore, newStore db.Store) error {
	migrator := &Migrator{}
	return migrator.Migrate(oldStore, newStore)
}
