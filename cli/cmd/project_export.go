package cmd

import (
	"fmt"
	"os"
	"strings"

	projectService "github.com/semaphoreui/semaphore/services/project"
	"github.com/spf13/cobra"
)

type projectExportArgs struct {
	projectID   int
	projectName string
	file        string
}

var targetProjectExportArgs projectExportArgs

func init() {
	projectExportCmd.PersistentFlags().IntVar(&targetProjectExportArgs.projectID, "project-id", 0, "Project ID to export")
	projectExportCmd.PersistentFlags().StringVar(&targetProjectExportArgs.projectName, "project-name", "", "Project name to export")
	projectExportCmd.PersistentFlags().StringVar(&targetProjectExportArgs.file, "file", "", "Output file path (default: stdout)")
	projectCmd.AddCommand(projectExportCmd)
}

var projectExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export project backup",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true
		if targetProjectExportArgs.projectID == 0 && targetProjectExportArgs.projectName == "" {
			fmt.Println("Argument --project-id or --project-name required")
			ok = false
		}

		if targetProjectExportArgs.projectID != 0 && targetProjectExportArgs.projectName != "" {
			fmt.Println("Only one of --project-id or --project-name can be specified")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore project export --help` for details.")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		projectID := targetProjectExportArgs.projectID

		if targetProjectExportArgs.projectName != "" {
			projects, err := store.GetAllProjects()
			if err != nil {
				fmt.Printf("Failed to get projects: %v\n", err)
				os.Exit(1)
			}

			found := false
			searchName := strings.ToLower(targetProjectExportArgs.projectName)
			for _, p := range projects {
				if strings.ToLower(p.Name) == searchName {
					projectID = p.ID
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("Project with name '%s' not found\n", targetProjectExportArgs.projectName)
				os.Exit(1)
			}
		}

		backup, err := projectService.GetBackup(projectID, store)
		if err != nil {
			fmt.Printf("Failed to create backup: %v\n", err)
			os.Exit(1)
		}

		data, err := backup.Marshal()
		if err != nil {
			fmt.Printf("Failed to marshal backup: %v\n", err)
			os.Exit(1)
		}

		if targetProjectExportArgs.file == "" {
			fmt.Println(data)
		} else {
			if err := os.WriteFile(targetProjectExportArgs.file, []byte(data), 0644); err != nil {
				fmt.Printf("Failed to write file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Project exported to %s\n", targetProjectExportArgs.file)
		}
	},
}
