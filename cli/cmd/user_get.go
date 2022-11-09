package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	userGetCmd.PersistentFlags().StringVar(&targetUserArgs.login, "login", "", "Login of the user you want to see")
	userGetCmd.PersistentFlags().StringVar(&targetUserArgs.email, "email", "", "Email of the user you want to see")
	userCmd.AddCommand(userGetCmd)
}

var userGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show user's data",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.login == "" && targetUserArgs.email == "" {
			fmt.Println("Argument --email or --login required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore user get --help` for details.")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := store.GetUserByLoginOrEmail(targetUserArgs.login, targetUserArgs.email)
		if err != nil {
			panic(err)
		}

		fmt.Printf("ID: %d\n", user.ID)
		fmt.Printf("Created: %s\n", user.Created)
		fmt.Printf("Login: %s\n", user.Username)
		fmt.Printf("Name: %s\n", user.Name)
		fmt.Printf("Email: %s\n", user.Email)
		fmt.Printf("Admin: %t\n", user.Admin)
	},
}
