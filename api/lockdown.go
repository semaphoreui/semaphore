package api

import (
	"net/http"

	"github.com/Digital-Data-Co/forge/api/helpers"
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/services/lockdown"
	"github.com/Digital-Data-Co/forge/util"
	"github.com/gorilla/mux"
)

// LockdownController handles Ansible Lockdown integration endpoints
type LockdownController struct {
	importService *lockdown.ImportService
}

// NewLockdownController creates a new LockdownController
func NewLockdownController(store db.Store) *LockdownController {
	return &LockdownController{
		importService: lockdown.NewImportService(store),
	}
}

// GetSupportedOS returns supported operating systems for compliance
func (c *LockdownController) GetSupportedOS(w http.ResponseWriter, r *http.Request) {
	lockdownService := lockdown.NewLockdownService()
	osList := lockdownService.GetSupportedOS()
	
	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"operating_systems": osList,
	})
}

// GetSupportedFrameworks returns supported compliance frameworks
func (c *LockdownController) GetSupportedFrameworks(w http.ResponseWriter, r *http.Request) {
	lockdownService := lockdown.NewLockdownService()
	frameworks := lockdownService.GetSupportedFrameworks()
	
	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"frameworks": frameworks,
	})
}

// GetComplianceRoles returns available compliance roles for a specific OS and framework
func (c *LockdownController) GetComplianceRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	os := vars["os"]
	framework := vars["framework"]
	
	if os == "" || framework == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "OS and framework parameters are required",
		})
		return
	}
	
	lockdownService := lockdown.NewLockdownService()
	roles, err := lockdownService.GetComplianceRoles(r.Context(), os, framework)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch compliance roles: " + err.Error(),
		})
		return
	}
	
	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"roles": roles,
	})
}

// GetComplianceTemplates returns compliance templates for a project
func (c *LockdownController) GetComplianceTemplates(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	
	templates, err := c.importService.GetComplianceTemplates(project.ID)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch compliance templates: " + err.Error(),
		})
		return
	}
	
	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

// ImportComplianceTasks imports compliance tasks for a project
func (c *LockdownController) ImportComplianceTasks(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	
	var request struct {
		Framework string `json:"framework" binding:"required"`
		OS        string `json:"os" binding:"required"`
	}
	
	if !helpers.Bind(w, r, &request) {
		return
	}
	
	err := c.importService.ImportComplianceTasks(r.Context(), project.ID, request.Framework, request.OS)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to import compliance tasks: " + err.Error(),
		})
		return
	}
	
	helpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Compliance tasks imported successfully",
	})
}

// CreateComplianceProject creates a new project with compliance framework settings
func (c *LockdownController) CreateComplianceProject(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)
	
	if !user.Admin && !util.Config.NonAdminCanCreateProject {
		helpers.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "Not authorized to create projects",
		})
		return
	}
	
	var request struct {
		Project             db.Project `json:"project" binding:"required"`
		ComplianceFramework string     `json:"compliance_framework" binding:"required"`
		ComplianceOS        string     `json:"compliance_os" binding:"required"`
		EnableSTIG          bool       `json:"enable_stig"`
	}
	
	if !helpers.Bind(w, r, &request) {
		return
	}
	
	project, err := c.importService.CreateComplianceProject(
		r.Context(),
		request.Project,
		request.ComplianceFramework,
		request.ComplianceOS,
		request.EnableSTIG,
	)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to create compliance project: " + err.Error(),
		})
		return
	}
	
	// Create project user relationship
	_, err = helpers.Store(r).CreateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      db.ProjectOwner,
	})
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to create project user relationship: " + err.Error(),
		})
		return
	}
	
	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      user.ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventProject,
		ObjectID:    project.ID,
		Description: "Compliance project created with " + request.ComplianceFramework + " " + request.ComplianceOS,
	})
	
	helpers.WriteJSON(w, http.StatusCreated, project)
}
