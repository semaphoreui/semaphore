package setup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"

	"github.com/ansible-semaphore/semaphore/util"
)

const (
	yesLong  = "yes"
	yesShort = "y"
)

func InteractiveSetup(conf *util.ConfigType) {
	fmt.Print(`
 Hello! You will now be guided through a setup to:

 1. Set up configuration for a MySQL/MariaDB database
 2. Set up a path for your playbooks (auto-created)
 3. Run database Migrations
 4. Set up initial semaphore user & password

`)

	db := 1
	fmt.Println(" > DB")
	fmt.Println("   1 - MySQL")
	fmt.Println("   2 - bbolt")
	fmt.Print("   (default 1): ")
	scanErrorChecker(fmt.Scanln(&db))

	switch db {
	case 1:
		scanMySQL(conf)
	case 2:
		scanBoltDb(conf)
	}

	defaultPlaybookPath := path.Join(os.TempDir(), "semaphore")
	fmt.Print(" > Playbook path (default " + defaultPlaybookPath + "): ")
	scanErrorChecker(fmt.Scanln(&conf.TmpPath))

	if len(conf.TmpPath) == 0 {
		conf.TmpPath = defaultPlaybookPath
	}
	conf.TmpPath = path.Clean(conf.TmpPath)

	fmt.Print(" > Web root URL (optional, see https://github.com/ansible-semaphore/semaphore/wiki/Web-root-URL): ")
	scanErrorChecker(fmt.Scanln(&conf.WebHost))

	var EmailAlertAnswer string
	fmt.Print(" > Enable email alerts (y/n, default n): ")
	scanErrorChecker(fmt.Scanln(&EmailAlertAnswer))
	if EmailAlertAnswer == yesLong || EmailAlertAnswer == yesShort {

		conf.EmailAlert = true

		fmt.Print(" > Mail server host (default localhost): ")
		scanErrorChecker(fmt.Scanln(&conf.EmailHost))

		if len(conf.EmailHost) == 0 {
			conf.EmailHost = "localhost"
		}

		fmt.Print(" > Mail server port (default 25): ")
		scanErrorChecker(fmt.Scanln(&conf.EmailPort))

		if len(conf.EmailPort) == 0 {
			conf.EmailPort = "25"
		}

		fmt.Print(" > Mail sender address (default semaphore@localhost): ")
		scanErrorChecker(fmt.Scanln(&conf.EmailSender))

		if len(conf.EmailSender) == 0 {
			conf.EmailSender = "semaphore@localhost"
		}

	} else {
		conf.EmailAlert = false
	}

	var TelegramAlertAnswer string
	fmt.Print(" > Enable telegram alerts (y/n, default n): ")
	scanErrorChecker(fmt.Scanln(&TelegramAlertAnswer))
	if TelegramAlertAnswer == yesLong || TelegramAlertAnswer == yesShort {

		conf.TelegramAlert = true

		fmt.Print(" > Telegram bot token (you can get it from @BotFather) (default ''): ")
		scanErrorChecker(fmt.Scanln(&conf.TelegramToken))

		if len(conf.TelegramToken) == 0 {
			conf.TelegramToken = ""
		}

		fmt.Print(" > Telegram chat ID (default ''): ")
		scanErrorChecker(fmt.Scanln(&conf.TelegramChat))

		if len(conf.TelegramChat) == 0 {
			conf.TelegramChat = ""
		}

	} else {
		conf.TelegramAlert = false
	}

	var LdapAnswer string
	fmt.Print(" > Enable LDAP authentication (y/n, default n): ")
	scanErrorChecker(fmt.Scanln(&LdapAnswer))
	if LdapAnswer == yesLong || LdapAnswer == yesShort {

		conf.LdapEnable = true

		fmt.Print(" > LDAP server host (default localhost:389): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapServer))

		if len(conf.LdapServer) == 0 {
			conf.LdapServer = "localhost:389"
		}

		var LdapTLSAnswer string
		fmt.Print(" > Enable LDAP TLS connection (y/n, default n): ")
		scanErrorChecker(fmt.Scanln(&LdapTLSAnswer))
		if LdapTLSAnswer == yesLong || LdapTLSAnswer == yesShort {
			conf.LdapNeedTLS = true
		} else {
			conf.LdapNeedTLS = false
		}

		fmt.Print(" > LDAP DN for bind (default cn=user,ou=users,dc=example): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapBindDN))

		if len(conf.LdapBindDN) == 0 {
			conf.LdapBindDN = "cn=user,ou=users,dc=example"
		}

		fmt.Print(" > Password for LDAP bind user (default pa55w0rd): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapBindPassword))

		if len(conf.LdapBindPassword) == 0 {
			conf.LdapBindPassword = "pa55w0rd"
		}

		fmt.Print(" > LDAP DN for user search (default ou=users,dc=example): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapSearchDN))

		if len(conf.LdapSearchDN) == 0 {
			conf.LdapSearchDN = "ou=users,dc=example"
		}

		fmt.Print(" > LDAP search filter (default (uid=" + "%" + "s)): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapSearchFilter))

		if len(conf.LdapSearchFilter) == 0 {
			conf.LdapSearchFilter = "(uid=%s)"
		}

		fmt.Print(" > LDAP mapping for DN field (default dn): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapMappings.DN))

		if len(conf.LdapMappings.DN) == 0 {
			conf.LdapMappings.DN = "dn"
		}

		fmt.Print(" > LDAP mapping for username field (default uid): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapMappings.UID))

		if len(conf.LdapMappings.UID) == 0 {
			conf.LdapMappings.UID = "uid"
		}

		fmt.Print(" > LDAP mapping for full name field (default cn): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapMappings.CN))

		if len(conf.LdapMappings.CN) == 0 {
			conf.LdapMappings.CN = "cn"
		}

		fmt.Print(" > LDAP mapping for email field (default mail): ")
		scanErrorChecker(fmt.Scanln(&conf.LdapMappings.Mail))

		if len(conf.LdapMappings.Mail) == 0 {
			conf.LdapMappings.Mail = "mail"
		}
	} else {
		conf.LdapEnable = false
	}
}

func scanBoltDb(conf *util.ConfigType) {
	defaultBoltDBPath := path.Join(os.TempDir(), "boltdb")
	fmt.Print(" > DB filename (default " + defaultBoltDBPath + "): ")
	scanErrorChecker(fmt.Scanln(&conf.BoltDb.Hostname))
	if len(conf.BoltDb.Hostname) == 0 {
		conf.BoltDb.Hostname = defaultBoltDBPath
	}
}

func scanMySQL(conf *util.ConfigType) {
	fmt.Print(" > DB Hostname (default 127.0.0.1:3306): ")
	scanErrorChecker(fmt.Scanln(&conf.MySQL.Hostname))
	if len(conf.MySQL.Hostname) == 0 {
		conf.MySQL.Hostname = "127.0.0.1:3306"
	}

	fmt.Print(" > DB User (default root): ")
	scanErrorChecker(fmt.Scanln(&conf.MySQL.Username))
	if len(conf.MySQL.Username) == 0 {
		conf.MySQL.Username = "root"
	}

	fmt.Print(" > DB Password: ")
	scanErrorChecker(fmt.Scanln(&conf.MySQL.Password))

	fmt.Print(" > DB Name (default semaphore): ")
	scanErrorChecker(fmt.Scanln(&conf.MySQL.DbName))
	if len(conf.MySQL.DbName) == 0 {
		conf.MySQL.DbName = "semaphore"
	}
}

func ScanConfigPathAndSave(config *util.ConfigType) string {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		configDirectory = path.Join(configDirectory, "semaphore")
	} else {
		configDirectory = "/etc/semaphore"
	}
	fmt.Print(" > Config output directory (default " + configDirectory + "): ")

	var answer string
	scanErrorChecker(fmt.Scanln(&answer))
	if len(answer) > 0 {
		configDirectory = answer
	}

	fmt.Printf(" Running: mkdir -p %v..\n", configDirectory)
	err = os.MkdirAll(configDirectory, 0755)
	if err != nil {
		log.Panic("Could not create config directory: " + err.Error())
	}

	// Marshal config to json
	bytes, err := config.ToJSON()
	if err != nil {
		panic(err)
	}

	configPath := path.Join(configDirectory, "/config.json")
	if err = ioutil.WriteFile(configPath, bytes, 0644); err != nil {
		panic(err)
	}

	fmt.Printf(" Configuration written to %v..\n", configPath)
	return configPath
}

func VerifyConfig(config *util.ConfigType) bool {
	bytes, err := config.ToJSON()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n Generated configuration:\n %v\n\n", string(bytes))
	fmt.Print(" > Is this correct? (yes/no): ")

	var answer string
	scanErrorChecker(fmt.Scanln(&answer))
	return answer == yesLong || answer == yesShort
}

// scanErrorChecker deals with errors encountered while scanning lines
// since we do not fail on these errors currently we can simply note them
// and move on
func scanErrorChecker(n int, err error) {
	if err != nil {
		log.Warn("An input error occurred:" + err.Error())
	}
}
