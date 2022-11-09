package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.login, "login", "", "New user login")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.name, "name", "", "New user name")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.email, "email", "", "New user email")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.password, "password", "", "New user password")
	userAddCmd.PersistentFlags().BoolVar(&targetUserArgs.admin, "admin", false, "Mark new user as admin")
	userCmd.AddCommand(userAddCmd)
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new user",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true
		if targetUserArgs.name == "" {
			fmt.Println("Argument --name required")
			ok = false
		}
		if targetUserArgs.login == "" {
			fmt.Println("Argument --login required")
			ok = false
		}

		if targetUserArgs.email == "" {
			fmt.Println("Argument --email required")
			ok = false
		}

		if targetUserArgs.password == "" {
			fmt.Println("Argument --password required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore user add --help` for details.")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		if _, err := store.CreateUser(db.UserWithPwd{
			Pwd: targetUserArgs.password,
			User: db.User{
				Name:     targetUserArgs.name,
				Username: targetUserArgs.login,
				Email:    targetUserArgs.email,
				Admin:    targetUserArgs.admin,
			},
		}); err != nil {
			panic(err)
		}

		fmt.Printf("User %s <%s> added!\n", targetUserArgs.login, targetUserArgs.email)
	},
}
