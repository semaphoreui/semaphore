package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	userDeleteCmd.PersistentFlags().StringVar(&targetUserArgs.login, "login", "", "Login of the user you want to delete")
	userDeleteCmd.PersistentFlags().StringVar(&targetUserArgs.email, "email", "", "Email of the user you want to delete")
	userCmd.AddCommand(userDeleteCmd)
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove existing user",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.login == "" && targetUserArgs.email == "" {
			fmt.Println("Argument --email or --login required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore user delete --help` for details.")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := store.GetUserByLoginOrEmail(targetUserArgs.login, targetUserArgs.email)
		if err != nil {
			panic(err)
		}

		if err := store.DeleteUser(user.ID); err != nil {
			panic(err)
		}

		fmt.Printf("User %s <%s> deleted!\n", user.Username, user.Email)
	},
}
