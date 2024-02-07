package setup

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/ansible-semaphore/semaphore/util"
)

const interactiveSetupBlurb = `
Hello! You will now be guided through a setup to:

1. Set up configuration for a MySQL/MariaDB database
2. Set up a path for your playbooks (auto-created)
3. Run database Migrations
4. Set up initial semaphore user & password

`

func InteractiveSetup(conf *util.ConfigType, stdin *bufio.Reader) {
	fmt.Print(interactiveSetupBlurb)

	dbPrompt := `What database to use:
   1 - MySQL
   2 - BoltDB
   3 - PostgreSQL
`

	var db int
	askValue(dbPrompt, "1", &db, stdin)

	switch db {
	case 1:
		conf.Dialect = util.DbDriverMySQL
		scanMySQL(conf, stdin)
	case 2:
		conf.Dialect = util.DbDriverBolt
		scanBoltDb(conf, stdin)
	case 3:
		conf.Dialect = util.DbDriverPostgres
		scanPostgres(conf, stdin)
	}

	defaultPlaybookPath := filepath.Join(os.TempDir(), "semaphore")
	askValue("Playbook path", defaultPlaybookPath, &conf.TmpPath, stdin)
	conf.TmpPath = filepath.Clean(conf.TmpPath)

	askValue("Public URL (optional, example: https://example.com/semaphore)", "", &conf.WebHost, stdin)

	askConfirmation("Enable email alerts?", false, &conf.EmailAlert, stdin)
	if conf.EmailAlert {
		askValue("Mail server host", "localhost", &conf.EmailHost, stdin)
		askValue("Mail server port", "25", &conf.EmailPort, stdin)
		askValue("Mail sender address", "semaphore@localhost", &conf.EmailSender, stdin)
	}

	askConfirmation("Enable telegram alerts?", false, &conf.TelegramAlert, stdin)
	if conf.TelegramAlert {
		askValue("Telegram bot token (you can get it from @BotFather)", "", &conf.TelegramToken, stdin)
		askValue("Telegram chat ID", "", &conf.TelegramChat, stdin)
	}

	askConfirmation("Enable slack alerts?", false, &conf.SlackAlert, stdin)
	if conf.SlackAlert {
		askValue("Slack Webhook URL", "", &conf.SlackUrl, stdin)
	}

	askConfirmation("Enable LDAP authentication?", false, &conf.LdapEnable, stdin)
	if conf.LdapEnable {
		askValue("LDAP server host", "localhost:389", &conf.LdapServer, stdin)
		askConfirmation("Enable LDAP TLS connection", false, &conf.LdapNeedTLS, stdin)
		askValue("LDAP DN for bind", "cn=user,ou=users,dc=example", &conf.LdapBindDN, stdin)
		askValue("Password for LDAP bind user", "pa55w0rd", &conf.LdapBindPassword, stdin)
		askValue("LDAP DN for user search", "ou=users,dc=example", &conf.LdapSearchDN, stdin)
		askValue("LDAP search filter", `(uid=%s)`, &conf.LdapSearchFilter, stdin)
		askValue("LDAP mapping for DN field", "dn", &conf.LdapMappings.DN, stdin)
		askValue("LDAP mapping for username field", "uid", &conf.LdapMappings.UID, stdin)
		askValue("LDAP mapping for full name field", "cn", &conf.LdapMappings.CN, stdin)
		askValue("LDAP mapping for email field", "mail", &conf.LdapMappings.Mail, stdin)
	}
}

func scanBoltDb(conf *util.ConfigType, stdin *bufio.Reader) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		workingDirectory = os.TempDir()
	}
	defaultBoltDBPath := filepath.Join(workingDirectory, "database.boltdb")
	askValue("db filename", defaultBoltDBPath, &conf.BoltDb.Hostname, stdin)
}

func scanMySQL(conf *util.ConfigType, stdin *bufio.Reader) {
	askValue("db Hostname", "127.0.0.1:3306", &conf.MySQL.Hostname, stdin)
	askValue("db User", "root", &conf.MySQL.Username, stdin)
	askValue("db Password", "", &conf.MySQL.Password, stdin)
	askValue("db Name", "semaphore", &conf.MySQL.DbName, stdin)
}

func scanPostgres(conf *util.ConfigType, stdin *bufio.Reader) {
	askValue("db Hostname", "127.0.0.1:5432", &conf.Postgres.Hostname, stdin)
	askValue("db User", "root", &conf.Postgres.Username, stdin)
	askValue("db Password", "", &conf.Postgres.Password, stdin)
	askValue("db Name", "semaphore", &conf.Postgres.DbName, stdin)
	if conf.Postgres.Options == nil {
		conf.Postgres.Options = make(map[string]string)
	}
	if _, exists := conf.Postgres.Options["sslmode"]; !exists {
		conf.Postgres.Options["sslmode"] = "disable"
	}
}

func SaveConfig(config *util.ConfigType, stdin *bufio.Reader) (configPath string) {
	configDirectory, err := os.Getwd()
	if err != nil {
		configDirectory, err = os.UserConfigDir()
		if err != nil {
			// Final fallback
			configDirectory = "/etc/semaphore"
		}
		configDirectory = filepath.Join(configDirectory, "semaphore")
	}
	askValue("Config output directory", configDirectory, &configDirectory, stdin)

	fmt.Printf("Running: mkdir -p %v..\n", configDirectory)

	if _, err = os.Stat(configDirectory); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(configDirectory, 0755)
		}
	}

	if err != nil {
		log.Panic("Could not create config directory: " + err.Error())
	}

	// Marshal config to json
	bytes, err := config.ToJSON()
	if err != nil {
		panic(err)
	}

	configPath = filepath.Join(configDirectory, "config.json")
	if err = ioutil.WriteFile(configPath, bytes, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("Configuration written to %v..\n", configPath)
	return
}

func AskConfigConfirmation(config *util.ConfigType, stdin *bufio.Reader) bool {
	bytes, err := config.ToJSON()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nGenerated configuration:\n %v\n\n", string(bytes))

	var correct bool
	askConfirmation("Is this correct?", true, &correct, stdin)
	return correct
}

func askValue(prompt string, defaultValue string, item interface{}, stdin *bufio.Reader) {
	// Print prompt with optional default value
	fmt.Print(prompt)
	if len(defaultValue) != 0 {
		fmt.Print(" (default " + defaultValue + ")")
	}
	fmt.Print(": ")

	value, err := stdin.ReadString('\n')
	if err != nil {
		fmt.Println("An input error occurred: ", err, ", value: '", value, "'")
		return
	}

	// If nothing is entered, use the default value.
	// Otherwise, use the entered value.
	value = strings.TrimSpace(value)
	if value == "" {
		value = defaultValue
	}

	// This code block dynamically assigns a new value to the variable pointed to by 'item', based on the variable's type.
	// It first checks if 'item' is a non-nil pointer. If it is, the block then checks the type of the value that 'item' points to.
	// - If the pointed-to value is a string, it sets the value to the 'value' parameter directly as a string.
	// - If the pointed-to value is an integer, it attempts to convert the 'value' parameter from a string to an integer before setting it.
	// This dynamic assignment is made possible using the reflect package, allowing for type introspection and modification at runtime.
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		if v.Elem().Kind() == reflect.String {
			v.Elem().SetString(value)
		} else if v.Elem().Kind() == reflect.Int {
			num, err := strconv.Atoi(value)
			if err == nil {
				v.Elem().SetInt(int64(num))
			}
		}
	}

	// Empty line after prompt
	fmt.Println("")
}

func askConfirmation(prompt string, defaultValue bool, item *bool, stdin *bufio.Reader) {
	defString := "yes"
	if !defaultValue {
		defString = "no"
	}

	fmt.Print(prompt + " (yes/no) (default " + defString + "): ")

	answer, err := stdin.ReadString('\n')
	if err != nil {
		fmt.Println("An input error occurred: ", err, ", answer: '", answer, "'")
		return
	}

	answer = strings.TrimSpace(answer)

	switch strings.ToLower(answer) {
	case "y", "yes":
		*item = true
	case "n", "no":
		*item = false
	default:
		*item = defaultValue
	}

	// Empty line after prompt
	fmt.Println("")
}
