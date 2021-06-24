package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"path"
	"path/filepath"

	"net/url"

	"io"
	"strings"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

// Cookie is a runtime generated secure cookie used for authentication
var Cookie *securecookie.SecureCookie

// Migration indicates that the user wishes to run database migrations, deprecated
var Migration bool

// InteractiveSetup indicates that the cli should perform interactive setup mode
var InteractiveSetup bool

// Upgrade indicates that we should perform an upgrade action
var Upgrade bool

// WebHostURL is the public route to the semaphore server
var WebHostURL *url.URL

const (
	longPos  = "yes"
	shortPos = "y"
)

type DbDriver int

const (
	DbDriverMySQL DbDriver = iota
	DbDriverBolt
)

type DbConfig struct {
	Dialect  DbDriver `json:"-"`
	Hostname string   `json:"host"`
	Username string   `json:"user"`
	Password string   `json:"pass"`
	DbName   string   `json:"name"`
}

type ldapMappings struct {
	DN   string `json:"dn"`
	Mail string `json:"mail"`
	UID  string `json:"uid"`
	CN   string `json:"cn"`
}

//ConfigType mapping between Config and the json file that sets it
type ConfigType struct {
	MySQL      DbConfig `json:"mysql"`
	BoltDb	   DbConfig `json:"bolt"`

	// Format `:port_num` eg, :3000
	// if : is missing it will be corrected
	Port string `json:"port"`

	// Interface ip, put in front of the port.
	// defaults to empty
	Interface string `json:"interface"`

	// semaphore stores ephemeral projects here
	TmpPath string `json:"tmp_path"`

	// cookie hashing & encryption
	CookieHash       string `json:"cookie_hash"`
	CookieEncryption string `json:"cookie_encryption"`

	// email alerting
	EmailSender string `json:"email_sender"`
	EmailHost   string `json:"email_host"`
	EmailPort   string `json:"email_port"`

	// web host
	WebHost string `json:"web_host"`

	// ldap settings
	LdapBindDN       string       `json:"ldap_binddn"`
	LdapBindPassword string       `json:"ldap_bindpassword"`
	LdapServer       string       `json:"ldap_server"`
	LdapSearchDN     string       `json:"ldap_searchdn"`
	LdapSearchFilter string       `json:"ldap_searchfilter"`
	LdapMappings     ldapMappings `json:"ldap_mappings"`

	// telegram alerting
	TelegramChat  string `json:"telegram_chat"`
	TelegramToken string `json:"telegram_token"`

	// task concurrency
	ConcurrencyMode  string `json:"concurrency_mode"`
	MaxParallelTasks int    `json:"max_parallel_tasks"`

	// configType field ordering with bools at end reduces struct size
	// (maligned check)

	// feature switches
	EmailAlert    bool `json:"email_alert"`
	TelegramAlert bool `json:"telegram_alert"`
	LdapEnable    bool `json:"ldap_enable"`
	LdapNeedTLS   bool `json:"ldap_needtls"`
}

//Config exposes the application configuration storage for use in the application
var Config *ConfigType

var confPath *string

// NewConfig returns a reference to a new blank configType
// nolint: golint
func NewConfig() *ConfigType {
	return &ConfigType{}
}

// ScanErrorChecker deals with errors encountered while scanning lines
// since we do not fail on these errors currently we can simply note them
// and move on
func ScanErrorChecker(n int, err error) {
	if err != nil {
		log.Warn("An input error occurred:" + err.Error())
	}
}

// ConfigInit reads in cli flags, and switches actions appropriately on them
func ConfigInit() {
	flag.BoolVar(&InteractiveSetup, "setup", false, "perform interactive setup")
	flag.BoolVar(&Migration, "migrate", false, "execute migrations")
	flag.BoolVar(&Upgrade, "upgrade", false, "upgrade semaphore")
	confPath = flag.String("config", "", "config path")

	var unhashedPwd string
	flag.StringVar(&unhashedPwd, "hash", "", "generate hash of given password")

	var printConfig bool
	flag.BoolVar(&printConfig, "printConfig", false, "print example configuration")

	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "print the semaphore version")

	flag.Parse()

	if InteractiveSetup {
		return
	}

	if printVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if printConfig {
		cfg := &ConfigType{
			MySQL: DbConfig{
				Hostname: "127.0.0.1:3306",
				Username: "root",
				DbName:   "semaphore",
			},
			Port:    ":3000",
			TmpPath: "/tmp/semaphore",
		}
		cfg.GenerateCookieSecrets()

		b, _ := json.MarshalIndent(cfg, "", "\t")
		fmt.Println(string(b))

		os.Exit(0)
	}

	if len(unhashedPwd) > 0 {
		password, _ := bcrypt.GenerateFromPassword([]byte(unhashedPwd), 11)
		fmt.Println("Generated password: ", string(password))

		os.Exit(0)
	}

	loadConfig()
	validateConfig()

	var encryption []byte

	hash, _ := base64.StdEncoding.DecodeString(Config.CookieHash)
	if len(Config.CookieEncryption) > 0 {
		encryption, _ = base64.StdEncoding.DecodeString(Config.CookieEncryption)
	}

	Cookie = securecookie.New(hash, encryption)
	WebHostURL, _ = url.Parse(Config.WebHost)
	if len(WebHostURL.String()) == 0 {
		WebHostURL = nil
	}
}

func loadConfig() {

	//If the confPath option has been set try to load and decode it
	if confPath != nil && len(*confPath) > 0 {
		file, err := os.Open(*confPath)
		exitOnConfigError(err)
		decodeConfig(file)
	} else {
		// if no confPath look in the cwd
		cwd, err := os.Getwd()
		exitOnConfigError(err)
		cwd = cwd + "/config.json"
		confPath = &cwd
		file, err := os.Open(*confPath)
		exitOnConfigError(err)
		decodeConfig(file)
	}
	fmt.Println("Using config file: " + *confPath)
}

func validateConfig() {

	validatePort()

	if len(Config.TmpPath) == 0 {
		Config.TmpPath = "/tmp/semaphore"
	}

	if Config.MaxParallelTasks < 1 {
		Config.MaxParallelTasks = 10
	}
}

func validatePort() {

	//TODO - why do we do this only with this variable?
	if len(os.Getenv("PORT")) > 0 {
		Config.Port = ":" + os.Getenv("PORT")
	}
	if len(Config.Port) == 0 {
		Config.Port = ":3000"
	}
	if !strings.HasPrefix(Config.Port, ":") {
		Config.Port = ":" + Config.Port
	}
}

func exitOnConfigError(err error) {
	if err != nil {
		fmt.Println("Cannot Find configuration! Use -c parameter to point to a JSON file generated by -setup.\n\n Hint: have you run `-setup` ?")
		os.Exit(1)
	}
}

func decodeConfig(file io.Reader) {
	if err := json.NewDecoder(file).Decode(&Config); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
}

func (d DbDriver) String() string {
	return [...]string{"mysql"}[d]
}

func (d *DbConfig) isPresent() bool {
	return d.Hostname != ""
}

func (d *DbConfig) HasSupportMultipleDatabases() bool {
	return true
}

func (d *DbConfig) GetConnectionString(includeDbName bool) (connectionString string, err error) {
	switch d.Dialect {
	case DbDriverBolt:
		connectionString = d.Hostname
	case DbDriverMySQL:
		if includeDbName {
			connectionString = fmt.Sprintf(
				"%s:%s@tcp(%s)/%s?parseTime=true&interpolateParams=true",
				d.Username,
				d.Password,
				d.Hostname,
				d.DbName)
		} else {
			connectionString = fmt.Sprintf(
				"%s:%s@tcp(%s)?parseTime=true&interpolateParams=true",
				d.Username,
				d.Password,
				d.Hostname)
		}
	default:
		err = fmt.Errorf("unsupported database driver: %s", d.Dialect)
	}
	return
}

func (conf *ConfigType) GetDBConfig() (dbConfig DbConfig, err error) {
	switch {
	case conf.MySQL.isPresent():
		dbConfig = conf.MySQL
		dbConfig.Dialect = DbDriverMySQL
	case conf.BoltDb.isPresent():
		dbConfig = conf.BoltDb
		dbConfig.Dialect = DbDriverBolt
	default:
		err = errors.New("database configuration not found")
	}
	return
}

//GenerateCookieSecrets generates cookie secret during setup
func (conf *ConfigType) GenerateCookieSecrets() {
	hash := securecookie.GenerateRandomKey(32)
	encryption := securecookie.GenerateRandomKey(32)

	conf.CookieHash = base64.StdEncoding.EncodeToString(hash)
	conf.CookieEncryption = base64.StdEncoding.EncodeToString(encryption)
}

func (conf *ConfigType) ScanBoltDb() {
	filename, err := os.Getwd() // os.UserHomeDir()
	exitOnConfigError(err)
	filename = filepath.Join(filename, "database.bolt")
	fmt.Print(" > DB filename (default " + filename + "): ")
	ScanErrorChecker(fmt.Scanln(&conf.BoltDb.Hostname))
	if len(conf.BoltDb.Hostname) == 0 {
		conf.BoltDb.Hostname = filename
	}
}

func (conf *ConfigType) ScanMySQL() {

	fmt.Print(" > DB Hostname (default 127.0.0.1:3306): ")
	ScanErrorChecker(fmt.Scanln(&conf.MySQL.Hostname))
	if len(conf.MySQL.Hostname) == 0 {
		conf.MySQL.Hostname = "127.0.0.1:3306"
	}

	fmt.Print(" > DB User (default root): ")
	ScanErrorChecker(fmt.Scanln(&conf.MySQL.Username))
	if len(conf.MySQL.Username) == 0 {
		conf.MySQL.Username = "root"
	}

	fmt.Print(" > DB Password: ")
	ScanErrorChecker(fmt.Scanln(&conf.MySQL.Password))

	fmt.Print(" > DB Name (default semaphore): ")
	ScanErrorChecker(fmt.Scanln(&conf.MySQL.DbName))
	if len(conf.MySQL.DbName) == 0 {
		conf.MySQL.DbName = "semaphore"
	}
}

//nolint: gocyclo
func (conf *ConfigType) Scan() {
	db := 1
	fmt.Println(" > DB")
	fmt.Println("   1 - MySQL")
	fmt.Println("   2 - bbolt")
	fmt.Print("   (default 1): ")
	ScanErrorChecker(fmt.Scanln(&db))

	switch db {
	case 1:
		conf.ScanMySQL()
	case 2:
		conf.ScanBoltDb()
	}

	fmt.Print(" > Playbook path (default /tmp/semaphore): ")
	ScanErrorChecker(fmt.Scanln(&conf.TmpPath))

	if len(conf.TmpPath) == 0 {
		conf.TmpPath = "/tmp/semaphore"
	}
	conf.TmpPath = path.Clean(conf.TmpPath)

	fmt.Print(" > Web root URL (optional, see https://github.com/ansible-semaphore/semaphore/wiki/Web-root-URL): ")
	ScanErrorChecker(fmt.Scanln(&conf.WebHost))

	var EmailAlertAnswer string
	fmt.Print(" > Enable email alerts (y/n, default n): ")
	ScanErrorChecker(fmt.Scanln(&EmailAlertAnswer))
	if EmailAlertAnswer == longPos || EmailAlertAnswer == shortPos {

		conf.EmailAlert = true

		fmt.Print(" > Mail server host (default localhost): ")
		ScanErrorChecker(fmt.Scanln(&conf.EmailHost))

		if len(conf.EmailHost) == 0 {
			conf.EmailHost = "localhost"
		}

		fmt.Print(" > Mail server port (default 25): ")
		ScanErrorChecker(fmt.Scanln(&conf.EmailPort))

		if len(conf.EmailPort) == 0 {
			conf.EmailPort = "25"
		}

		fmt.Print(" > Mail sender address (default semaphore@localhost): ")
		ScanErrorChecker(fmt.Scanln(&conf.EmailSender))

		if len(conf.EmailSender) == 0 {
			conf.EmailSender = "semaphore@localhost"
		}

	} else {
		conf.EmailAlert = false
	}

	var TelegramAlertAnswer string
	fmt.Print(" > Enable telegram alerts (y/n, default n): ")
	ScanErrorChecker(fmt.Scanln(&TelegramAlertAnswer))
	if TelegramAlertAnswer == longPos || TelegramAlertAnswer == shortPos {

		conf.TelegramAlert = true

		fmt.Print(" > Telegram bot token (you can get it from @BotFather) (default ''): ")
		ScanErrorChecker(fmt.Scanln(&conf.TelegramToken))

		if len(conf.TelegramToken) == 0 {
			conf.TelegramToken = ""
		}

		fmt.Print(" > Telegram chat ID (default ''): ")
		ScanErrorChecker(fmt.Scanln(&conf.TelegramChat))

		if len(conf.TelegramChat) == 0 {
			conf.TelegramChat = ""
		}

	} else {
		conf.TelegramAlert = false
	}

	var LdapAnswer string
	fmt.Print(" > Enable LDAP authentication (y/n, default n): ")
	ScanErrorChecker(fmt.Scanln(&LdapAnswer))
	if LdapAnswer == longPos || LdapAnswer == shortPos {

		conf.LdapEnable = true

		fmt.Print(" > LDAP server host (default localhost:389): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapServer))

		if len(conf.LdapServer) == 0 {
			conf.LdapServer = "localhost:389"
		}

		var LdapTLSAnswer string
		fmt.Print(" > Enable LDAP TLS connection (y/n, default n): ")
		ScanErrorChecker(fmt.Scanln(&LdapTLSAnswer))
		if LdapTLSAnswer == longPos || LdapTLSAnswer == shortPos {
			conf.LdapNeedTLS = true
		} else {
			conf.LdapNeedTLS = false
		}

		fmt.Print(" > LDAP DN for bind (default cn=user,ou=users,dc=example): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapBindDN))

		if len(conf.LdapBindDN) == 0 {
			conf.LdapBindDN = "cn=user,ou=users,dc=example"
		}

		fmt.Print(" > Password for LDAP bind user (default pa55w0rd): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapBindPassword))

		if len(conf.LdapBindPassword) == 0 {
			conf.LdapBindPassword = "pa55w0rd"
		}

		fmt.Print(" > LDAP DN for user search (default ou=users,dc=example): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapSearchDN))

		if len(conf.LdapSearchDN) == 0 {
			conf.LdapSearchDN = "ou=users,dc=example"
		}

		fmt.Print(" > LDAP search filter (default (uid=" + "%" + "s)): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapSearchFilter))

		if len(conf.LdapSearchFilter) == 0 {
			conf.LdapSearchFilter = "(uid=%s)"
		}

		fmt.Print(" > LDAP mapping for DN field (default dn): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapMappings.DN))

		if len(conf.LdapMappings.DN) == 0 {
			conf.LdapMappings.DN = "dn"
		}

		fmt.Print(" > LDAP mapping for username field (default uid): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapMappings.UID))

		if len(conf.LdapMappings.UID) == 0 {
			conf.LdapMappings.UID = "uid"
		}

		fmt.Print(" > LDAP mapping for full name field (default cn): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapMappings.CN))

		if len(conf.LdapMappings.CN) == 0 {
			conf.LdapMappings.CN = "cn"
		}

		fmt.Print(" > LDAP mapping for email field (default mail): ")
		ScanErrorChecker(fmt.Scanln(&conf.LdapMappings.Mail))

		if len(conf.LdapMappings.Mail) == 0 {
			conf.LdapMappings.Mail = "mail"
		}
	} else {
		conf.LdapEnable = false
	}
}
