package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/util"
)

func main() {
	// Initialize database connection
	store, err := db.CreateStore()
	if err != nil {
		log.Fatal("Failed to create store:", err)
	}
	defer store.Close("update compliance folders")

	// Get all projects
	projects, err := store.GetProjects(db.RetrieveQueryParams{})
	if err != nil {
		log.Fatal("Failed to get projects:", err)
	}

	for _, project := range projects {
		fmt.Printf("Processing project: %s (ID: %d)\n", project.Name, project.ID)

		// Get all templates for this project
		templates, err := store.GetTemplates(project.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
		if err != nil {
			fmt.Printf("Failed to get templates for project %d: %v\n", project.ID, err)
			continue
		}

		updated := 0
		for _, template := range templates {
			// Check if this is a compliance template that needs folder assignment
			if template.Folder == nil || *template.Folder == "" {
				folderName := ""
				
				// Check if this is a STIG template
				if strings.Contains(strings.ToLower(template.Name), "stig") {
					if strings.Contains(strings.ToLower(template.Name), "rhel") || strings.Contains(strings.ToLower(template.Name), "redhat") {
						folderName = "STIG RHEL 9"
					} else if strings.Contains(strings.ToLower(template.Name), "ubuntu") {
						folderName = "STIG Ubuntu 20.04"
					} else if strings.Contains(strings.ToLower(template.Name), "windows") {
						folderName = "STIG Windows Server 2019"
					}
				} else if strings.Contains(strings.ToLower(template.Name), "cis") {
					if strings.Contains(strings.ToLower(template.Name), "rhel") || strings.Contains(strings.ToLower(template.Name), "redhat") {
						folderName = "CIS RHEL 9"
					} else if strings.Contains(strings.ToLower(template.Name), "ubuntu") {
						folderName = "CIS Ubuntu 20.04"
					} else if strings.Contains(strings.ToLower(template.Name), "windows") {
						folderName = "CIS Windows Server 2019"
					}
				}

				if folderName != "" {
					template.Folder = &folderName
					err := store.UpdateTemplate(template)
					if err != nil {
						fmt.Printf("Failed to update template %d: %v\n", template.ID, err)
					} else {
						fmt.Printf("  Updated template '%s' to folder '%s'\n", template.Name, folderName)
						updated++
					}
				}
			}
		}

		fmt.Printf("  Updated %d templates\n", updated)
	}

	fmt.Println("Done!")
}
