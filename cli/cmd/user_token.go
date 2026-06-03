package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/spf13/cobra"
)

type tokenArgs struct {
	login string
	name  string
	ttl   string
}

var targetTokenArgs tokenArgs

func init() {
	tokenCreateCmd.PersistentFlags().StringVar(&targetTokenArgs.login, "login", "", "Login of the token owner")
	tokenCreateCmd.PersistentFlags().StringVar(&targetTokenArgs.ttl, "ttl", "", "Token lifetime (e.g. 1h, 30m, 24h). Token never expires if omitted")
	tokenCreateCmd.PersistentFlags().StringVar(&targetTokenArgs.name, "name", "", "Token name")

	tokenListCmd.PersistentFlags().StringVar(&targetTokenArgs.login, "login", "", "Login of the token owner")

	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCmd.AddCommand(tokenListCmd)
	userCmd.AddCommand(tokenCmd)
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage user API tokens",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

func getTokenUser(store db.Store) db.User {
	if targetTokenArgs.login == "" {
		fmt.Println("Argument --login required")
		os.Exit(1)
	}

	user, err := store.GetUserByLoginOrEmail(targetTokenArgs.login, "")
	if errors.Is(err, db.ErrNotFound) {
		fmt.Printf("User with login %s not found\n", targetTokenArgs.login)
		os.Exit(1)
	}
	if err != nil {
		panic(err)
	}

	return user
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new API token",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		defer store.Close("")

		user := getTokenUser(store)

		var expiresAt *time.Time
		if targetTokenArgs.ttl != "" {
			d, err := time.ParseDuration(targetTokenArgs.ttl)
			if err != nil {
				fmt.Printf("Invalid --ttl value: %s\n", err)
				os.Exit(1)
			}
			t := tz.Now().Add(d)
			expiresAt = &t
		}

		tokenID := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
			panic(err)
		}

		token, err := store.CreateAPIToken(db.APIToken{
			ID:        strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
			UserID:    user.ID,
			Expired:   false,
			ExpiresAt: expiresAt,
			Name:      targetTokenArgs.name,
		})
		if err != nil {
			panic(err)
		}

		fmt.Println(token.ID)
	},
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List user API tokens",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		defer store.Close("")

		user := getTokenUser(store)

		tokens, err := store.GetAPITokens(user.ID)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			panic(err)
		}

		now := tz.Now()
		for _, token := range tokens {
			status := "active"
			if token.IsExpiredAt(now) {
				status = "expired"
			}

			expires := "never"
			if token.ExpiresAt != nil {
				expires = token.ExpiresAt.Format(time.RFC3339)
			}

			fmt.Printf("%s\t%s\t%s\n", token.Name, status, expires)
		}
	},
}
