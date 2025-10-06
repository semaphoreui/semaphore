package lockdown

import (
	"context"
	"fmt"
	"strings"

	"github.com/Digital-Data-Co/forge/db"
	log "github.com/sirupsen/logrus"
)

// ImportService handles importing Ansible tasks from Lockdown repositories
type ImportService struct {
	lockdownService *LockdownService
	store           db.Store
}

// NewImportService creates a new ImportService instance
func NewImportService(store db.Store) *ImportService {
	return &ImportService{
		lockdownService: NewLockdownService(),
		store:           store,
	}
}

// ImportComplianceTasks imports compliance tasks from Ansible Lockdown repositories
func (s *ImportService) ImportComplianceTasks(ctx context.Context, projectID int, complianceFramework, complianceOS string) error {
	log.Infof("Importing compliance tasks for project %d: %s %s", projectID, complianceFramework, complianceOS)

	// Get available roles for the specified OS and framework
	roles, err := s.lockdownService.GetComplianceRoles(ctx, complianceOS, complianceFramework)
	if err != nil {
		return fmt.Errorf("failed to get compliance roles: %w", err)
	}

	if len(roles) == 0 {
		return fmt.Errorf("no compliance roles found for %s %s", complianceFramework, complianceOS)
	}

	// Get or create default environment
	environments, err := s.store.GetEnvironments(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return fmt.Errorf("failed to get environments: %w", err)
	}

	var defaultEnvID *int
	if len(environments) > 0 {
		defaultEnvID = &environments[0].ID
	}

	// Get or create default inventory
	inventories, err := s.store.GetInventories(projectID, db.RetrieveQueryParams{}, []db.InventoryType{})
	if err != nil {
		return fmt.Errorf("failed to get inventories: %w", err)
	}

	var defaultInventoryID *int
	if len(inventories) > 0 {
		defaultInventoryID = &inventories[0].ID
	} else {
		// Create a default inventory for compliance tasks
		defaultInventory := db.Inventory{
			Name:      "Default Compliance Inventory",
			ProjectID: projectID,
			Inventory: "localhost ansible_connection=local",
			Type:      db.InventoryStatic,
		}
		createdInventory, err := s.store.CreateInventory(defaultInventory)
		if err != nil {
			return fmt.Errorf("failed to create default inventory: %w", err)
		}
		defaultInventoryID = &createdInventory.ID
	}

	// Get or create default repository
	repositories, err := s.store.GetRepositories(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return fmt.Errorf("failed to get repositories: %w", err)
	}

	var defaultRepoID *int
	if len(repositories) > 0 {
		defaultRepoID = &repositories[0].ID
	} else {
		// Create a default compliance repository
		repo, err := s.store.CreateRepository(db.Repository{
			Name:      fmt.Sprintf("%s %s Compliance", complianceFramework, complianceOS),
			ProjectID: projectID,
			GitURL:    fmt.Sprintf("https://github.com/ansible-lockdown/%s-%s", complianceOS, complianceFramework),
			GitBranch: "master",
			SSHKeyID:  1, // Use default SSH key
		})
		if err != nil {
			return fmt.Errorf("failed to create default repository: %w", err)
		}
		defaultRepoID = &repo.ID
	}

	// Import tasks for each role
	for _, role := range roles {
		if err := s.importRoleTasks(ctx, projectID, role, defaultEnvID, defaultRepoID, defaultInventoryID, complianceFramework, complianceOS); err != nil {
			log.Errorf("Failed to import tasks for role %s: %v", role.Name, err)
			continue
		}
	}

	log.Infof("Successfully imported compliance tasks for project %d", projectID)
	return nil
}

// importRoleTasks imports tasks for a specific compliance role
func (s *ImportService) importRoleTasks(ctx context.Context, projectID int, role LockdownRole, envID, repoID, inventoryID *int, complianceFramework, complianceOS string) error {
	log.Infof("Importing tasks for role: %s", role.Name)

	// Get tasks from the role
	tasks, err := s.lockdownService.GetRoleTasks(ctx, role.Repository, "tasks/main.yml")
	if err != nil {
		return fmt.Errorf("failed to get role tasks: %w", err)
	}

	// Create folder name based on framework and OS (matching logging format)
	folderName := fmt.Sprintf("%s %s", complianceFramework, complianceOS)

	// Get or create Compliance view
	complianceViewID, err := s.getOrCreateComplianceView(projectID)
	if err != nil {
		return fmt.Errorf("failed to get or create compliance view: %w", err)
	}

	// Create templates for each task
	for _, task := range tasks {
		template := db.Template{
			Name:          fmt.Sprintf("%s - %s", role.Name, task.Name),
			Folder:        &folderName,
			Type:          db.TemplateTask,
			Playbook:      s.generatePlaybook(task),
			ProjectID:     projectID,
			EnvironmentID: envID,
			InventoryID:   inventoryID,
			RepositoryID:  *repoID,
			App:           db.AppAnsible,
			Description:   &task.Description,
			ViewID:        &complianceViewID,
		}

		_, err := s.store.CreateTemplate(template)
		if err != nil {
			log.Errorf("Failed to create template for task %s: %v", task.Name, err)
			continue
		}

		log.Infof("Created template: %s", template.Name)
	}

	return nil
}

// generatePlaybook generates an Ansible playbook for a compliance task
func (s *ImportService) generatePlaybook(task LockdownTask) string {
	var playbook strings.Builder

	playbook.WriteString("---\n")
	playbook.WriteString("- name: " + task.Description + "\n")
	playbook.WriteString("  hosts: all\n")
	playbook.WriteString("  become: yes\n")
	playbook.WriteString("  tasks:\n")
	playbook.WriteString("    - name: " + task.Name + "\n")

	// Add module and arguments
	playbook.WriteString("      " + task.Module + ":\n")
	for key, value := range task.Args {
		playbook.WriteString(fmt.Sprintf("        %s: %v\n", key, value))
	}

	// Add tags if present
	if len(task.Tags) > 0 {
		playbook.WriteString("      tags:\n")
		for _, tag := range task.Tags {
			playbook.WriteString(fmt.Sprintf("        - %s\n", tag))
		}
	}

	// Add when condition if present
	if task.When != "" {
		playbook.WriteString(fmt.Sprintf("      when: %s\n", task.When))
	}

	return playbook.String()
}

// CreateComplianceProject creates a new project with compliance framework settings
func (s *ImportService) CreateComplianceProject(ctx context.Context, project db.Project, complianceFramework, complianceOS string, enableSTIG bool) (*db.Project, error) {
	// Set compliance fields
	project.ComplianceFramework = &complianceFramework
	project.ComplianceOS = &complianceOS
	project.EnableSTIG = enableSTIG

	// Create the project
	createdProject, err := s.store.CreateProject(project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Import compliance tasks if STIG is enabled
	if enableSTIG {
		if err := s.ImportComplianceTasks(ctx, createdProject.ID, complianceFramework, complianceOS); err != nil {
			log.Errorf("Failed to import compliance tasks: %v", err)
			// Don't fail project creation if task import fails
		}
	}

	return &createdProject, nil
}

// GetComplianceTemplates returns all compliance-related templates for a project
func (s *ImportService) GetComplianceTemplates(projectID int) ([]db.Template, error) {
	templates, err := s.store.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	var complianceTemplates []db.Template
	for _, template := range templates {
		// Check if template name contains compliance keywords
		name := strings.ToUpper(template.Name)
		if strings.Contains(name, "CIS") || strings.Contains(name, "STIG") ||
			strings.Contains(name, "COMPLIANCE") || strings.Contains(name, "SECURITY") {
			complianceTemplates = append(complianceTemplates, template)
		}
	}

	return complianceTemplates, nil
}

// getOrCreateComplianceView gets an existing Compliance view or creates a new one
func (s *ImportService) getOrCreateComplianceView(projectID int) (int, error) {
	// First, try to get existing views
	views, err := s.store.GetViews(projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get views: %w", err)
	}

	// Look for existing Compliance view
	for _, view := range views {
		if view.Title == "Compliance" {
			return view.ID, nil
		}
	}

	// If no Compliance view exists, create one
	// Find the highest position to place it after existing views
	maxPosition := -1
	for _, view := range views {
		if view.Position > maxPosition {
			maxPosition = view.Position
		}
	}

	complianceView, err := s.store.CreateView(db.View{
		ProjectID: projectID,
		Title:     "Compliance",
		Position:  maxPosition + 1,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create compliance view: %w", err)
	}

	return complianceView.ID, nil
}
