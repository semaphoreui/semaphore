package setup

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

func InteractiveSetup(conf *util.ConfigType) {
	stdin := bufio.NewReader(os.Stdin)

	fmt.Print(interactiveSetupBlurb)

	dbPrompt := `What database to use:
   1 - MySQL
   2 - BoltDB
`

	var db int
	promptValue(stdin, dbPrompt, "1", &db)

	switch db {
	case 1:
		scanMySQL(conf, stdin)
	case 2:
		scanBoltDb(conf, stdin)
	}

	defaultPlaybookPath := filepath.Join(os.TempDir(), "semaphore")
	promptValue(stdin, "Playbook path", defaultPlaybookPath, &conf.TmpPath)
	conf.TmpPath = filepath.Clean(conf.TmpPath)

	promptValue(stdin, "Web root URL (optional, see https://github.com/ansible-semaphore/semaphore/wiki/Web-root-URL)", "", &conf.WebHost)

	promptConfirmation(stdin, "Enable email alerts?", false, &conf.EmailAlert)
	if conf.EmailAlert {
		promptValue(stdin, "Mail server host", "localhost", &conf.EmailHost)
		promptValue(stdin, "Mail server port", "25", &conf.EmailPort)
		promptValue(stdin, "Mail sender address", "semaphore@localhost", &conf.EmailSender)
	}

	promptConfirmation(stdin, "Enable telegram alerts?", false, &conf.TelegramAlert)
	if conf.TelegramAlert {
		promptValue(stdin, "Telegram bot token (you can get it from @BotFather)", "", &conf.TelegramToken)
		promptValue(stdin, "Telegram chat ID", "", &conf.TelegramChat)
	}

	promptConfirmation(stdin, "Enable LDAP authentication?", false, &conf.LdapEnable)
	if conf.LdapEnable {
		promptValue(stdin, "LDAP server host", "localhost:389", &conf.LdapServer)
		promptConfirmation(stdin, "Enable LDAP TLS connection", false, &conf.LdapNeedTLS)
		promptValue(stdin, "LDAP DN for bind", "cn=user,ou=users,dc=example", &conf.LdapBindDN)
		promptValue(stdin, "Password for LDAP bind user", "pa55w0rd", &conf.LdapBindPassword)
		promptValue(stdin, "LDAP DN for user search", "ou=users,dc=example", &conf.LdapSearchDN)
		promptValue(stdin, "LDAP search filter", `(uid=%s)`, &conf.LdapSearchFilter)
		promptValue(stdin, "LDAP mapping for DN field", "dn", &conf.LdapMappings.DN)
		promptValue(stdin, "LDAP mapping for username field", "uid", &conf.LdapMappings.UID)
		promptValue(stdin, "LDAP mapping for full name field", "cn", &conf.LdapMappings.CN)
		promptValue(stdin, "LDAP mapping for email field", "mail", &conf.LdapMappings.Mail)
	}
}

func scanBoltDb(conf *util.ConfigType, stdin *bufio.Reader) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		workingDirectory = os.TempDir()
	}
	defaultBoltDBPath := filepath.Join(workingDirectory, "database.boltdb")
	promptValue(stdin, "DB filename", defaultBoltDBPath, &conf.BoltDb.Hostname)
}

func scanMySQL(conf *util.ConfigType, stdin *bufio.Reader) {
	promptValue(stdin, "DB Hostname", "127.0.0.1:3306", &conf.MySQL.Hostname)
	promptValue(stdin, "DB User", "root", &conf.MySQL.Username)
	promptValue(stdin, "DB Password", "", &conf.MySQL.Password)
	promptValue(stdin, "DB Name", "semaphore", &conf.MySQL.DbName)
}

func ScanConfigPathAndSave(config *util.ConfigType) string {
	stdin := bufio.NewReader(os.Stdin)

	configDirectory, err := os.Getwd()
	if err != nil {
		configDirectory, err = os.UserConfigDir()
		if err != nil {
			// Final fallback
			configDirectory = "/etc/semaphore"
		}
		configDirectory = filepath.Join(configDirectory, "semaphore")
	}
	promptValue(stdin, "Config output directory", configDirectory, &configDirectory)

	fmt.Printf("Running: mkdir -p %v..\n", configDirectory)
	err = os.MkdirAll(configDirectory, 0755)
	if err != nil {
		log.Panic("Could not create config directory: " + err.Error())
	}

	// Marshal config to json
	bytes, err := config.ToJSON()
	if err != nil {
		panic(err)
	}

	configPath := filepath.Join(configDirectory, "config.json")
	if err = ioutil.WriteFile(configPath, bytes, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("Configuration written to %v..\n", configPath)
	return configPath
}

func VerifyConfig(config *util.ConfigType) bool {
	stdin := bufio.NewReader(os.Stdin)

	bytes, err := config.ToJSON()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nGenerated configuration:\n %v\n\n", string(bytes))

	var correct bool
	promptConfirmation(stdin, "Is this correct?", true, &correct)
	return correct
}

func promptValue(stdin *bufio.Reader, prompt string, def string, item interface{}) {
	// Print prompt with optional default value
	fmt.Print(prompt)
	if len(def) != 0 {
		fmt.Print(" (default " + def + ")")
	}
	fmt.Print("\n> ")

	str, err := stdin.ReadString('\n')
	if err != nil {
		log.WithFields(log.Fields{"level": "Warn"}).Warn(err.Error())
	}

	// Remove newlines
	str = strings.TrimSuffix(str, "\n")
	str = strings.TrimSuffix(str, "\r")

	// If default, print default on input line
	if len(str) == 0 {
		str = def
		fmt.Print("\033[1A")
		fmt.Println("> " + def)
	}

	//Parse
	if _, err := fmt.Sscanln(str, item); err != nil && err != io.EOF {
		log.WithFields(log.Fields{"level": "Warn"}).Warn(err.Error())
	}

	// Empty line after prompt
	fmt.Println("")
}

func promptConfirmation(stdin *bufio.Reader, prompt string, def bool, item *bool) {
	defString := "yes"
	if !def {
		defString = "no"
	}

	fmt.Print(prompt + " (yes/no) (default " + defString + ")")
	fmt.Print("\n> ")

	str, err := stdin.ReadString('\n')
	if err != nil {
		log.WithFields(log.Fields{"level": "Warn"}).Warn(err.Error())
	}

	switch strings.ToLower(str) {
	case "y", "yes":
		*item = true
	case "n", "no":
		*item = false
	default:
		*item = def
		fmt.Print("\033[1A")
		fmt.Println("> " + defString)
	}

	// Empty line after prompt
	fmt.Println("")
}
