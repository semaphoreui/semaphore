package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	projectService "github.com/semaphoreui/semaphore/services/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type projectImportArgs struct {
	dir  string
	file string
}

var targetProjectImportArgs projectImportArgs

func init() {
	projectImportCmd.PersistentFlags().StringVar(&targetProjectImportArgs.dir, "dir", "", "Directory path with project backups to import")
	projectImportCmd.PersistentFlags().StringVar(&targetProjectImportArgs.file, "file", "", "Backup file path to import")
	projectCmd.AddCommand(projectImportCmd)
}

var projectImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import project(s)",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true
		if targetProjectImportArgs.dir == "" && targetProjectImportArgs.file == "" {
			fmt.Println("Argument --dir or --file required")
			ok = false
		}

		if targetProjectImportArgs.dir != "" && targetProjectImportArgs.file != "" {
			fmt.Println("Only one of --dir or --file can be specified")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore project import --help` for details.")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := resolveImportUser(store)
		if err != nil {
			log.Errorf("cannot resolve user for import: %v", err)
			os.Exit(1)
		}

		files := make([]string, 0)
		if targetProjectImportArgs.file != "" {
			files = append(files, targetProjectImportArgs.file)
		}

		if targetProjectImportArgs.dir != "" {
			dir := targetProjectImportArgs.dir
			err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					return nil
				}
				// include likely backup files
				lower := strings.ToLower(d.Name())
				if strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".backup") || strings.HasSuffix(lower, ".bk") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return
			}
		}

		if len(files) == 0 {
			fmt.Println("No backup files found to import")
			os.Exit(1)
		}

		// sort for deterministic order
		sort.Strings(files)

		okCount := 0
		for _, f := range files {
			if err := importProjectFromFile(f, user, store); err != nil {
				log.Errorf("failed to import %s: %v", f, err)
				continue
			}
			fmt.Printf("Imported project from %s\n", f)
			okCount++
		}

		if okCount == 0 {
			os.Exit(1)
		}

		fmt.Printf("Project(s) imported: %d/%d\n", okCount, len(files))
	},
}

func resolveImportUser(store db.Store) (res db.User, err error) {
	admins, err := store.GetAllAdmins()
	if err != nil {
		return
	}

	if len(admins) > 0 {
		res = admins[0]
		return
	}
	users, err := store.GetUsers(db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	if len(users) == 0 {
		err = errors.New("no admins found in database; create a admin first")
		return
	}

	res = users[0]
	return
}

func importProjectFromFile(path string, user db.User, store db.Store) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var backup projectService.BackupFormat
	if err := backup.Unmarshal(string(data)); err != nil {
		return err
	}
	if err := backup.Verify(); err != nil {
		return err
	}
	_, err = backup.Restore(user, store)
	return err
}
