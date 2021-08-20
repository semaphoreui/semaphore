package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"

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

type UserAddArgs struct {
	Username string
	Name     string
	Email    string
	Password string
}

var UserAdd *UserAddArgs

// WebHostURL is the public route to the semaphore server
var WebHostURL *url.URL

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
	MySQL  DbConfig `json:"mysql"`
	BoltDb DbConfig `json:"bolt"`

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

// ToJSON returns a JSON string of the config
func (config *ConfigType) ToJSON() ([]byte, error) {
	return json.MarshalIndent(&config, " ", "\t")
}

// ConfigInit reads in cli flags, and switches actions appropriately on them
func ConfigInit() {
	flag.BoolVar(&InteractiveSetup, "setup", false, "perform interactive setup")
	flag.BoolVar(&Migration, "migrate", false, "execute migrations")
	flag.BoolVar(&Upgrade, "upgrade", false, "upgrade semaphore")
	configPath := flag.String("config", "", "config path")

	var unhashedPwd string
	flag.StringVar(&unhashedPwd, "hash", "", "generate hash of given password")

	var printConfig bool
	flag.BoolVar(&printConfig, "printConfig", false, "print example configuration")

	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "print the semaphore version")

	var userAdd bool
	flag.BoolVar(&userAdd, "useradd", false, "add new user")

	var userAddArgs UserAddArgs

	flag.StringVar(&userAddArgs.Username, "login", "", "new user login")
	flag.StringVar(&userAddArgs.Password, "password", "", "new user password")
	flag.StringVar(&userAddArgs.Name, "name", "", "new user name")
	flag.StringVar(&userAddArgs.Email, "email", "", "new user email")

	flag.Parse()

	if userAdd {
		if userAddArgs.Username == "" || userAddArgs.Password == "" || userAddArgs.Name == "" || userAddArgs.Email == "" {
			fmt.Println("Required options:\n  -login\n  -name\n  -email\n  -password")
			os.Exit(0)
		}

		UserAdd = &userAddArgs
	}

	if InteractiveSetup {
		return
	}

	if printVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if printConfig {
		config := &ConfigType{
			MySQL: DbConfig{
				Hostname: "127.0.0.1:3306",
				Username: "root",
				DbName:   "semaphore",
			},
			Port:    ":3000",
			TmpPath: "/tmp/semaphore",
		}
		config.GenerateCookieSecrets()

		bytes, _ := config.ToJSON()
		fmt.Println(string(bytes))

		os.Exit(0)
	}

	if len(unhashedPwd) > 0 {
		password, _ := bcrypt.GenerateFromPassword([]byte(unhashedPwd), 11)
		fmt.Println("Generated password: ", string(password))

		os.Exit(0)
	}

	loadConfig(configPath)
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

func loadConfig(configPath *string) {
	//If the configPath option has been set try to load and decode it
	var usedPath string
	if configPath != nil && len(*configPath) > 0 {
		path := *configPath
		file, err := os.Open(path)
		exitOnConfigError(err)
		decodeConfig(file)
		usedPath = path
	} else {
		// if no configPath look in the cwd
		cwd, err := os.Getwd()
		exitOnConfigError(err)
		defaultPath := path.Join(cwd, "config.json")
		file, err := os.Open(defaultPath)
		exitOnConfigError(err)
		decodeConfig(file)
		usedPath = defaultPath
	}

	fmt.Println("Using config file: " + usedPath)
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
		fmt.Println("Cannot Find configuration! Use -config parameter to point to a JSON file generated by -setup.\n\n Hint: have you run `-setup` ?")
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

func (d *DbConfig) IsPresent() bool {
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
				"%s:%s@tcp(%s)/?parseTime=true&interpolateParams=true",
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
	case conf.MySQL.IsPresent():
		dbConfig = conf.MySQL
		dbConfig.Dialect = DbDriverMySQL
	case conf.BoltDb.IsPresent():
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
