package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gorilla/securecookie"
)

// Cookie is a runtime generated secure cookie used for authentication
var Cookie *securecookie.SecureCookie

// WebHostURL is the public route to the semaphore server
var WebHostURL *url.URL

type DbDriver string

const (
	DbDriverMySQL    DbDriver = "mysql"
	DbDriverBolt     DbDriver = "bolt"
	DbDriverPostgres DbDriver = "postgres"
)

type DbConfig struct {
	Dialect DbDriver `json:"-"`

	Hostname string            `json:"host"`
	Username string            `json:"user"`
	Password string            `json:"pass"`
	DbName   string            `json:"name"`
	Options  map[string]string `json:"options"`
}

type ldapMappings struct {
	DN   string `json:"dn"`
	Mail string `json:"mail"`
	UID  string `json:"uid"`
	CN   string `json:"cn"`
}

type VariablesPassingMethod string

const (
	VariablesPassingNone  VariablesPassingMethod = "none"
	VariablesPassingEnv   VariablesPassingMethod = "env_vars"
	VariablesPassingExtra VariablesPassingMethod = "extra_vars"
	VariablesPassingBoth  VariablesPassingMethod = ""
)

//ConfigType mapping between Config and the json file that sets it
type ConfigType struct {
	MySQL    DbConfig `json:"mysql"`
	BoltDb   DbConfig `json:"bolt"`
	Postgres DbConfig `json:"postgres"`

	Dialect DbDriver `json:"dialect"`

	// Format `:port_num` eg, :3000
	// if : is missing it will be corrected
	Port string `json:"port"`

	// Interface ip, put in front of the port.
	// defaults to empty
	Interface string `json:"interface"`

	// semaphore stores ephemeral projects here
	TmpPath string `json:"tmp_path"`

	// cookie hashing & encryption
	CookieHash          string `json:"cookie_hash"`
	CookieEncryption    string `json:"cookie_encryption"`
	AccessKeyEncryption string `json:"access_key_encryption"`

	// email alerting
	EmailSender   string `json:"email_sender"`
	EmailHost     string `json:"email_host"`
	EmailPort     string `json:"email_port"`
	EmailUsername string `json:"email_username"`
	EmailPassword string `json:"email_password"`

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
	EmailSecure   bool `json:"email_secure"`
	TelegramAlert bool `json:"telegram_alert"`
	LdapEnable    bool `json:"ldap_enable"`
	LdapNeedTLS   bool `json:"ldap_needtls"`

	SshConfigPath string `json:"ssh_config_path"`

	// VariablesPassingMethod defines how Semaphore will pass variables to Ansible.
	// Default both via environment variables and via extra vars.
	VariablesPassingMethod VariablesPassingMethod `json:"variables_passing_method"`

	// RegisterFirstUser allows register new user from web interface if no user exists in database.
	RegisterFirstUser bool `json:"register_first_user"`
}

//Config exposes the application configuration storage for use in the application
var Config *ConfigType

// ToJSON returns a JSON string of the config
func (conf *ConfigType) ToJSON() ([]byte, error) {
	return json.MarshalIndent(&conf, " ", "\t")
}

// ConfigInit reads in cli flags, and switches actions appropriately on them
func ConfigInit(configPath string) {
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

func loadConfig(configPath string) {
	if configPath == "" {
		configPath = os.Getenv("SEMAPHORE_CONFIG_PATH")
	}

	//If the configPath option has been set try to load and decode it
	//var usedPath string

	if configPath == "" {
		// if no configPath look in the cwd
		cwd, err := os.Getwd()
		exitOnConfigError(err)
		defaultPath := path.Join(cwd, "config.json")
		file, err := os.Open(defaultPath)
		exitOnConfigError(err)
		decodeConfig(file)
		//usedPath = defaultPath
	} else {
		path := configPath
		file, err := os.Open(path)
		exitOnConfigError(err)
		decodeConfig(file)
		//usedPath = path
	}
	//fmt.Println("Using config file: " + usedPath)
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
		fmt.Println("Cannot Find configuration! Use --config parameter to point to a JSON file generated by `semaphore setup`.")
		os.Exit(1)
	}
}

func decodeConfig(file io.Reader) {
	if err := json.NewDecoder(file).Decode(&Config); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
}

func mapToQueryString(m map[string]string) (str string) {
	for option, value := range m {
		if str != "" {
			str += "&"
		}
		str += option + "=" + value
	}
	if str != "" {
		str = "?" + str
	}
	return
}

// String returns dialect name for GORP.
func (d DbDriver) String() string {
	return string(d)
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
				"%s:%s@tcp(%s)/%s",
				d.Username,
				d.Password,
				d.Hostname,
				d.DbName)
		} else {
			connectionString = fmt.Sprintf(
				"%s:%s@tcp(%s)/",
				d.Username,
				d.Password,
				d.Hostname)
		}
		options := map[string]string{
			"parseTime":         "true",
			"interpolateParams": "true",
		}
		for v, k := range d.Options {
			options[v] = k
		}
		connectionString += mapToQueryString(options)
	case DbDriverPostgres:
		if includeDbName {
			connectionString = fmt.Sprintf(
				"postgres://%s:%s@%s/%s",
				d.Username,
				url.QueryEscape(d.Password),
				d.Hostname,
				d.DbName)
		} else {
			connectionString = fmt.Sprintf(
				"postgres://%s:%s@%s",
				d.Username,
				url.QueryEscape(d.Password),
				d.Hostname)
		}
		connectionString += mapToQueryString(d.Options)
	default:
		err = fmt.Errorf("unsupported database driver: %s", d.Dialect)
	}
	return
}

func (conf *ConfigType) GetDialect() (dialect DbDriver, err error) {
	if conf.Dialect == "" {
		switch {
		case conf.MySQL.IsPresent():
			dialect = DbDriverMySQL
		case conf.BoltDb.IsPresent():
			dialect = DbDriverBolt
		case conf.Postgres.IsPresent():
			dialect = DbDriverPostgres
		default:
			err = errors.New("database configuration not found")
		}
		return
	}

	dialect = conf.Dialect
	return
}

func (conf *ConfigType) GetDBConfig() (dbConfig DbConfig, err error) {
	var dialect DbDriver
	dialect, err = conf.GetDialect()

	if err != nil {
		return
	}

	switch dialect {
	case DbDriverBolt:
		dbConfig = conf.BoltDb
	case DbDriverPostgres:
		dbConfig = conf.Postgres
	case DbDriverMySQL:
		dbConfig = conf.MySQL
	default:
		err = errors.New("database configuration not found")
	}

	dbConfig.Dialect = dialect

	return
}

//GenerateSecrets generates cookie secret during setup
func (conf *ConfigType) GenerateSecrets() {
	hash := securecookie.GenerateRandomKey(32)
	encryption := securecookie.GenerateRandomKey(32)
	accessKeyEncryption := securecookie.GenerateRandomKey(32)

	conf.CookieHash = base64.StdEncoding.EncodeToString(hash)
	conf.CookieEncryption = base64.StdEncoding.EncodeToString(encryption)
	conf.AccessKeyEncryption = base64.StdEncoding.EncodeToString(accessKeyEncryption)
}
