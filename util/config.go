package util

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/google/go-github/github"
	"github.com/gorilla/securecookie"
)

// Cookie is a runtime generated secure cookie used for authentication
var Cookie *securecookie.SecureCookie

// WebHostURL is the public route to the semaphore server
var WebHostURL *url.URL

const (
	DbDriverMySQL    = "mysql"
	DbDriverBolt     = "bolt" // Deprecated: replaced with sqlite
	DbDriverPostgres = "postgres"
	DbDriverSQLite   = "sqlite"
)

type EventLogAction string

const (
	EventLogCreate EventLogAction = "create"
	EventLogUpdate EventLogAction = "update"
	EventLogDelete EventLogAction = "delete"
)

type DbConfig struct {
	Dialect string `json:"-"`

	Hostname string            `json:"host,omitempty" env:"SEMAPHORE_DB_HOST" default:"0.0.0.0"`
	Username string            `json:"user,omitempty" env:"SEMAPHORE_DB_USER"`
	Password string            `json:"pass,omitempty" env:"SEMAPHORE_DB_PASS"`
	DbName   string            `json:"name,omitempty" env:"SEMAPHORE_DB" default:"semaphore"`
	Options  map[string]string `json:"options,omitempty" env:"SEMAPHORE_DB_OPTIONS"`
}

type LdapMappings struct {
	DN   string `json:"dn" env:"SEMAPHORE_LDAP_MAPPING_DN" default:"dn"`
	Mail string `json:"mail" env:"SEMAPHORE_LDAP_MAPPING_MAIL" default:"mail"`
	UID  string `json:"uid" env:"SEMAPHORE_LDAP_MAPPING_UID" default:"uid"`
	CN   string `json:"cn" env:"SEMAPHORE_LDAP_MAPPING_CN" default:"cn"`
}

func (p *LdapMappings) GetUsernameClaim() string {
	return p.UID
}

func (p *LdapMappings) GetEmailClaim() string {
	return p.Mail
}

func (p *LdapMappings) GetNameClaim() string {
	return p.CN
}

type oidcEndpoint struct {
	IssuerURL   string   `json:"issuer"`
	AuthURL     string   `json:"auth"`
	TokenURL    string   `json:"token"`
	UserInfoURL string   `json:"userinfo"`
	JWKSURL     string   `json:"jwks"`
	Algorithms  []string `json:"algorithms"`
}

const (
	// GoGitClientId is builtin Git client. It is not require external dependencies and is preferred.
	// Use it if you don't need external SSH authorization.
	GoGitClientId = "go_git"
	// CmdGitClientId is external Git client.
	// Default Git client. It is use external Git binary to clone repositories.
	CmdGitClientId = "cmd_git"
)

// // basic config validation using regex
// /* NOTE: other basic regex could be used:
//
//	ipv4: ^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$
//	ipv6: ^(?:[A-Fa-f0-9]{1,4}:|:){3,7}[A-Fa-f0-9]{1,4}$
//	domain: ^([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}$
//	path+filename: ^([\\/[a-zA-Z0-9_\\-${}:~]*]*\\/)?[a-zA-Z0-9\\.~_${}\\-:]*$
//	email address: ^(|.*@[A-Za-z0-9-\\.]*)$
//
// */

type RunnerConfig struct {
	RegistrationToken string `json:"-" env:"SEMAPHORE_RUNNER_REGISTRATION_TOKEN"`
	Token             string `json:"token,omitempty" env:"SEMAPHORE_RUNNER_TOKEN"`
	TokenFile         string `json:"token_file,omitempty" env:"SEMAPHORE_RUNNER_TOKEN_FILE"`
	PrivateKeyFile    string `json:"private_key_file,omitempty" env:"SEMAPHORE_RUNNER_PRIVATE_KEY_FILE"`

	// OneOff indicates than runner runs only one job and exit. It is very useful for dynamic runners.
	// How it works?
	// Example:
	// 1) User starts the task.
	// 2) Semaphore found runner for task and calls runner's webhook if it provided.
	// 3) Your server or lambda handling the call and starts the one-off runner.
	// 4) The runner connects to the Semaphore server and handles the enqueued task(s).
	OneOff bool `json:"one_off,omitempty" env:"SEMAPHORE_RUNNER_ONE_OFF"`

	Webhook string `json:"webhook,omitempty" env:"SEMAPHORE_RUNNER_WEBHOOK"`

	MaxParallelTasks int `json:"max_parallel_tasks,omitempty" default:"1" env:"SEMAPHORE_RUNNER_MAX_PARALLEL_TASKS"`
}

type TLSConfig struct {
	Enabled          bool   `json:"enabled" env:"SEMAPHORE_TLS_ENABLED"`
	CertFile         string `json:"cert_file" env:"SEMAPHORE_TLS_CERT_FILE"`
	KeyFile          string `json:"key_file" env:"SEMAPHORE_TLS_KEY_FILE"`
	HTTPRedirectPort *int   `json:"http_redirect_port,omitempty" env:"SEMAPHORE_TLS_HTTP_REDIRECT_PORT"`
}

type TotpConfig struct {
	Enabled       bool `json:"enabled" env:"SEMAPHORE_TOTP_ENABLED"`
	AllowRecovery bool `json:"allow_recovery" env:"SEMAPHORE_TOTP_ALLOW_RECOVERY"`
}

type EventLogType struct {
	Format  FileLogFormat      `json:"format,omitempty" env:"SEMAPHORE_EVENT_LOG_FORMAT"`
	Enabled bool               `json:"enabled" env:"SEMAPHORE_EVENT_LOG_ENABLED"`
	Logger  *lumberjack.Logger `json:"logger,omitempty" env:"SEMAPHORE_EVENT_LOGGER"`
}

type FileLogFormat string

const (
	FileLogJSON FileLogFormat = "json"
	FileLogRaw  FileLogFormat = ""
)

type TaskLogType struct {
	Enabled      bool               `json:"enabled" env:"SEMAPHORE_TASK_LOG_ENABLED"`
	Format       FileLogFormat      `json:"format,omitempty" env:"SEMAPHORE_TASK_LOG_FORMAT"`
	Logger       *lumberjack.Logger `json:"logger,omitempty" env:"SEMAPHORE_TASK_LOGGER"`
	ResultLogger *lumberjack.Logger `json:"result_logger,omitempty" env:"SEMAPHORE_TASK_RESULT_LOGGER"`
}

type ConfigLog struct {
	Events *EventLogType `json:"events,omitempty"`
	Tasks  *TaskLogType  `json:"tasks,omitempty"`
}

type ConfigProcess struct {
	User   string `json:"user,omitempty" env:"SEMAPHORE_PROCESS_USER"`
	UID    *int   `json:"uid,omitempty" env:"SEMAPHORE_PROCESS_UID"`
	Chroot string `json:"chroot,omitempty" env:"SEMAPHORE_PROCESS_CHROOT"`
	GID    *int   `json:"gid,omitempty" env:"SEMAPHORE_PROCESS_GID"`
}

type ScheduleConfig struct {
	Timezone string `json:"timezone,omitempty" env:"SEMAPHORE_SCHEDULE_TIMEZONE" default:"UTC"`
}

type DebuggingConfig struct {
	ApiDelay     string `json:"api_delay,omitempty" env:"SEMAPHORE_API_DELAY"`
	PprofDumpDir string `json:"pprof_dump_dir,omitempty" env:"SEMAPHORE_PPROF_DUMP_DIR"`
}

type HARedisConfig struct {
	Addr          string `json:"addr,omitempty" env:"SEMAPHORE_HA_REDIS_ADDR"`
	DB            int    `json:"db,omitempty" env:"SEMAPHORE_HA_REDIS_DB"`
	Pass          string `json:"pass,omitempty" env:"SEMAPHORE_HA_REDIS_PASS"`
	User          string `json:"user,omitempty" env:"SEMAPHORE_HA_REDIS_USER"`
	TLS           bool   `json:"tls,omitempty" env:"SEMAPHORE_HA_REDIS_TLS"`
	TLSSkipVerify bool   `json:"tls_skip_verify,omitempty" env:"SEMAPHORE_HA_REDIS_TLS_SKIP_VERIFY"`
}

type HAConfig struct {
	Enabled bool           `json:"enabled" env:"SEMAPHORE_HA_ENABLED"`
	Redis   *HARedisConfig `json:"redis,omitempty"`
}

type TeamInviteType string

const (
	TeamInviteEmail    TeamInviteType = "email"
	TeamInviteUsername TeamInviteType = "username"
	TeamInviteBoth     TeamInviteType = "both"
)

type TeamsConfig struct {
	InvitesEnabled  bool           `json:"invites_enabled,omitempty" env:"SEMAPHORE_TEAMS_INVITES_ENABLED"`
	InviteType      TeamInviteType `json:"invite_type,omitempty" env:"SEMAPHORE_TEAMS_INVITE_TYPE" default:"username"`
	MembersCanLeave bool           `json:"members_can_leave,omitempty" env:"SEMAPHORE_TEAMS_MEMBERS_CAN_LEAVE"`
}

// ConfigType mapping between Config and the json file that sets it
type ConfigType struct {
	MySQL    *DbConfig `json:"mysql,omitempty"`
	BoltDb   *DbConfig `json:"bolt,omitempty"` // Deprecated
	Postgres *DbConfig `json:"postgres,omitempty"`
	SQLite   *DbConfig `json:"sqlite,omitempty"`

	Dialect string `json:"dialect,omitempty" default:"bolt" rule:"^mysql|bolt|postgres|sqlite$" env:"SEMAPHORE_DB_DIALECT"`

	// Format `:port_num` eg, :3000
	// if : is missing it will be corrected
	Port string     `json:"port,omitempty" default:":3000" rule:"^:?([0-9]{1,5})$" env:"SEMAPHORE_PORT"`
	TLS  *TLSConfig `json:"tls,omitempty"`

	Auth *AuthConfig `json:"auth,omitempty"`

	// Interface ip, put in front of the port.
	// defaults to empty
	Interface string `json:"interface,omitempty" env:"SEMAPHORE_INTERFACE"`

	// semaphore stores ephemeral projects here
	TmpPath string `json:"tmp_path,omitempty" default:"/tmp/semaphore" env:"SEMAPHORE_TMP_PATH"`

	// SshConfigPath is a path to the custom SSH config file.
	// Default path is ~/.ssh/config.
	SshConfigPath string `json:"ssh_config_path,omitempty" env:"SEMAPHORE_SSH_PATH"`

	GitClientId string `json:"git_client,omitempty" rule:"^go_git|cmd_git$" env:"SEMAPHORE_GIT_CLIENT" default:"cmd_git"`

	// web host
	WebHost string `json:"web_host,omitempty" env:"SEMAPHORE_WEB_ROOT"`

	// cookie hashing & encryption
	CookieHash       string `json:"cookie_hash,omitempty" env:"SEMAPHORE_COOKIE_HASH"`
	CookieEncryption string `json:"cookie_encryption,omitempty" env:"SEMAPHORE_COOKIE_ENCRYPTION"`
	// AccessKeyEncryption is BASE64 encoded byte array used
	// for encrypting and decrypting access keys stored in database.
	AccessKeyEncryption string `json:"access_key_encryption,omitempty" env:"SEMAPHORE_ACCESS_KEY_ENCRYPTION"`

	// email alerting
	EmailAlert         bool   `json:"email_alert,omitempty" env:"SEMAPHORE_EMAIL_ALERT"`
	EmailSender        string `json:"email_sender,omitempty" env:"SEMAPHORE_EMAIL_SENDER"`
	EmailHost          string `json:"email_host,omitempty" env:"SEMAPHORE_EMAIL_HOST"`
	EmailPort          string `json:"email_port,omitempty" rule:"^(|[0-9]{1,5})$" env:"SEMAPHORE_EMAIL_PORT"`
	EmailUsername      string `json:"email_username,omitempty" env:"SEMAPHORE_EMAIL_USERNAME"`
	EmailPassword      string `json:"email_password,omitempty" env:"SEMAPHORE_EMAIL_PASSWORD"`
	EmailSecure        bool   `json:"email_secure,omitempty" env:"SEMAPHORE_EMAIL_SECURE"`
	EmailTls           bool   `json:"email_tls,omitempty" env:"SEMAPHORE_EMAIL_TLS"`
	EmailTlsMinVersion string `json:"email_tls_min_version,omitempty" default:"1.2" rule:"^(1\\.[0123])$" env:"SEMAPHORE_EMAIL_TLS_MIN_VERSION"`

	// ldap settings
	LdapEnable       bool          `json:"ldap_enable,omitempty" env:"SEMAPHORE_LDAP_ENABLE"`
	LdapBindDN       string        `json:"ldap_binddn,omitempty" env:"SEMAPHORE_LDAP_BIND_DN"`
	LdapBindPassword string        `json:"ldap_bindpassword,omitempty" env:"SEMAPHORE_LDAP_BIND_PASSWORD"`
	LdapServer       string        `json:"ldap_server,omitempty" env:"SEMAPHORE_LDAP_SERVER"`
	LdapSearchDN     string        `json:"ldap_searchdn,omitempty" env:"SEMAPHORE_LDAP_SEARCH_DN"`
	LdapSearchFilter string        `json:"ldap_searchfilter,omitempty" env:"SEMAPHORE_LDAP_SEARCH_FILTER"`
	LdapMappings     *LdapMappings `json:"ldap_mappings,omitempty"`
	LdapNeedTLS      bool          `json:"ldap_needtls,omitempty" env:"SEMAPHORE_LDAP_NEEDTLS"`

	// Telegram, Slack, Rocket.Chat, Microsoft Teams, DingTalk, and Gotify alerting
	TelegramAlert       bool   `json:"telegram_alert,omitempty" env:"SEMAPHORE_TELEGRAM_ALERT"`
	TelegramChat        string `json:"telegram_chat,omitempty" env:"SEMAPHORE_TELEGRAM_CHAT"`
	TelegramToken       string `json:"telegram_token,omitempty" env:"SEMAPHORE_TELEGRAM_TOKEN"`
	SlackAlert          bool   `json:"slack_alert,omitempty" env:"SEMAPHORE_SLACK_ALERT"`
	SlackUrl            string `json:"slack_url,omitempty" env:"SEMAPHORE_SLACK_URL"`
	RocketChatAlert     bool   `json:"rocketchat_alert,omitempty" env:"SEMAPHORE_ROCKETCHAT_ALERT"`
	RocketChatUrl       string `json:"rocketchat_url,omitempty" env:"SEMAPHORE_ROCKETCHAT_URL"`
	MicrosoftTeamsAlert bool   `json:"microsoft_teams_alert,omitempty" env:"SEMAPHORE_MICROSOFT_TEAMS_ALERT"`
	MicrosoftTeamsUrl   string `json:"microsoft_teams_url,omitempty" env:"SEMAPHORE_MICROSOFT_TEAMS_URL"`
	DingTalkAlert       bool   `json:"dingtalk_alert,omitempty" env:"SEMAPHORE_DINGTALK_ALERT"`
	DingTalkUrl         string `json:"dingtalk_url,omitempty" env:"SEMAPHORE_DINGTALK_URL"`
	GotifyAlert         bool   `json:"gotify_alert,omitempty" env:"SEMAPHORE_GOTIFY_ALERT"`
	GotifyUrl           string `json:"gotify_url,omitempty" env:"SEMAPHORE_GOTIFY_URL"`
	GotifyToken         string `json:"gotify_token,omitempty" env:"SEMAPHORE_GOTIFY_TOKEN"`

	// oidc settings
	OidcProviders map[string]OidcProvider `json:"oidc_providers,omitempty" env:"SEMAPHORE_OIDC_PROVIDERS"`

	MaxTaskDurationSec  int `json:"max_task_duration_sec,omitempty" env:"SEMAPHORE_MAX_TASK_DURATION_SEC"`
	MaxTasksPerTemplate int `json:"max_tasks_per_template,omitempty" env:"SEMAPHORE_MAX_TASKS_PER_TEMPLATE"`

	// task concurrency
	MaxParallelTasks int `json:"max_parallel_tasks,omitempty" default:"10" rule:"^[0-9]{1,10}$" env:"SEMAPHORE_MAX_PARALLEL_TASKS"`

	RunnerRegistrationToken string `json:"runner_registration_token,omitempty" env:"SEMAPHORE_RUNNER_REGISTRATION_TOKEN"`

	// feature switches
	PasswordLoginDisable     bool `json:"password_login_disable,omitempty" env:"SEMAPHORE_PASSWORD_LOGIN_DISABLED"`
	NonAdminCanCreateProject bool `json:"non_admin_can_create_project,omitempty" env:"SEMAPHORE_NON_ADMIN_CAN_CREATE_PROJECT"`

	UseRemoteRunner bool `json:"use_remote_runner,omitempty" env:"SEMAPHORE_USE_REMOTE_RUNNER"`

	IntegrationAlias string `json:"global_integration_alias,omitempty" env:"SEMAPHORE_INTEGRATION_ALIAS"`

	Apps map[string]App `json:"apps,omitempty" env:"SEMAPHORE_APPS"`

	Runner *RunnerConfig `json:"runner,omitempty"`

	EnvVars map[string]string `json:"env_vars,omitempty" env:"SEMAPHORE_ENV_VARS"`

	ForwardedEnvVars []string `json:"forwarded_env_vars,omitempty" env:"SEMAPHORE_FORWARDED_ENV_VARS"`

	Teams *TeamsConfig `json:"teams,omitempty"`

	Log *ConfigLog `json:"log,omitempty"`

	Process *ConfigProcess `json:"process,omitempty"`

	Schedule *ScheduleConfig `json:"schedule,omitempty"`

	Debugging *DebuggingConfig `json:"debugging,omitempty"`

	HA *HAConfig `json:"ha,omitempty"`
}

func NewConfigType() *ConfigType {
	return &ConfigType{
		LdapMappings: &LdapMappings{},
	}
}

// Config exposes the application configuration storage for use in the application
var Config *ConfigType

func ClearDir(dir string, preserveFiles bool, prefix string) error {
	d, err := os.Open(dir)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	defer d.Close() //nolint:errcheck

	files, err := d.ReadDir(0)
	if err != nil {
		return err
	}

	for _, f := range files {
		if preserveFiles && !f.IsDir() {
			continue
		}

		if prefix != "" && !strings.HasPrefix(f.Name(), prefix) {
			continue
		}

		err = os.RemoveAll(path.Join(dir, f.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func (conf *ConfigType) ClearTmpDir() error {
	return ClearDir(conf.TmpPath, false, "")
}

func (conf *ConfigType) GetProjectTmpDir(projectID int) string {
	return path.Join(conf.TmpPath, fmt.Sprintf("project_%d", projectID))
}

func (conf *ConfigType) ClearProjectTmpDir(projectID int) error {
	return ClearDir(conf.GetProjectTmpDir(projectID), false, "")
}

// ToJSON returns a JSON string of the config
func (conf *ConfigType) ToJSON() ([]byte, error) {
	return json.MarshalIndent(&conf, " ", "\t")
}

// ConfigInit reads in cli flags, and switches actions appropriately on them
func ConfigInit(configPath string, noConfigFile bool) (usedConfigPath *string) {
	fmt.Println("Loading config")

	Config = NewConfigType()
	Config.Apps = map[string]App{}

	if !noConfigFile {
		usedConfigPath = loadConfigFile(configPath)
	}

	loadConfigEnvironment()
	loadConfigDefaults()

	fmt.Println("Validating config")
	validateConfig()

	var encryption []byte

	hash, _ := base64.StdEncoding.DecodeString(Config.CookieHash)
	if len(Config.CookieEncryption) > 0 {
		encryption, _ = base64.StdEncoding.DecodeString(Config.CookieEncryption)
	}

	Cookie = securecookie.New(hash, encryption)

	if Config.WebHost != "" {
		var err error
		WebHostURL, err = url.Parse(Config.WebHost)
		if err != nil {
			panic(err)
		}

		if len(WebHostURL.String()) == 0 {
			WebHostURL = nil
		}
	} else {
		WebHostURL = nil
	}

	if Config.Runner != nil && Config.Runner.TokenFile != "" {
		runnerTokenBytes, err := os.ReadFile(Config.Runner.TokenFile)
		if err == nil {
			Config.Runner.Token = strings.TrimSpace(string(runnerTokenBytes))
		}
	}

	return
}

func loadConfigFile(configPath string) (usedConfigPath *string) {
	if configPath == "" {
		configPath = os.Getenv("SEMAPHORE_CONFIG_PATH")
	}

	// If the configPath option has been set try to load and decode it
	// var usedPath string

	if configPath == "" {
		cwd, err := os.Getwd()
		exitOnConfigFileError(err)
		paths := []string{
			path.Join(cwd, "config.json"),
			"/usr/local/etc/semaphore/config.json",
			"/etc/semaphore/config.json",
		}
		for _, p := range paths {
			_, err = os.Stat(p)
			if err != nil {
				continue
			}
			var file *os.File
			file, err = os.Open(p)
			if err != nil {
				continue
			}
			decodeConfig(file)
			usedConfigPath = &p
			break
		}
		exitOnConfigFileError(err)
	} else {
		p := configPath
		file, err := os.Open(p)
		exitOnConfigFileError(err)
		usedConfigPath = &p
		decodeConfig(file)
	}

	return
}

func loadDefaultsToObject(obj any) error {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = reflect.Indirect(v)
	}

	for i := 0; i < t.NumField(); i++ {
		fieldInfo := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.IsZero() && fieldInfo.Type.Kind() != reflect.Struct && fieldInfo.Type.Kind() != reflect.Map {
			continue
		}

		if fieldInfo.Type.Kind() == reflect.Struct {
			err := loadDefaultsToObject(fieldValue.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		} else if fieldInfo.Type.Kind() == reflect.Map {
			for _, key := range fieldValue.MapKeys() {
				val := fieldValue.MapIndex(key)

				if val.Type().Kind() != reflect.Struct {
					continue
				}

				newVal := reflect.New(val.Type())
				pointerValue := newVal.Elem()
				pointerValue.Set(val)

				err := loadDefaultsToObject(newVal.Interface())
				if err != nil {
					return err
				}

				fieldValue.SetMapIndex(key, newVal.Elem())
			}
			continue
		}

		defaultVar := fieldInfo.Tag.Get("default")
		if defaultVar == "" {
			continue
		}

		setConfigValue(fieldValue, defaultVar) // defaultVar always string!!!
	}

	return nil
}

func loadConfigDefaults() {
	err := loadDefaultsToObject(Config)
	if err != nil {
		panic(err)
	}
}

func castStringToInt(value string) int {
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return valueInt
}

func castStringToBool(value string) bool {
	var valueBool bool
	if value == "1" || strings.ToLower(value) == "true" || strings.ToLower(value) == "yes" {
		valueBool = true
	} else {
		valueBool = false
	}
	return valueBool
}

func AssignMapToStruct[P *S, S any](m map[string]any, s P) error {
	v := reflect.ValueOf(s).Elem()
	return assignMapToStructRecursive(m, v)
}

func cloneStruct(origValue reflect.Value) reflect.Value {
	// Create a new instance of the same type as the original struct
	cloneValue := reflect.New(origValue.Type()).Elem()

	// Iterate over the fields of the struct
	for i := 0; i < origValue.NumField(); i++ {
		// Get the field value
		fieldValue := origValue.Field(i)
		// Set the field value in the clone
		cloneValue.Field(i).Set(fieldValue)
	}

	// Return the cloned struct
	return cloneValue
}

func assignMapToStructRecursive(m map[string]any, structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		} else {
			jsonTag = strings.Split(jsonTag, ",")[0]
		}

		if value, ok := m[jsonTag]; ok {
			fieldValue := structValue.FieldByName(field.Name)
			if fieldValue.CanSet() {

				val := reflect.ValueOf(value)

				switch fieldValue.Kind() {
				case reflect.Struct:

					if val.Kind() != reflect.Map {
						return fmt.Errorf("expected map for nested struct field %s but got %T", field.Name, value)
					}

					mapValue, ok := value.(map[string]any)
					if !ok {
						return fmt.Errorf("cannot assign value of type %T to field %s of type %s", value, field.Name, field.Type)
					}
					err := assignMapToStructRecursive(mapValue, fieldValue)
					if err != nil {
						return err
					}
				case reflect.Map:
					if fieldValue.IsNil() {
						mapValue := reflect.MakeMap(fieldValue.Type())
						fieldValue.Set(mapValue)
					}

					// Handle map
					if val.Kind() != reflect.Map {
						return fmt.Errorf("expected map for field %s but got %T", field.Name, value)
					}

					for _, key := range val.MapKeys() {
						mapElemValue := val.MapIndex(key)
						mapElemType := fieldValue.Type().Elem()

						srcVal := fieldValue.MapIndex(key)
						var mapElem reflect.Value
						if srcVal.IsValid() {
							mapElem = cloneStruct(srcVal)
						} else {
							mapElem = reflect.New(mapElemType).Elem()
						}

						if mapElemType.Kind() == reflect.Struct {
							if err := assignMapToStructRecursive(mapElemValue.Interface().(map[string]any), mapElem); err != nil {
								return err
							}
						} else {
							if mapElemValue.Type().ConvertibleTo(mapElemType) {
								mapElem.Set(mapElemValue.Convert(mapElemType))
							} else {
								newVal, converted := CastValueToKind(mapElemValue.Interface(), mapElemType.Kind())
								if !converted {
									return fmt.Errorf("cannot assign value of type %s to map element of type %s",
										mapElemValue.Type(), mapElemType)
								}

								mapElem.Set(reflect.ValueOf(newVal))
							}
						}

						fieldValue.SetMapIndex(key, mapElem)
					}

				default:
					// Handle simple types
					if val.Type().ConvertibleTo(fieldValue.Type()) {
						fieldValue.Set(val.Convert(fieldValue.Type()))
					} else {

						newVal, converted := CastValueToKind(val.Interface(), fieldValue.Type().Kind())
						if !converted {
							return fmt.Errorf("cannot assign value of type %s to map element of type %s",
								val.Type(), val)
						}

						fieldValue.Set(reflect.ValueOf(newVal))
					}
				}
			}
		}
	}
	return nil
}

func CastValueToKind(value any, kind reflect.Kind) (res any, ok bool) {
	res = value

	switch kind {
	case reflect.String:
		ok = true
	case reflect.Int:
		if reflect.ValueOf(value).Kind() != reflect.Int {
			res = castStringToInt(fmt.Sprintf("%v", reflect.ValueOf(value)))
			ok = true
		}
	case reflect.Bool:
		if reflect.ValueOf(value).Kind() != reflect.Bool {
			res = castStringToBool(fmt.Sprintf("%v", reflect.ValueOf(value)))
			ok = true
		}
	default:
	}

	return
}

func setConfigValue(attribute reflect.Value, value string) {
	if attribute.IsValid() {
		kind := attribute.Kind()

		switch kind {
		case reflect.Slice:
			var arr []string
			err := json.Unmarshal([]byte(value), &arr)
			if err != nil {
				panic(err)
			}
			attribute.Set(reflect.ValueOf(arr))
		case reflect.Map:
			mapType := attribute.Type()
			mapValue := reflect.New(mapType)
			err := json.Unmarshal([]byte(value), mapValue.Interface())
			if err != nil {
				panic(err)
			}
			attribute.Set(mapValue.Elem())
		default:
			newValue, _ := CastValueToKind(value, kind)
			attribute.Set(reflect.ValueOf(newValue))
		}

	} else {
		panic(fmt.Errorf("got non-existent config attribute"))
	}
}

func getConfigValue(path string) string {
	attribute := reflect.ValueOf(Config)
	nested_path := strings.Split(path, ".")

	for i, nested := range nested_path {
		attribute = reflect.Indirect(attribute).FieldByName(nested)
		lastDepth := len(nested_path) == i+1
		if !lastDepth && attribute.Kind() != reflect.Struct && attribute.Kind() != reflect.Pointer ||
			lastDepth && attribute.Kind() == reflect.Invalid {
			panic(fmt.Errorf("got non-existent config attribute '%v'", path))
		}
	}

	return fmt.Sprintf("%v", attribute)
}

func validate(value any) error {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = reflect.Indirect(v)
	}

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		rule := fieldType.Tag.Get("rule")
		if rule == "" {
			continue
		}

		var strVal string

		if fieldType.Type.Kind() == reflect.Int {
			strVal = strconv.FormatInt(fieldValue.Int(), 10)
		} else if fieldType.Type.Kind() == reflect.Uint {
			strVal = strconv.FormatUint(fieldValue.Uint(), 10)
		} else {
			strVal = fieldValue.String()
		}

		match, _ := regexp.MatchString(rule, strVal)

		if match {
			continue
		}

		fieldName := strings.ToLower(fieldType.Name)

		if strings.Contains(fieldName, "password") || strings.Contains(fieldName, "secret") || strings.Contains(fieldName, "key") {
			strVal = "***"
		}

		return fmt.Errorf(
			"value of field '%v' is not valid: %v (Must match regex: '%v')",
			fieldType.Name, strVal, rule,
		)
	}

	return nil
}

func validateConfig() {
	err := validate(Config)
	if err != nil {
		panic(err)
	}
}

func loadEnvironmentToObject(obj any) error {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = reflect.Indirect(v)
	}

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		if fieldType.Type.Kind() == reflect.Struct {
			err := loadEnvironmentToObject(fieldValue.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		} else if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
			if fieldValue.IsZero() {
				newValue := reflect.New(fieldType.Type.Elem())
				fieldValue.Set(newValue)
			}
			err := loadEnvironmentToObject(fieldValue.Interface())
			if err != nil {
				return err
			}
			continue
		}

		envVar := fieldType.Tag.Get("env")
		if envVar == "" {
			continue
		}

		envValue, exists := os.LookupEnv(envVar)

		if !exists {
			continue
		}

		setConfigValue(fieldValue, envValue) // envValue always string!!!
	}

	return nil
}

func loadConfigEnvironment() {
	err := loadEnvironmentToObject(Config)
	if err != nil {
		panic(err)
	}
}

func exitOnConfigError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func exitOnConfigFileError(err error) {
	if err != nil {
		exitOnConfigError("Cannot Find configuration! Use --config parameter to point to a JSON file generated by `semaphore setup`.")
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

// FindSemaphore looks in the PATH for the semaphore variable
// if not found it will attempt to find the absolute path of the first
// os argument, the semaphore command, and return it
func FindSemaphore() string {
	cmdPath, _ := exec.LookPath("semaphore") //nolint: gas

	if len(cmdPath) == 0 {
		cmdPath, _ = filepath.Abs(os.Args[0]) // nolint: gas
	}

	return cmdPath
}

func AnsibleVersion() string {
	bytes, err := exec.Command("ansible", "--version").Output()
	if err != nil {
		return ""
	}
	return string(bytes)
}

// CheckUpdate uses the GitHub client to check for new tags in the semaphore repo
func CheckUpdate() (updateAvailable *github.RepositoryRelease, err error) {
	// fetch releases
	gh := github.NewClient(nil)
	releases, _, err := gh.Repositories.ListReleases(context.TODO(), "semaphoreui", "semaphore", nil)
	if err != nil {
		return
	}

	updateAvailable = nil
	if (*releases[0].TagName)[1:] != Version() {
		updateAvailable = releases[0]
	}

	return
}

func (d *DbConfig) IsPresent() bool {
	return d.GetHostname() != ""
}

func (d *DbConfig) HasSupportMultipleDatabases() bool {
	return true
}

func (d *DbConfig) GetDbName() string {
	dbName := os.Getenv("SEMAPHORE_DB_NAME")
	if dbName != "" {
		return dbName
	}
	return d.DbName
}

func (d *DbConfig) GetUsername() string {
	username := os.Getenv("SEMAPHORE_DB_USER")
	if username != "" {
		return username
	}
	return d.Username
}

func (d *DbConfig) GetPassword() string {
	password := os.Getenv("SEMAPHORE_DB_PASS")
	if password != "" {
		return password
	}
	return d.Password
}

func (d *DbConfig) GetHostname() string {
	hostname := os.Getenv("SEMAPHORE_DB_HOST")
	if hostname != "" {
		return hostname
	}
	return d.Hostname
}

// GetConnectionString constructs the database connection string based on the current configuration.
// It supports MySQL, BoltDB, and PostgreSQL dialects.
// If the dialect is unsupported, it returns an error.
//
// Parameters:
// - includeDbName: a boolean indicating whether to include the database name in the connection string.
//
// Returns:
// - connectionString: the constructed database connection string.
// - err: an error if the dialect is unsupported.
func (d *DbConfig) GetConnectionString(includeDbName bool) (connectionString string, err error) {
	dbName := d.GetDbName()
	dbUser := d.GetUsername()
	dbPass := d.GetPassword()
	dbHost := d.GetHostname()

	switch d.Dialect {
	case DbDriverBolt:
		connectionString = dbHost
	case DbDriverMySQL:
		if includeDbName {
			connectionString = fmt.Sprintf(
				"%s:%s@tcp(%s)/%s",
				dbUser,
				dbPass,
				dbHost,
				dbName)
		} else {
			connectionString = fmt.Sprintf(
				"%s:%s@tcp(%s)/",
				dbUser,
				dbPass,
				dbHost)
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
				dbUser,
				url.QueryEscape(dbPass),
				dbHost,
				dbName)
		} else {
			connectionString = fmt.Sprintf(
				"postgres://%s:%s@%s/postgres",
				dbUser,
				url.QueryEscape(dbPass),
				dbHost)
		}
		connectionString += mapToQueryString(d.Options)
	case DbDriverSQLite:
		connectionString = "file:" + dbHost
		connectionString += mapToQueryString(d.Options)
	default:
		err = fmt.Errorf("unsupported database driver: %s", d.Dialect)
	}
	return
}

// PrintDbInfo prints the database connection information based on the current configuration.
// It retrieves the database dialect and prints the corresponding connection details.
// If the dialect is not found, it panics with an error message.
func (conf *ConfigType) PrintDbInfo() {
	// Get the database dialect
	dialect, err := conf.GetDialect()
	if err != nil {
		panic(err)
	}

	// Print database connection information based on the dialect
	switch dialect {
	case DbDriverMySQL:
		fmt.Printf("MySQL %v@%v %v\n", conf.MySQL.GetUsername(), conf.MySQL.GetHostname(), conf.MySQL.GetDbName())
	case DbDriverBolt:
		fmt.Printf("BoltDB %v\n", conf.BoltDb.GetHostname())
	case DbDriverPostgres:
		fmt.Printf("Postgres %v@%v %v\n", conf.Postgres.GetUsername(), conf.Postgres.GetHostname(), conf.Postgres.GetDbName())
	case DbDriverSQLite:
		fmt.Printf("SQLite %v@%v %v\n", conf.SQLite.GetUsername(), conf.SQLite.GetHostname(), conf.SQLite.GetDbName())
	default:
		panic(fmt.Errorf("database configuration not found"))
	}
}

func (conf *ConfigType) GetDialect() (dialect string, err error) {
	if conf.Dialect == "" {
		switch {
		case conf.MySQL.IsPresent():
			dialect = DbDriverMySQL
		case conf.BoltDb.IsPresent():
			dialect = DbDriverBolt
		case conf.Postgres.IsPresent():
			dialect = DbDriverPostgres
		case conf.SQLite.IsPresent():
			dialect = DbDriverSQLite
		default:
			err = errors.New("database configuration not found")
		}
		return
	}

	dialect = conf.Dialect
	return
}

func (conf *ConfigType) GetDBConfig() (dbConfig DbConfig, err error) {
	var dialect string
	dialect, err = conf.GetDialect()
	if err != nil {
		return
	}

	switch dialect {
	case DbDriverBolt:
		dbConfig = *conf.BoltDb
	case DbDriverPostgres:
		dbConfig = *conf.Postgres
	case DbDriverSQLite:
		dbConfig = *conf.SQLite
	case DbDriverMySQL:
		dbConfig = *conf.MySQL
	default:
		err = errors.New("database configuration not found")
	}

	dbConfig.Dialect = dialect

	return
}

// GenerateSecrets generates cookie secret during setup
func (conf *ConfigType) GenerateSecrets() {
	hash := securecookie.GenerateRandomKey(32)
	encryption := securecookie.GenerateRandomKey(32)
	accessKeyEncryption := securecookie.GenerateRandomKey(32)

	conf.CookieHash = base64.StdEncoding.EncodeToString(hash)
	conf.CookieEncryption = base64.StdEncoding.EncodeToString(encryption)
	conf.AccessKeyEncryption = base64.StdEncoding.EncodeToString(accessKeyEncryption)
}

var appCommands = map[string]string{
	"ansible":    "ansible-playbook",
	"terraform":  "terraform",
	"tofu":       "tofu",
	"terragrunt": "terragrunt",
	"bash":       "bash",
}

var appPriorities = map[string]int{
	"ansible":    1000,
	"terraform":  900,
	"tofu":       800,
	"terragrunt": 850,
	"bash":       700,
	"powershell": 600,
	"python":     500,
}

func LookupDefaultApps() {
	for appID, cmd := range appCommands {
		if _, ok := Config.Apps[appID]; ok {
			continue
		}

		_, err := exec.LookPath(cmd)
		if err != nil {
			continue
		}

		if Config.Apps == nil {
			Config.Apps = make(map[string]App)
		}

		Config.Apps[appID] = App{
			Active: true,
		}
	}

	for k, v := range appPriorities {
		app := Config.Apps[k]
		if app.Priority <= 0 {
			app.Priority = v
		}
		Config.Apps[k] = app
	}
}

func GetPublicHost() string {
	aliasURL := Config.WebHost
	port := Config.Port
	if port == "" {
		port = "3000"
	}

	if strings.HasPrefix(port, ":") {
		port = port[1:]
	}

	if aliasURL == "" {
		aliasURL = "http://localhost:" + port
	}

	return aliasURL
}

func GetPublicAliasURL(scope string, alias string) string {
	aliasURL := GetPublicHost()

	if !strings.HasSuffix(aliasURL, "/") {
		aliasURL += "/"
	}

	aliasURL += "api/" + scope + "/" + alias

	return aliasURL
}

func GenerateRecoveryCode() (code string, hash string, err error) {
	buf := make([]byte, 10)
	_, err = io.ReadFull(rand.Reader, buf)
	if err != nil {
		return
	}

	code = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	hash = string(hashBytes)
	return
}

func VerifyRecoveryCode(inputCode, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputCode))
	return err == nil
}
