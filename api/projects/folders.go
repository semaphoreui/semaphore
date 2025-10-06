package projects

import (
	"net/http"
	"sort"
	"strings"

	"github.com/Digital-Data-Co/forge/api/helpers"
	"github.com/Digital-Data-Co/forge/db"
)

// GetFolders returns all unique folders for a project
func GetFolders(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	
	// Get all templates for the project
	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	// Extract unique folders
	folderMap := make(map[string]bool)
	for _, template := range templates {
		if template.Folder != nil && *template.Folder != "" {
			folderMap[*template.Folder] = true
		}
	}
	
	// Convert to slice and sort
	var folders []string
	for folder := range folderMap {
		folders = append(folders, folder)
	}
	sort.Strings(folders)
	
	// Add "No Folder" option for templates without folders
	folders = append([]string{"No Folder"}, folders...)
	
	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"folders": folders,
	})
}

// GetTemplatesByFolder returns templates grouped by folder
func GetTemplatesByFolder(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	
	// Get all templates for the project
	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	// Group templates by folder
	folderMap := make(map[string][]db.Template)
	
	for _, template := range templates {
		folderName := "No Folder"
		if template.Folder != nil && *template.Folder != "" {
			folderName = *template.Folder
		}
		folderMap[folderName] = append(folderMap[folderName], template)
	}
	
	// Convert to slice of folder objects
	var folders []map[string]interface{}
	for folderName, templates := range folderMap {
		folders = append(folders, map[string]interface{}{
			"name":      folderName,
			"templates": templates,
			"count":     len(templates),
		})
	}
	
	// Sort folders (No Folder first, then alphabetically)
	sort.Slice(folders, func(i, j int) bool {
		nameI := folders[i]["name"].(string)
		nameJ := folders[j]["name"].(string)
		
		if nameI == "No Folder" {
			return true
		}
		if nameJ == "No Folder" {
			return false
		}
		return nameI < nameJ
	})
	
	helpers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"folders": folders,
	})
}

// CreateFolder creates a new folder (this is handled by updating templates)
func CreateFolder(w http.ResponseWriter, r *http.Request) {
	// Folders are created implicitly when templates are assigned to them
	// This endpoint is for future use if we want explicit folder management
	helpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Folders are created automatically when templates are assigned to them",
	})
}

// DeleteFolder deletes a folder by moving all templates to "No Folder"
func DeleteFolder(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	
	var request struct {
		FolderName string `json:"folder_name" binding:"required"`
	}
	
	if !helpers.Bind(w, r, &request) {
		return
	}
	
	if request.FolderName == "No Folder" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Cannot delete 'No Folder'",
		})
		return
	}
	
	// Get all templates in the folder
	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	// Move templates from the folder to "No Folder"
	for _, template := range templates {
		if template.Folder != nil && *template.Folder == request.FolderName {
			template.Folder = nil
			err := helpers.Store(r).UpdateTemplate(template)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}
	
	helpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Folder deleted successfully",
	})
}

// RenameFolder renames a folder by updating all templates in that folder
func RenameFolder(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	
	var request struct {
		OldName string `json:"old_name" binding:"required"`
		NewName string `json:"new_name" binding:"required"`
	}
	
	if !helpers.Bind(w, r, &request) {
		return
	}
	
	if request.OldName == "No Folder" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Cannot rename 'No Folder'",
		})
		return
	}
	
	if strings.TrimSpace(request.NewName) == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "New folder name cannot be empty",
		})
		return
	}
	
	// Get all templates in the old folder
	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{}, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	// Rename the folder for all templates
	for _, template := range templates {
		if template.Folder != nil && *template.Folder == request.OldName {
			template.Folder = &request.NewName
			err := helpers.Store(r).UpdateTemplate(template)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}
	
	helpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Folder renamed successfully",
	})
}
