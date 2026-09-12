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

func getTokenUser(store db.Store, login string) (db.User, error) {
	if login == "" {
		return db.User{}, errors.New("argument --login required")
	}

	user, err := store.GetUserByLoginOrEmail(login, "")
	if errors.Is(err, db.ErrNotFound) {
		return db.User{}, fmt.Errorf("user with login %s not found", login)
	}
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

func createUserToken(store db.Store, out io.Writer, args tokenArgs) error {
	user, err := getTokenUser(store, args.login)
	if err != nil {
		return err
	}

	var expiresAt *time.Time
	if args.ttl != "" {
		d, err := time.ParseDuration(args.ttl)
		if err != nil {
			return fmt.Errorf("invalid --ttl value: %w", err)
		}
		t := tz.Now().Add(d)
		expiresAt = &t
	}

	tokenID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
		return err
	}

	token, err := store.CreateAPIToken(db.APIToken{
		ID:        strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
		UserID:    user.ID,
		Expired:   false,
		ExpiresAt: expiresAt,
		Name:      args.name,
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(out, token.ID)
	return nil
}

func listUserTokens(store db.Store, out io.Writer, args tokenArgs) error {
	user, err := getTokenUser(store, args.login)
	if err != nil {
		return err
	}

	tokens, err := store.GetAPITokens(user.ID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return err
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

		fmt.Fprintf(out, "%s\t%s\t%s\n", token.Name, status, expires)
	}

	return nil
}

func runTokenCmd(fn func(db.Store, io.Writer, tokenArgs) error) {
	store := createStore("")
	defer store.Close()

	if err := fn(store, os.Stdout, targetTokenArgs); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new API token",
	Run: func(cmd *cobra.Command, args []string) {
		runTokenCmd(createUserToken)
	},
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List user API tokens",
	Run: func(cmd *cobra.Command, args []string) {
		runTokenCmd(listUserTokens)
	},
}
