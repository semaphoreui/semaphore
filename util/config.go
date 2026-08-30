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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
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

const (
	// HomeDirModeUserHome does not override HOME.
	// Sets ANSIBLE_HOME per template to isolate .ansible/ across parallel tasks.
	HomeDirModeUserHome = "user_home"

	// HomeDirModeProjectHome sets HOME to the project temp directory.
	// This is the legacy behavior. Parallel ansible-galaxy runs may conflict.
	HomeDirModeProjectHome = "project_home"

	// HomeDirModeTemplateDir does not override HOME.
	// Sets ANSIBLE_HOME to a per-template "_home/.ansible" directory
	// (e.g. repository_15_template_114_home/.ansible) to isolate
	// .ansible/ artifacts across parallel tasks.
	HomeDirModeTemplateDir = "template_dir"
)

type DbConfig struct {
	Dialect string `json:"-"`

	Hostname string            `json:"host,omitempty" env:"SEMAPHORE_DB_HOST" default:"0.0.0.0"`
	Username string            `json:"user,omitempty" env:"SEMAPHORE_DB_USER"`
	Password string            `json:"pass,omitempty" env:"SEMAPHORE_DB_PASS,sensitive"`
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

// RunnerConnectionConfig controls how the runner connects to the
// Semaphore server.
type RunnerConnectionConfig struct {
	// ServerCACertFile is a PEM bundle used to verify the Semaphore
	// server's certificate, in addition to the system trust store.
	// Set this when the server uses a self-signed or internal-CA cert.
	ServerCACertFile string `json:"server_ca_cert_file,omitempty" env:"SEMAPHORE_RUNNER_SERVER_CA_CERT_FILE"`
	// SkipTLSVerify disables server certificate verification entirely.
	// This is insecure (vulnerable to MITM) — use only for testing.
	SkipTLSVerify bool `json:"skip_tls_verify,omitempty" env:"SEMAPHORE_RUNNER_SKIP_TLS_VERIFY"`
}

// ExecutorType identifies the strategy the runner uses to execute each task. The default
// "local" executor runs tasks as subprocesses on the runner host. "kubernetes" dispatches
// each task into an ephemeral pod, GitLab-runner-style.
type ExecutorType string

const (
	ExecutorTypeLocal      ExecutorType = "local"
	ExecutorTypeKubernetes ExecutorType = "k8s"
	ExecutorTypeDocker     ExecutorType = "docker"
)

type ExecutorConfig struct {
	Type   ExecutorType       `json:"type" default:"local" env:"SEMAPHORE_RUNNER_EXECUTOR_TYPE"`
	K8s    RunnerK8sConfig    `json:"k8s"`
	Docker RunnerDockerConfig `json:"docker"`
}

type RunnerConfig struct {
	RegistrationToken     string `json:"-" env:"SEMAPHORE_RUNNER_REGISTRATION_TOKEN"`
	RegistrationTokenFile string `json:"registration_token_file,omitempty" env:"SEMAPHORE_RUNNER_REGISTRATION_TOKEN_FILE"`
	Token                 string `json:"token,omitempty" env:"SEMAPHORE_RUNNER_TOKEN,sensitive"`
	TokenFile             string `json:"token_file,omitempty" env:"SEMAPHORE_RUNNER_TOKEN_FILE"`

	// OneOff indicates than runner runs only one job and exit. It is very useful for dynamic runners.
	// How it works?
	// Example:
	// 1) User starts the task.
	// 2) Semaphore found runner for task and calls runner's webhook if it provided.
	// 3) Your server or lambda handling the call and starts the one-off runner.
	// 4) The runner connects to the Semaphore server and handles the enqueued task(s).
	OneOff bool `json:"one_off,omitempty" env:"SEMAPHORE_RUNNER_ONE_OFF"`

	Enabled          bool     `json:"enabled,omitempty" env:"SEMAPHORE_RUNNER_ENABLED"`
	Webhook          string   `json:"webhook,omitempty" env:"SEMAPHORE_RUNNER_WEBHOOK"`
	Name             string   `json:"name,omitempty" env:"SEMAPHORE_RUNNER_NAME"`
	Tags             []string `json:"tags,omitempty" env:"SEMAPHORE_RUNNER_TAGS"`
	MaxParallelTasks int      `json:"max_parallel_tasks,omitempty" default:"9999" env:"SEMAPHORE_RUNNER_MAX_PARALLEL_TASKS"`
	ProjectID        *int     `json:"project_id,omitempty" env:"SEMAPHORE_RUNNER_PROJECT_ID"`

	Connection *RunnerConnectionConfig `json:"connection,omitempty"`

	Executor *ExecutorConfig `json:"executor,omitempty" env:"SEMAPHORE_RUNNER_EXECUTOR"`
}

// RunnerK8sConfig holds runner-side configuration for the Kubernetes executor. Field
// shape mirrors the GitLab runner Kubernetes executor for familiarity. Empty values
// fall back to the documented defaults at consumption time (see pro/services/tasks/k8s).
type RunnerK8sConfig struct {
	// KubeconfigPath is the path to a kubeconfig file. When empty, in-cluster
	// configuration is used (ServiceAccount token + CA cert mounted by Kubernetes).
	KubeconfigPath string `json:"kubeconfig,omitempty" env:"SEMAPHORE_RUNNER_K8S_KUBECONFIG"`

	// Namespace is where ephemeral task Pods are created.
	Namespace string `json:"namespace,omitempty" default:"semaphore" env:"SEMAPHORE_RUNNER_K8S_NAMESPACE"`

	// Image is the default container image used for the build container of each
	// task Pod. Templates may override this in a future phase.
	Image string `json:"image,omitempty" default:"semaphoreui/job:latest" env:"SEMAPHORE_RUNNER_K8S_IMAGE"`

	// HelperImage is the image used for the git-clone init container (Phase 3+).
	HelperImage string `json:"helper_image,omitempty" default:"semaphoreui/helper:latest" env:"SEMAPHORE_RUNNER_K8S_HELPER_IMAGE"`

	// ServiceAccount that task Pods run under. Defaults to the namespace's default SA.
	ServiceAccount string `json:"service_account,omitempty" default:"default" env:"SEMAPHORE_RUNNER_K8S_SERVICE_ACCOUNT"`

	// PullSecrets is a comma-separated list of imagePullSecrets attached to each Pod.
	PullSecrets string `json:"pull_secrets,omitempty" env:"SEMAPHORE_RUNNER_K8S_PULL_SECRETS"`

	// PollIntervalSeconds controls how often the executor polls Pod status. Defaults
	// to 3 seconds. Kept as a plain int (not time.Duration) for env-binding simplicity.
	PollIntervalSeconds int `json:"poll_interval_seconds,omitempty" default:"3" env:"SEMAPHORE_RUNNER_K8S_POLL_INTERVAL_SECONDS"`

	// CleanupGraceSeconds is the grace period when deleting Pods. Defaults to 30s.
	CleanupGraceSeconds int `json:"cleanup_grace_seconds,omitempty" default:"30" env:"SEMAPHORE_RUNNER_K8S_CLEANUP_GRACE_SECONDS"`
}

// RunnerDockerConfig holds runner-side configuration for the Docker executor. Each task
// runs in an ephemeral container created against a local or remote Docker daemon,
// GitLab-Docker-executor-style. Empty values fall back to the documented defaults at
// consumption time (see pro/services/tasks/docker).
type RunnerDockerConfig struct {
	// Host is the Docker daemon URL. Supports unix://, tcp:// and npipe:// schemes.
	// When empty the standard environment (DOCKER_HOST) and the platform default
	// socket are used.
	Host string `json:"host,omitempty" env:"SEMAPHORE_RUNNER_DOCKER_HOST"`

	// TLSVerify enables TLS certificate verification for tcp:// connections.
	TLSVerify bool `json:"tls_verify,omitempty" env:"SEMAPHORE_RUNNER_DOCKER_TLS_VERIFY"`

	// CertPath is the directory holding ca.pem, cert.pem and key.pem for mutual TLS
	// against a remote daemon.
	CertPath string `json:"cert_path,omitempty" env:"SEMAPHORE_RUNNER_DOCKER_CERT_PATH"`

	// Image is the default image used for the build container of each task.
	Image string `json:"image,omitempty" default:"semaphoreui/job:latest" env:"SEMAPHORE_RUNNER_DOCKER_IMAGE"`

	// HelperImage is the image used for the transient git-clone container.
	HelperImage string `json:"helper_image,omitempty" default:"semaphoreui/helper:latest" env:"SEMAPHORE_RUNNER_DOCKER_HELPER_IMAGE"`

	// Network is the Docker network the build container joins. Defaults to "bridge".
	Network string `json:"network,omitempty" default:"bridge" env:"SEMAPHORE_RUNNER_DOCKER_NETWORK"`

	// PullPolicy controls image pulling: always, if-not-present or never.
	PullPolicy string `json:"pull_policy,omitempty" default:"if-not-present" env:"SEMAPHORE_RUNNER_DOCKER_PULL_POLICY"`

	// CPULimit, when > 0, caps the build container CPU (passed as --cpus).
	CPULimit float64 `json:"cpu_limit,omitempty" env:"SEMAPHORE_RUNNER_DOCKER_CPU_LIMIT"`

	// MemoryLimit, when non-empty, caps the build container memory (e.g. "2g").
	MemoryLimit string `json:"memory_limit,omitempty" env:"SEMAPHORE_RUNNER_DOCKER_MEMORY_LIMIT"`

	// PollIntervalSeconds controls how often container status is polled. Defaults to 2s.
	PollIntervalSeconds int `json:"poll_interval_seconds,omitempty" default:"2" env:"SEMAPHORE_RUNNER_DOCKER_POLL_INTERVAL_SECONDS"`

	// CleanupGraceSeconds is the timeout passed to docker stop. Defaults to 30s.
	CleanupGraceSeconds int `json:"cleanup_grace_seconds,omitempty" default:"30" env:"SEMAPHORE_RUNNER_DOCKER_CLEANUP_GRACE_SECONDS"`

	// Privileged runs the build container with --privileged. Dangerous; off by default.
	Privileged bool `json:"privileged,omitempty" env:"SEMAPHORE_RUNNER_DOCKER_PRIVILEGED"`
}

type DefultGlobalRunnerMode string

const (
	DefultGlobalRunnerNone    DefultGlobalRunnerMode = ""
	DefultGlobalRunnerDisable DefultGlobalRunnerMode = "disable"
	DefultGlobalRunnerPrefer  DefultGlobalRunnerMode = "prefer"
	DefultGlobalRunnerRequire DefultGlobalRunnerMode = "require"
)

// RunnersConfig holds server-side settings describing how the server treats
// its runner fleet. It is unrelated to RunnerConfig, which configures a runner
// process itself: server-side fleet settings use the SEMAPHORE_RUNNERS_* env
// prefix, runner-process settings use SEMAPHORE_RUNNER_*.
type RunnersConfig struct {
	// OfflineTimeoutSec is the heartbeat staleness after which a runner is
	// considered offline: it receives no new tasks and its "starting" tasks
	// are reassigned to another runner. Must be comfortably larger than the
	// runner poll interval (a few multiples) so a healthy-but-slow runner is
	// never marked offline.
	OfflineTimeoutSec int `json:"offline_timeout_sec,omitempty" default:"120" env:"SEMAPHORE_RUNNERS_OFFLINE_TIMEOUT_SEC"`

	// TaskFailTimeoutSec is the heartbeat staleness after which a runner's
	// "running" tasks are failed. Between OfflineTimeoutSec and this value a
	// running task is deliberately left alone: an offline runner may still be
	// executing its jobs and resumes reporting if it reconnects in time.
	// Values below OfflineTimeoutSec are clamped to it.
	TaskFailTimeoutSec int `json:"task_fail_timeout_sec,omitempty" default:"420" env:"SEMAPHORE_RUNNERS_TASK_FAIL_TIMEOUT_SEC"`

	// ReconcileIntervalSec is how often the server scans dispatched tasks
	// against runner liveness.
	ReconcileIntervalSec int `json:"reconcile_interval_sec,omitempty" default:"30" env:"SEMAPHORE_RUNNERS_RECONCILE_INTERVAL_SEC"`

	// RunnerRegistrationToken is deprecated, use Runners field instead of it.
	RegistrationToken string `json:"registration_token,omitempty" env:"SEMAPHORE_RUNNER_REGISTRATION_TOKEN"`

	DefaultGlobalRunnersMode DefultGlobalRunnerMode `json:"default_global_runners_mode" env:"SEMAPHORE_DEFAULT_GLOBAL_RUNNERS_MODE"`
}

type TLSConfig struct {
	Enabled          bool   `json:"enabled" env:"SEMAPHORE_TLS_ENABLED"`
	CertFile         string `json:"cert_file" env:"SEMAPHORE_TLS_CERT_FILE"`
	KeyFile          string `json:"key_file" env:"SEMAPHORE_TLS_KEY_FILE"`
	HTTPRedirectAddr string `json:"http_redirect_addr,omitempty" env:"SEMAPHORE_TLS_HTTP_REDIRECT_ADDR"`
	HTTPRedirectPort *int   `json:"http_redirect_port,omitempty" env:"SEMAPHORE_TLS_HTTP_REDIRECT_PORT"`
}

type TotpConfig struct {
	Enabled       bool   `json:"enabled" env:"SEMAPHORE_TOTP_ENABLED"`
	AllowRecovery bool   `json:"allow_recovery" env:"SEMAPHORE_TOTP_ALLOW_RECOVERY"`
	Issuer        string `json:"app_name" env:"SEMAPHORE_TOTP_ISSUER"`
}

type EventLogType struct {
	Format  string             `json:"format,omitempty" env:"SEMAPHORE_EVENT_LOG_FORMAT"`
	Enabled bool               `json:"enabled" env:"SEMAPHORE_EVENT_LOG_ENABLED"`
	Logger  *lumberjack.Logger `json:"logger,omitempty" env:"SEMAPHORE_EVENT_LOGGER"`
}

const (
	FileLogJSON string = "json"
	FileLogRaw  string = ""
)

type TaskLogType struct {
	Enabled      bool               `json:"enabled" env:"SEMAPHORE_TASK_LOG_ENABLED"`
	Format       string             `json:"format,omitempty" env:"SEMAPHORE_TASK_LOG_FORMAT"`
	Logger       *lumberjack.Logger `json:"logger,omitempty" env:"SEMAPHORE_TASK_LOGGER"`
	ResultLogger *lumberjack.Logger `json:"result_logger,omitempty" env:"SEMAPHORE_TASK_RESULT_LOGGER"`
}

type ConfigLog struct {
	Events *EventLogType `json:"events,omitempty"`
	Tasks  *TaskLogType  `json:"tasks,omitempty"`
}

type SyslogFormat string

const (
	SyslogDefault SyslogFormat = ""
	SyslogRFC5424 SyslogFormat = "rfc5424"
)

type SyslogConfig struct {
	Enabled bool         `json:"enabled" env:"SEMAPHORE_SYSLOG_ENABLED"`
	Network string       `json:"network,omitempty" env:"SEMAPHORE_SYSLOG_NETWORK"`
	Address string       `json:"address,omitempty" env:"SEMAPHORE_SYSLOG_ADDRESS"`
	Tag     string       `json:"tag,omitempty" env:"SEMAPHORE_SYSLOG_TAG"`
	Format  SyslogFormat `json:"format,omitempty" env:"SEMAPHORE_SYSLOG_FORMAT"`
}

type MetricsConfig struct {
	Enabled  bool   `json:"enabled" env:"SEMAPHORE_METRICS_ENABLED"`
	Username string `json:"username,omitempty" env:"SEMAPHORE_METRICS_USERNAME"`
	Password string `json:"password,omitempty" env:"SEMAPHORE_METRICS_PASSWORD,sensitive"`
}

type ConfigProcess struct {
	User       string  `json:"user,omitempty" env:"SEMAPHORE_PROCESS_USER"`
	UID        *uint32 `json:"uid,omitempty" env:"SEMAPHORE_PROCESS_UID"`
	Chroot     string  `json:"chroot,omitempty" env:"SEMAPHORE_PROCESS_CHROOT"`
	GID        *uint32 `json:"gid,omitempty" env:"SEMAPHORE_PROCESS_GID"`
	NoNewPrivs bool    `json:"no_new_privs,omitempty" env:"SEMAPHORE_PROCESS_NO_NEW_PRIVS"`

	// AppNamespaces controls Linux namespace isolation for child apps
	// (ansible, terraform, shell templates). Git is never isolated —
	// SSH agent forwarding and credential helpers need host access.
	AppNamespaces ConfigAppNamespaces `json:"app_namespaces"`
}

// ConfigAppNamespaces mirrors the CLONE_NEW* flags applied to app runs.
// Each flag is a standard Linux namespace and is a no-op on non-Linux.
type ConfigAppNamespaces struct {
	// User isolates UIDs/GIDs (CLONE_NEWUSER). Enables unprivileged use
	// of the other namespaces.
	User bool `json:"user,omitempty" env:"SEMAPHORE_PROCESS_APP_NS_USER"`
	// Mount hides host mount points such as secret tmpfs (CLONE_NEWNS).
	Mount bool `json:"mount,omitempty" env:"SEMAPHORE_PROCESS_APP_NS_MOUNT"`
	// PID hides host processes from child apps (CLONE_NEWPID).
	PID bool `json:"pid,omitempty" env:"SEMAPHORE_PROCESS_APP_NS_PID"`
	// IPC isolates SysV IPC and POSIX message queues (CLONE_NEWIPC).
	IPC bool `json:"ipc,omitempty" env:"SEMAPHORE_PROCESS_APP_NS_IPC"`
	// UTS isolates hostname and domain (CLONE_NEWUTS).
	UTS bool `json:"uts,omitempty" env:"SEMAPHORE_PROCESS_APP_NS_UTS"`
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
	Pass          string `json:"pass,omitempty" env:"SEMAPHORE_HA_REDIS_PASS,sensitive"`
	User          string `json:"user,omitempty" env:"SEMAPHORE_HA_REDIS_USER"`
	TLS           bool   `json:"tls,omitempty" env:"SEMAPHORE_HA_REDIS_TLS"`
	TLSSkipVerify bool   `json:"tls_skip_verify,omitempty" env:"SEMAPHORE_HA_REDIS_TLS_SKIP_VERIFY"`
}

type HAConfig struct {
	Enabled bool           `json:"enabled" env:"SEMAPHORE_HA_ENABLED"`
	NodeID  string         `json:"node_id,omitempty" env:"SEMAPHORE_HA_NODE_ID"` // auto-generated if empty
	Redis   *HARedisConfig `json:"redis,omitempty"`
}

// HAEnabled returns true when high-availability mode is configured.
func HAEnabled() bool {
	return Config.HA != nil && Config.HA.Enabled
}

// InitHANodeID generates a unique node identifier for this instance if one
// was not explicitly configured. Must be called after ConfigInit.
func InitHANodeID() {
	if Config.HA == nil {
		return
	}
	if Config.HA.NodeID == "" {
		Config.HA.NodeID = RandString(16)
	}
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

type ConfigDirs struct {
	Secrets         string `json:"secrets,omitempty" env:"SEMAPHORE_SECRETS_PATH" default:"/tmp/semaphore"`
	Repos           string `json:"repos,omitempty" env:"SEMAPHORE_REPOS_DIR"`
	SSHAgentSockets string `json:"ssh_agent_sockets,omitempty" env:"SEMAPHORE_SSH_AGENT_SOCKETS_DIR" default:"/tmp/semaphore"`
}

// JWTConfig issuance for task executions (used by playbooks to authenticate to
type JWTConfig struct {
	Enabled    bool   `json:"enabled,omitempty" env:"SEMAPHORE_JWT_ENABLED"`
	Issuer     string `json:"issuer,omitempty" env:"SEMAPHORE_JWT_ISSUER"`
	DefaultTTL string `json:"default_ttl,omitempty" env:"SEMAPHORE_JWT_DEFAULT_TTL" default:"1h"`
	MaxTTL     string `json:"max_ttl,omitempty" env:"SEMAPHORE_JWT_MAX_TTL" default:"24h"`
}

// KeySource supplies a single secret key either inline (Value) or from a file
// (File). Value and File are mutually exclusive.
type KeySource struct {
	Value string `json:"value,omitempty"`
	File  string `json:"file,omitempty"`
}

// ActivePointers names the active (encrypting) key per purpose. A key may be
// named by label (SecretKey/OptionKey, into the keys map) or by filename
// (SecretKeyFile/OptionKeyFile, a file in KeysFolder, relative to it). The
// label/filename is human-facing only; the id stored in the database is derived
// from the key material.
type ActivePointers struct {
	SecretKey     string `json:"secret_key,omitempty"`
	OptionKey     string `json:"option_key,omitempty"`
	SecretKeyFile string `json:"secret_key_file,omitempty"`
	OptionKeyFile string `json:"option_key_file,omitempty"`
}

// EncryptionKeysConfig is the content of the keys file: a registry of keys plus
// pointers to the active key per purpose. The registry can be an inline map
// (Keys: label -> source) and/or a folder of key files (KeysFolder: each regular
// file is one key, labelled by its filename); the two combine. Decryption looks a
// key up by the content-addressed id stamped into each ciphertext, so a key is
// "retired" simply by no longer being active while some rows still reference it.
// access_key protects Access Key secrets in the database; option_key protects
// encrypted DB options (the JWT signing key) and falls back to the access key.
type EncryptionKeysConfig struct {
	Keys       map[string]KeySource `json:"keys,omitempty"`
	KeysFolder string               `json:"keys_folder,omitempty"`
	Active     ActivePointers       `json:"active"`
}

// EncryptionConfig is the main-config "encryption" section. It points at the
// separate keys file and controls how often that file is polled for changes.
type EncryptionConfig struct {
	// KeysFile is the path to the EncryptionKeysConfig file (the keyrings).
	KeysFile string `json:"keys_file,omitempty" env:"SEMAPHORE_ENCRYPTION_KEYS_FILE"`
	// KeysPollInterval is how often the keys file is checked for changes (a Go
	// duration like "15s"). "0" disables polling (SIGHUP still forces a reload).
	KeysPollInterval string `json:"keys_poll_interval,omitempty" env:"SEMAPHORE_ENCRYPTION_KEYS_POLL_INTERVAL" default:"15s"`
}

type SshStrictHostKeyChecking string

const (
	SshStrictHostKeyCheckingNo        SshStrictHostKeyChecking = "no"
	SshStrictHostKeyCheckingYes       SshStrictHostKeyChecking = "yes"
	SshStrictHostKeyCheckingAcceptNew SshStrictHostKeyChecking = "accept-new"
)

type SshConfig struct {
	// SshConfigPath is a path to the custom SSH config file.
	// Default path is ~/.ssh/config.
	ConfigPath string `json:"config_path,omitempty" env:"SEMAPHORE_SSH_PATH"`

	// SshKnownHostsFile is a path to the SSH known_hosts file used to verify git
	// server host keys. When set, host-key checking is strict: a key that is
	// missing from (or changed relative to) this file aborts the connection,
	// preventing a network attacker from impersonating the git server. When
	// empty, Semaphore uses a persistent trust-on-first-use file under TmpPath
	// (StrictHostKeyChecking=accept-new): the first connection to a host is
	// trusted and pinned, and any later host-key change is rejected.
	KnownHostsFile string `json:"known_hosts_file,omitempty" env:"SEMAPHORE_SSH_KNOWN_HOSTS_FILE"`

	StrictHostKeyChecking SshStrictHostKeyChecking `json:"strict_host_key_checking,omitempty" env:"" default:"no"`
}

// ConfigType mapping between Config and the json file that sets it
type ConfigType struct {
	MySQL    *DbConfig `json:"mysql,omitempty"`
	Postgres *DbConfig `json:"postgres,omitempty"`
	SQLite   *DbConfig `json:"sqlite,omitempty"`

	Dialect string `json:"dialect,omitempty" default:"sqlite" rule:"^mysql|postgres|sqlite$" env:"SEMAPHORE_DB_DIALECT"`

	// Format `:port_num` eg, :3000
	// if : is missing it will be corrected
	Port string     `json:"port,omitempty" default:":3000" rule:"^:?([0-9]{1,5})$" env:"SEMAPHORE_PORT"`
	TLS  *TLSConfig `json:"tls,omitempty"`

	Mfa *MultifactorAuthConfig `json:"mfa,omitempty"`

	// Interface ip, put in front of the port.
	// defaults to empty
	Interface string `json:"interface,omitempty" env:"SEMAPHORE_INTERFACE"`

	// semaphore stores ephemeral projects here
	TmpPath string `json:"tmp_path,omitempty" default:"/tmp/semaphore" env:"SEMAPHORE_TMP_PATH"`

	// SecretsPath is a legacy top-level setting for backwards compatibility.
	// Users should prefer configuring dirs.secrets instead.
	SecretsPath string `json:"secrets_path,omitempty" env:"SEMAPHORE_SECRETS_PATH"`

	// HomeDirMode controls how the HOME environment variable is set for tasks.
	//   "template_home" (default) — HOME is set to a per-template directory,
	//       isolating .ansible/ across parallel tasks. Repo is cloned into a
	//       "src" subdirectory under HOME.
	//   "project_home" — HOME is set to the project temp directory (legacy
	//       behavior). Parallel ansible-galaxy runs in the same project may conflict.
	//   "user_home" — HOME is not overridden (keeps the real user HOME).
	//       ANSIBLE_HOME is set per template to isolate .ansible/ for Ansible tasks.
	HomeDirMode string `json:"home_dir_mode,omitempty" rule:"^(user_home|project_home|template_dir)?$" env:"SEMAPHORE_HOME_DIR_MODE" default:"template_dir"`

	// SshConfigPath is a path to the custom SSH config file.
	// Default path is ~/.ssh/config.
	SshConfigPath string `json:"ssh_config_path,omitempty" env:"SEMAPHORE_SSH_PATH"`

	Ssh *SshConfig `json:"ssh"`

	GitClientId string `json:"git_client,omitempty" rule:"^go_git|cmd_git$" env:"SEMAPHORE_GIT_CLIENT" default:"cmd_git"`

	// GitAttempts is how many times a git clone or pull is tried before the task
	// fails, for git servers which are intermittently unavailable. 1 tries once
	// and does not retry.
	//
	// Attempts rather than retries because a config value of 0 is
	// indistinguishable from an unset one and would be replaced by the default,
	// leaving no way to turn retrying off.
	GitAttempts int `json:"git_attempts,omitempty" env:"SEMAPHORE_GIT_ATTEMPTS" default:"4"`

	// web host
	WebHost string `json:"web_host,omitempty" env:"SEMAPHORE_WEB_ROOT"`

	// cookie hashing & encryption
	CookieHash       string `json:"cookie_hash,omitempty" env:"SEMAPHORE_COOKIE_HASH,sensitive"`
	CookieEncryption string `json:"cookie_encryption,omitempty" env:"SEMAPHORE_COOKIE_ENCRYPTION,sensitive"`
	// AccessKeyEncryption is BASE64 encoded byte array used
	// for encrypting and decrypting access keys stored in database.
	// Legacy entry point kept for backward compatibility; the access keyring is
	// configured via EncryptionKeys.AccessKey (encryption_keys.access_key).
	AccessKeyEncryption string `json:"access_key_encryption,omitempty" env:"SEMAPHORE_ACCESS_KEY_ENCRYPTION,sensitive"`

	// OptionEncryption is a BASE64 encoded key used to encrypt/decrypt DB options
	// (the JWT signing key) with the old single-key scheme (no rotation). It is
	// the option-keyring counterpart of AccessKeyEncryption: when set the option
	// keyring uses this one key; rotation is configured instead via the keys file
	// (encryption.keys_file → option_key). When unset, options fall back to the
	// access keyring.
	OptionEncryption string `json:"option_encryption,omitempty" env:"SEMAPHORE_OPTION_ENCRYPTION,sensitive"`

	// email alerting
	EmailAlert         bool   `json:"email_alert,omitempty" env:"SEMAPHORE_EMAIL_ALERT"`
	EmailSender        string `json:"email_sender,omitempty" env:"SEMAPHORE_EMAIL_SENDER"`
	EmailHost          string `json:"email_host,omitempty" env:"SEMAPHORE_EMAIL_HOST"`
	EmailPort          string `json:"email_port,omitempty" rule:"^(|[0-9]{1,5})$" env:"SEMAPHORE_EMAIL_PORT"`
	EmailUsername      string `json:"email_username,omitempty" env:"SEMAPHORE_EMAIL_USERNAME"`
	EmailPassword      string `json:"email_password,omitempty" env:"SEMAPHORE_EMAIL_PASSWORD,sensitive"`
	EmailSecure        bool   `json:"email_secure,omitempty" env:"SEMAPHORE_EMAIL_SECURE"`
	EmailTls           bool   `json:"email_tls,omitempty" env:"SEMAPHORE_EMAIL_TLS"`
	EmailTlsMinVersion string `json:"email_tls_min_version,omitempty" default:"1.2" rule:"^(1\\.[0123])$" env:"SEMAPHORE_EMAIL_TLS_MIN_VERSION"`

	// ldap settings
	LdapEnable       bool          `json:"ldap_enable,omitempty" env:"SEMAPHORE_LDAP_ENABLE"`
	LdapBindDN       string        `json:"ldap_binddn,omitempty" env:"SEMAPHORE_LDAP_BIND_DN"`
	LdapBindPassword string        `json:"ldap_bindpassword,omitempty" env:"SEMAPHORE_LDAP_BIND_PASSWORD,sensitive"`
	LdapServer       string        `json:"ldap_server,omitempty" env:"SEMAPHORE_LDAP_SERVER"`
	LdapSearchDN     string        `json:"ldap_searchdn,omitempty" env:"SEMAPHORE_LDAP_SEARCH_DN"`
	LdapSearchFilter string        `json:"ldap_searchfilter,omitempty" env:"SEMAPHORE_LDAP_SEARCH_FILTER"`
	LdapMappings     *LdapMappings `json:"ldap_mappings,omitempty"`
	LdapNeedTLS      bool          `json:"ldap_needtls,omitempty" env:"SEMAPHORE_LDAP_NEEDTLS"`
	// LdapTLSSkipVerify disables verification of the LDAP server's TLS
	// certificate for the legacy flat ldap_* config. Defaults to false
	// (certificates are verified). See LdapProvider.TLSSkipVerify.
	LdapTLSSkipVerify bool `json:"ldap_tls_skip_verify,omitempty" env:"SEMAPHORE_LDAP_TLS_SKIP_VERIFY"`

	// LdapProviders configures multiple LDAP directories (like OidcProviders
	// for OIDC). The key is the provider ID shown in identity records; the
	// ID "ldap" is reserved for the legacy flat ldap_* config above.
	LdapProviders map[string]LdapProvider `json:"ldap_providers,omitempty" env:"SEMAPHORE_LDAP_PROVIDERS"`

	// Telegram, Slack, Rocket.Chat, Microsoft Teams, DingTalk, and Gotify alerting
	TelegramAlert       bool   `json:"telegram_alert,omitempty" env:"SEMAPHORE_TELEGRAM_ALERT"`
	TelegramChat        string `json:"telegram_chat,omitempty" env:"SEMAPHORE_TELEGRAM_CHAT"`
	TelegramToken       string `json:"telegram_token,omitempty" env:"SEMAPHORE_TELEGRAM_TOKEN,sensitive"`
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
	GotifyToken         string `json:"gotify_token,omitempty" env:"SEMAPHORE_GOTIFY_TOKEN,sensitive"`

	// oidc settings
	OidcProviders map[string]OidcProvider `json:"oidc_providers,omitempty" env:"SEMAPHORE_OIDC_PROVIDERS"`

	MaxTaskDurationSec  int `json:"max_task_duration_sec,omitempty" env:"SEMAPHORE_MAX_TASK_DURATION_SEC"`
	MaxTasksPerTemplate int `json:"max_tasks_per_template,omitempty" env:"SEMAPHORE_MAX_TASKS_PER_TEMPLATE"`

	// task concurrency
	MaxParallelTasks int `json:"max_parallel_tasks,omitempty" default:"9999" rule:"^[0-9]{1,10}$" env:"SEMAPHORE_MAX_PARALLEL_TASKS"`

	// RunnerRegistrationToken is deprecated, use Runners field instead of it.
	RunnerRegistrationToken string `json:"runner_registration_token,omitempty" env:"SEMAPHORE_RUNNER_REGISTRATION_TOKEN"`

	JWT *JWTConfig `json:"jwt,omitempty"`

	// feature switches
	PasswordLoginDisable bool `json:"password_login_disable,omitempty" env:"SEMAPHORE_PASSWORD_LOGIN_DISABLED"`
	// ExternalAuthEmailMatching controls whether an LDAP/OIDC login may be
	// linked to an existing user by email when no external identity record
	// exists yet:
	//   "auto" (default) - only external users without any linked identity
	//                      (one-time adoption of pre-2.20 accounts);
	//   "always"         - any external user (needed when the same person
	//                      logs in via several providers);
	//   "never"          - identities are matched strictly by provider ID.
	// Local (password) accounts are never matched regardless of the mode.
	ExternalAuthEmailMatching string `json:"external_auth_email_matching,omitempty" env:"SEMAPHORE_EXTERNAL_AUTH_EMAIL_MATCHING" rule:"^(auto|always|never)?$" default:"auto"`
	NonAdminCanCreateProject  bool   `json:"non_admin_can_create_project,omitempty" env:"SEMAPHORE_NON_ADMIN_CAN_CREATE_PROJECT"`

	// UseRemoteRunner is deprecated. Use Runners field instead of it.
	UseRemoteRunner bool `json:"use_remote_runner,omitempty" env:"SEMAPHORE_USE_REMOTE_RUNNER"`

	Apps map[string]App `json:"apps,omitempty" env:"SEMAPHORE_APPS"`

	EnvVars map[string]string `json:"env_vars,omitempty" env:"SEMAPHORE_ENV_VARS"`

	ForwardedEnvVars []string `json:"forwarded_env_vars,omitempty" env:"SEMAPHORE_FORWARDED_ENV_VARS"`

	Teams *TeamsConfig `json:"teams,omitempty"`

	Syslog *SyslogConfig `json:"syslog,omitempty"`

	Metrics *MetricsConfig `json:"metrics,omitempty"`

	Log *ConfigLog `json:"log,omitempty"`

	Process *ConfigProcess `json:"process,omitempty"`

	Schedule *ScheduleConfig `json:"schedule,omitempty"`

	Debugging *DebuggingConfig `json:"debugging,omitempty"`

	HA *HAConfig `json:"ha,omitempty"`

	Subscription *SubscriptionConfig `json:"subscription,omitempty"`

	Dirs *ConfigDirs `json:"dirs,omitempty"`

	Runner *RunnerConfig `json:"runner,omitempty"`

	Runners *RunnersConfig `json:"runners,omitempty"`

	// Encryption groups the keys-file path and reload-poll settings. The keyrings
	// live exclusively in Encryption.KeysFile (a separate file, JSON or YAML),
	// which is watched for changes — edits are applied without restarting the
	// server. When unset, the legacy flat AccessKeyEncryption field is used.
	Encryption *EncryptionConfig `json:"encryption,omitempty"`

	// keys holds the resolved runtime keyrings behind atomic pointers so they
	// can be hot-swapped during key rotation without restarting (see
	// ReloadEncryptionKeys). Unexported so it is ignored by JSON, env, defaults
	// and validation reflection.
	keys *keyringStore
}

// Default values for RunnersConfig, applied when the "runners" config section
// or its fields are absent.
const (
	defaultRunnersOfflineTimeoutSec    = 120
	defaultRunnersTaskFailTimeoutSec   = 420
	defaultRunnersReconcileIntervalSec = 30
)

// GetSecretsPath returns the secrets path from configuration.
// Used for backward compatibility with legacy top-level secrets_path.
func (conf *ConfigType) GetSecretsPath() string {
	if conf.Dirs.Secrets != "" && conf.Dirs.Secrets != "/tmp/semaphore" {
		return conf.Dirs.Secrets
	}
	if conf.SecretsPath != "" {
		return conf.SecretsPath
	}
	if conf.Dirs.Secrets != "" {
		return conf.Dirs.Secrets
	}
	return "/tmp/semaphore"
}

// GetSshConfigPath return SSH config path from configuration.
// Used for backward compatibility.
func (conf *ConfigType) GetSshConfigPath() string {
	if conf.Ssh.ConfigPath != "" {
		return conf.Ssh.ConfigPath
	}
	return conf.SshConfigPath
}

func (conf *ConfigType) GetRunnerRegistrationToken() string {
	if conf.Runners.RegistrationToken != "" {
		return conf.Runners.RegistrationToken
	}
	return conf.RunnerRegistrationToken
}

func (conf *ConfigType) IsUseRemoteRunner() bool {
	switch conf.Runners.DefaultGlobalRunnersMode {
	case DefultGlobalRunnerDisable:
		return false
	case DefultGlobalRunnerRequire:
		return true
	case DefultGlobalRunnerPrefer:
		return true
	default:
		return conf.UseRemoteRunner
	}
}

// RunnersOfflineTimeout returns the heartbeat staleness after which a runner
// is considered offline (no new tasks; its "starting" tasks are reassigned).
func (conf *ConfigType) RunnersOfflineTimeout() time.Duration {
	sec := defaultRunnersOfflineTimeoutSec
	if conf.Runners != nil && conf.Runners.OfflineTimeoutSec > 0 {
		sec = conf.Runners.OfflineTimeoutSec
	}
	return time.Duration(sec) * time.Second
}

// RunnersTaskFailTimeout returns the heartbeat staleness after which a
// runner's "running" tasks are failed. It is never below the offline timeout.
func (conf *ConfigType) RunnersTaskFailTimeout() time.Duration {
	sec := defaultRunnersTaskFailTimeoutSec
	if conf.Runners != nil && conf.Runners.TaskFailTimeoutSec > 0 {
		sec = conf.Runners.TaskFailTimeoutSec
	}
	res := time.Duration(sec) * time.Second
	if offline := conf.RunnersOfflineTimeout(); res < offline {
		res = offline
	}
	return res
}

// RunnersReconcileInterval returns how often the server reconciles dispatched
// tasks against runner liveness.
func (conf *ConfigType) RunnersReconcileInterval() time.Duration {
	sec := defaultRunnersReconcileIntervalSec
	if conf.Runners != nil && conf.Runners.ReconcileIntervalSec > 0 {
		sec = conf.Runners.ReconcileIntervalSec
	}
	return time.Duration(sec) * time.Second
}

type SubscriptionConfig struct {
	// Key is a subscription key or token that can be set via config.
	// When this is set, subscription activation from the web interface is disabled.
	Key       string `json:"key,omitempty" db:"-" env:"SEMAPHORE_SUBSCRIPTION_KEY,sensitive"`
	KeyFile   string `json:"key_file,omitempty" db:"-" env:"SEMAPHORE_SUBSCRIPTION_KEY_FILE"`
	ServerURL string `json:"server_url,omitempty" env:"SEMAPHORE_SUBSCRIPTION_SERVER_URL" default:"https://portal.semaphoreui.com/billing"`
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
	//fmt.Println("Loading config")

	Config = NewConfigType()
	Config.Apps = map[string]App{}

	if !noConfigFile {
		usedConfigPath = loadConfigFile(configPath)
	}

	loadConfigEnvironment()
	loadConfigDefaults()

	// Resolve encryption keyrings (read key files, apply precedence, build the
	// runtime keyrings) before validation consumes the keys.
	resolveEncryptionKeys()

	//fmt.Println("Validating config")
	validateConfig()

	if Config.Process.NoNewPrivs {
		if err := SetNoNewPrivs(); err != nil {
			panic(fmt.Errorf("failed to set no_new_privs: %w", err))
		}
	}

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

	if Config.Runner != nil && Config.Runner.Token != "" && Config.Runner.TokenFile != "" {
		panic("SEMAPHORE_RUNNER_TOKEN and SEMAPHORE_RUNNER_TOKEN_FILE are mutually exclusive")
	}

	if Config.Runner != nil && Config.Runner.TokenFile != "" {
		runnerTokenBytes, err := os.ReadFile(Config.Runner.TokenFile)
		if err == nil {
			Config.Runner.Token = strings.TrimSpace(string(runnerTokenBytes))
		}
	}

	if Config.Subscription.KeyFile != "" {
		subscriptionKeyBytes, err := os.ReadFile(Config.Subscription.KeyFile)
		if err != nil {
			panic(err)
		}

		Config.Subscription.Key = strings.TrimSpace(string(subscriptionKeyBytes))
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
			path.Join(cwd, "config.yaml"),
			path.Join(cwd, "config.yml"),
			"/usr/local/etc/semaphore/config.json",
			"/usr/local/etc/semaphore/config.yaml",
			"/usr/local/etc/semaphore/config.yml",
			"/etc/semaphore/config.json",
			"/etc/semaphore/config.yaml",
			"/etc/semaphore/config.yml",
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
			decodeConfig(file, p)
			usedConfigPath = &p
			break
		}
		exitOnConfigFileError(err)
	} else {
		p := configPath
		file, err := os.Open(p)
		exitOnConfigFileError(err)
		usedConfigPath = &p
		decodeConfig(file, p)
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

		if !fieldInfo.IsExported() {
			continue
		}

		fieldKind := fieldInfo.Type.Kind()
		isPtrToStruct := fieldKind == reflect.Ptr && fieldInfo.Type.Elem().Kind() == reflect.Struct

		if !fieldValue.IsZero() && fieldKind != reflect.Struct && fieldKind != reflect.Map && !isPtrToStruct {
			continue
		}

		if fieldKind == reflect.Struct {
			err := loadDefaultsToObject(fieldValue.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		} else if isPtrToStruct {
			if fieldValue.IsNil() {
				continue
			}

			err := loadDefaultsToObject(fieldValue.Interface())
			if err != nil {
				return err
			}
			continue
		} else if fieldKind == reflect.Map {
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
	legacySecretsPath := Config.SecretsPath
	if Config.Dirs == nil {
		Config.Dirs = &ConfigDirs{}
	}
	err := loadDefaultsToObject(Config)
	if err != nil {
		panic(err)
	}

	if legacySecretsPath != "" && (Config.Dirs.Secrets == "/tmp/semaphore" || Config.Dirs.Secrets == "") {
		Config.Dirs.Secrets = legacySecretsPath
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

		// Skip fields with db:"-" tag
		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}

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
				case reflect.Slice:
					// Handle slice assignment
					fieldElemType := fieldValue.Type().Elem()
					var sourceSlice reflect.Value
					if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
						sourceSlice = val
					} else if val.Kind() == reflect.String {
						// Try to parse JSON array from string
						str := val.String()
						// First, try to unmarshal into []any
						var anyArr []any
						if err := json.Unmarshal([]byte(str), &anyArr); err == nil {
							sourceSlice = reflect.ValueOf(anyArr)
						} else if fieldElemType.Kind() == reflect.String {
							// Fallback: treat as single element string
							sourceSlice = reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf("")), 1, 1)
							sourceSlice.Index(0).SetString(str)
						} else {
							return fmt.Errorf("expected slice or json array string for field %s but got %T", field.Name, value)
						}
					} else {
						return fmt.Errorf("expected slice for field %s but got %T", field.Name, value)
					}

					// Build destination slice
					newSlice := reflect.MakeSlice(fieldValue.Type(), 0, sourceSlice.Len())
					for i := 0; i < sourceSlice.Len(); i++ {
						srcElemVal := sourceSlice.Index(i)
						// When source is []any, elements come as interface{}, unwrap reflect.Value
						if srcElemVal.Kind() == reflect.Interface && !srcElemVal.IsNil() {
							srcElemVal = reflect.ValueOf(srcElemVal.Interface())
						}

						var dstElem reflect.Value
						// Prepare destination element
						if fieldElemType.Kind() == reflect.Struct {
							dstElem = reflect.New(fieldElemType).Elem()
							if srcElemVal.Kind() == reflect.Map {
								// Expect map[string]any
								mIface, ok := srcElemVal.Interface().(map[string]any)
								if !ok {
									return fmt.Errorf("cannot assign element of type %T to slice element of type %s", srcElemVal.Interface(), fieldElemType)
								}
								if err := assignMapToStructRecursive(mIface, dstElem); err != nil {
									return err
								}
							} else if srcElemVal.Type().ConvertibleTo(fieldElemType) {
								dstElem = srcElemVal.Convert(fieldElemType)
							} else {
								return fmt.Errorf("cannot assign element of type %s to slice element of type %s", srcElemVal.Type(), fieldElemType)
							}
						} else {
							// Primitive or other kinds
							if srcElemVal.Type().ConvertibleTo(fieldElemType) {
								dstElem = srcElemVal.Convert(fieldElemType)
							} else {
								newVal, converted := CastValueToKind(srcElemVal.Interface(), fieldElemType.Kind())
								if !converted {
									return fmt.Errorf("cannot assign element of type %s to slice element of type %s", srcElemVal.Type(), fieldElemType)
								}
								dstElem = reflect.ValueOf(newVal)
							}
						}

						newSlice = reflect.Append(newSlice, dstElem)
					}

					fieldValue.Set(newSlice)
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
		// strings are always acceptable as-is, or will be coerced upstream
		ok = true
	case reflect.Int:
		if reflect.ValueOf(value).Kind() == reflect.Int {
			ok = true
		} else {
			res = castStringToInt(fmt.Sprintf("%v", reflect.ValueOf(value)))
			ok = true
		}
	case reflect.Bool:
		if reflect.ValueOf(value).Kind() == reflect.Bool {
			ok = true
		} else {
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
			convertedValue := reflect.ValueOf(newValue)
			if convertedValue.Type().AssignableTo(attribute.Type()) {
				attribute.Set(convertedValue)
			} else if convertedValue.Type().ConvertibleTo(attribute.Type()) {
				attribute.Set(convertedValue.Convert(attribute.Type()))
			} else {
				panic(fmt.Errorf("cannot assign value of type %s to field of type %s", convertedValue.Type(), attribute.Type()))
			}
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

// resolveKeySource returns the key material from a KeySource: the inline Value,
// or the trimmed contents of File. Value and File are mutually exclusive.
func resolveKeySource(ks KeySource, name string) (string, error) {
	if ks.Value != "" && ks.File != "" {
		return "", fmt.Errorf("%s: 'value' and 'file' are mutually exclusive", name)
	}
	if ks.File != "" {
		data, err := os.ReadFile(ks.File)
		if err != nil {
			return "", fmt.Errorf("%s: read key file %q: %w", name, ks.File, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return ks.Value, nil
}

// resolveEncryptionKeysFrom builds the runtime keyset from the keys-file config
// plus the legacy flat fields, validating every resolved key. It does not mutate
// global state. The flat fields are added to the registry (so new writes can stamp
// them) and recorded as the legacy no-prefix decrypt keys.
func resolveEncryptionKeysFrom(enc *EncryptionKeysConfig, flatAccess, flatOption string) (*keyset, error) {
	ks := &keyset{
		byID:         map[string]string{},
		legacyAccess: flatAccess,
		legacyOption: flatOption,
	}

	// byLabel maps a human label (inline keys map key, or a folder filename) to its
	// material, for resolving the active pointers.
	byLabel := map[string]string{}
	addLabeled := func(label, material string) error {
		if material == "" {
			return nil
		}
		if err := validateAccessKeyEncryption(material); err != nil {
			return err
		}
		ks.byID[keyID(material)] = material
		byLabel[label] = material
		return nil
	}

	if enc != nil {
		for label, src := range enc.Keys {
			material, err := resolveKeySource(src, "encryption_keys.keys."+label)
			if err != nil {
				return nil, err
			}
			if err := addLabeled(label, material); err != nil {
				return nil, err
			}
		}
		if enc.KeysFolder != "" {
			if err := loadKeysFolder(enc.KeysFolder, addLabeled); err != nil {
				return nil, err
			}
		}
	}

	// Flat fields are registry entries (so new writes can stamp them); they are
	// labelled "" so they never satisfy a named active pointer.
	if err := addLabeled("", flatAccess); err != nil {
		return nil, err
	}
	if err := addLabeled("", flatOption); err != nil {
		return nil, err
	}

	accessActive, err := resolveActiveKey(enc, flatAccess, byLabel, addLabeled,
		activePointer(enc, func(a ActivePointers) (string, string) { return a.SecretKey, a.SecretKeyFile }), "access")
	if err != nil {
		return nil, err
	}
	ks.accessID = keyID(accessActive) // "" => encryption disabled

	optionActive, err := resolveActiveKey(enc, flatOption, byLabel, addLabeled,
		activePointer(enc, func(a ActivePointers) (string, string) { return a.OptionKey, a.OptionKeyFile }), "option")
	if err != nil {
		return nil, err
	}
	ks.optionID = keyID(optionActive) // "" => option falls back to access

	return ks, nil
}

func activePointer(enc *EncryptionKeysConfig, pick func(ActivePointers) (string, string)) [2]string {
	if enc == nil {
		return [2]string{}
	}
	l, f := pick(enc.Active)
	return [2]string{l, f}
}

// resolveActiveKey resolves the active key material for one purpose: an active
// label wins, then an active filename (in KeysFolder, relative), then the flat
// fallback. A filename not already loaded from the folder is read and registered.
func resolveActiveKey(enc *EncryptionKeysConfig, flat string, byLabel map[string]string,
	addLabeled func(string, string) error, ptr [2]string, kind string) (string, error) {

	label, file := ptr[0], ptr[1]

	if label != "" {
		material, ok := byLabel[label]
		if !ok {
			return "", fmt.Errorf("encryption_keys.active.%s_key: no key labelled %q", kind, label)
		}
		return material, nil
	}

	if file != "" {
		if material, ok := byLabel[file]; ok {
			return material, nil
		}
		path := file
		if !filepath.IsAbs(path) && enc != nil {
			path = filepath.Join(enc.KeysFolder, file)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("encryption_keys.active.%s_key_file: %w", kind, err)
		}
		material := strings.TrimSpace(string(data))
		if err := addLabeled(file, material); err != nil {
			return "", err
		}
		return material, nil
	}

	return flat, nil
}

// loadKeysFolder reads every regular file in folder as one key, labelled by its
// filename. Dot-prefixed entries (e.g. Kubernetes' "..data" / "..2024_*") are
// skipped; symlinks (how K8s mounts secret files) are followed via Stat.
func loadKeysFolder(folder string, addLabeled func(string, string) error) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("encryption_keys.keys_folder %q: %w", folder, err)
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		path := filepath.Join(folder, name)
		info, err := os.Stat(path) // follow symlink
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("encryption_keys.keys_folder: read %q: %w", name, err)
		}
		if err := addLabeled(name, strings.TrimSpace(string(data))); err != nil {
			return fmt.Errorf("encryption_keys.keys_folder: key %q: %w", name, err)
		}
	}
	return nil
}

// EncryptionKeysFile returns the configured keys-file path (encryption.keys_file),
// or "" when no encryption section is configured.
func (conf *ConfigType) EncryptionKeysFile() string {
	if conf.Encryption == nil {
		return ""
	}
	return conf.Encryption.KeysFile
}

// EncryptionKeysPollInterval returns how often the keys file is polled for
// changes. It defaults to 15s, and returns 0 when polling is disabled
// (encryption.keys_poll_interval set to "0"). An unparseable value falls back to
// the default.
func (conf *ConfigType) EncryptionKeysPollInterval() time.Duration {
	const def = 15 * time.Second
	if conf.Encryption == nil || conf.Encryption.KeysPollInterval == "" {
		return def
	}
	d, err := time.ParseDuration(conf.Encryption.KeysPollInterval)
	if err != nil {
		return def
	}
	return d
}

// loadEncryptionKeysSource returns the EncryptionKeysConfig to resolve from.
// The keyrings live exclusively in encryption.keys_file, re-read from disk so its
// edits (and edits to the key files it references) are picked up on reload.
// When unset there is no structured config and the legacy flat
// AccessKeyEncryption field is used instead.
func loadEncryptionKeysSource() (*EncryptionKeysConfig, error) {
	path := Config.EncryptionKeysFile()
	if path == "" {
		return nil, nil
	}
	return readEncryptionKeysConfigFile(path)
}

// resolveEncryptionKeys builds the runtime keyrings once at startup and stores
// them on Config. Invalid keys panic (fail fast at boot).
func resolveEncryptionKeys() {
	enc, err := loadEncryptionKeysSource()
	if err != nil {
		panic(err)
	}

	ks, err := resolveEncryptionKeysFrom(enc, Config.AccessKeyEncryption, Config.OptionEncryption)
	if err != nil {
		panic(err)
	}
	if Config.keys == nil {
		Config.keys = &keyringStore{}
	}
	Config.keys.current.Store(ks)
}

// ReloadEncryptionKeys re-reads the encryption keys (the dedicated
// EncryptionKeysFile or the encryption_keys section of the config file, plus any
// referenced key files) and atomically swaps the runtime keyrings, without
// restarting. It validates the new keys first and leaves the current keyrings
// untouched on any error. Safe to call concurrently with encryption/decryption.
func ReloadEncryptionKeys() error {
	_, err := reloadEncryptionKeys(true)
	return err
}

// ReloadEncryptionKeysIfChanged is like ReloadEncryptionKeys but performs the
// atomic swap only when the resolved keys actually differ from the active ones,
// returning whether a change was applied. It is the file watcher's entry point.
func ReloadEncryptionKeysIfChanged() (changed bool, err error) {
	return reloadEncryptionKeys(false)
}

func reloadEncryptionKeys(force bool) (bool, error) {
	enc, err := loadEncryptionKeysSource()
	if err != nil {
		return false, err
	}

	ks, err := resolveEncryptionKeysFrom(enc, Config.AccessKeyEncryption, Config.OptionEncryption)
	if err != nil {
		return false, err
	}

	if Config.keys == nil {
		Config.keys = &keyringStore{}
	}

	Config.keys.reloadMu.Lock()
	defer Config.keys.reloadMu.Unlock()

	if !force && keysetsEqual(ks, Config.keys.current.Load()) {
		return false, nil
	}

	Config.keys.current.Store(ks)
	return true, nil
}

// readEncryptionKeysConfigFile decodes a dedicated encryption-keys file (whose
// whole content is an EncryptionKeysConfig), as JSON or YAML.
func readEncryptionKeysConfigFile(path string) (*EncryptionKeysConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Parse via YAML regardless of extension: YAML 1.2 is a superset of JSON, so
	// this accepts both formats. The file is often a Kubernetes secret mounted at
	// a path with no .yaml/.yml extension, so extension-based detection is not
	// reliable here.
	var raw any
	if err := yaml.NewDecoder(file).Decode(&raw); err != nil {
		return nil, err
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var enc EncryptionKeysConfig
	if err := json.Unmarshal(data, &enc); err != nil {
		return nil, err
	}

	return &enc, nil
}

func validateAccessKeyEncryption(key string) error {
	if key == "" {
		return nil
	}

	encryption, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("access_key_encryption must be a valid base64 string: %w", err)
	}

	switch len(encryption) {
	case 16, 24, 32:
		return nil
	default:
		return fmt.Errorf(
			"access_key_encryption has invalid decoded length %d bytes; AES requires 16, 24, or 32 bytes (use `openssl rand -base64 32` to generate a valid key)",
			len(encryption),
		)
	}
}

func validateConfig() {
	err := validate(Config)
	if err != nil {
		panic(err)
	}

	if err := validateAccessKeyEncryption(Config.AccessKeyEncryption); err != nil {
		panic(err)
	}
	if err := validateAccessKeyEncryption(Config.OptionEncryption); err != nil {
		panic(err)
	}
	if Config.keys != nil {
		if ks := Config.keys.current.Load(); ks != nil {
			for _, material := range ks.byID {
				if err := validateAccessKeyEncryption(material); err != nil {
					panic(err)
				}
			}
		}
	}
}

// parseEnvTag splits an env tag value like "SEMAPHORE_DB_PASS,sensitive"
// into the environment variable name and whether it is sensitive.
func parseEnvTag(tag string) (envVar string, sensitive bool) {
	parts := strings.SplitN(tag, ",", 2)
	envVar = parts[0]
	if len(parts) > 1 && parts[1] == "sensitive" {
		sensitive = true
	}
	return
}

func loadEnvironmentToObject(obj any) (resultSensitiveEnvs []string, err error) {
	var currSensitiveEnvs []string

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
			currSensitiveEnvs, err = loadEnvironmentToObject(fieldValue.Addr().Interface())
			if err != nil {
				return
			}
			resultSensitiveEnvs = append(resultSensitiveEnvs, currSensitiveEnvs...)
			continue
		} else if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
			if fieldValue.IsZero() {
				newValue := reflect.New(fieldType.Type.Elem())
				fieldValue.Set(newValue)
			}

			envTag := fieldType.Tag.Get("env")
			if envTag != "" {
				envVar, sensitive := parseEnvTag(envTag)
				if envValue, exists := os.LookupEnv(envVar); exists {
					newValue := reflect.New(fieldType.Type.Elem())
					err = json.Unmarshal([]byte(envValue), newValue.Interface())
					if err != nil {
						return
					}
					fieldValue.Set(newValue)
					if sensitive {
						resultSensitiveEnvs = append(resultSensitiveEnvs, envVar)
					}
				}
			}

			currSensitiveEnvs, err = loadEnvironmentToObject(fieldValue.Interface())
			if err != nil {
				return
			}

			resultSensitiveEnvs = append(resultSensitiveEnvs, currSensitiveEnvs...)
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envVar, sensitive := parseEnvTag(envTag)

		envValue, exists := os.LookupEnv(envVar)

		if !exists {
			continue
		}

		setConfigValue(fieldValue, envValue) // envValue always string!!!

		if sensitive {
			resultSensitiveEnvs = append(resultSensitiveEnvs, envVar)
		}
	}

	slices.Sort(resultSensitiveEnvs)
	resultSensitiveEnvs = slices.Compact(resultSensitiveEnvs)
	return
}

func loadConfigEnvironment() {
	sensitiveEnvs, err := loadEnvironmentToObject(Config)
	if err != nil {
		panic(err)
	}

	for _, sensitiveEnv := range sensitiveEnvs {
		os.Unsetenv(sensitiveEnv)
	}
}

func exitOnConfigError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func exitOnConfigFileError(err error) {
	if err != nil {
		exitOnConfigError("Cannot Find configuration! Use --config parameter to point to a JSON or YAML file generated by `semaphore setup`.")
	}
}

func decodeConfig(file io.Reader, configPath string) {
	if isYAMLConfig(configPath) {
		decodeConfigYAML(file)
		return
	}
	var raw map[string]any
	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
	unmarshalConfigMap(raw)
}

// migrateLegacyConfigKeys rewrites deprecated top-level config keys so file-based
// settings survive struct renames (auth→mfa, flat subscription fields→subscription).
func migrateLegacyConfigKeys(m map[string]any) {
	if _, ok := m["mfa"]; !ok {
		if auth, ok := m["auth"]; ok {
			m["mfa"] = auth
		}
	}

	if _, ok := m["subscription"]; ok {
		return
	}

	sub := map[string]any{}
	var migrated bool

	if v, ok := m["subscription_key"]; ok {
		sub["key"] = v
		delete(m, "subscription_key")
		migrated = true
	}
	if v, ok := m["subscription_key_file"]; ok {
		sub["key_file"] = v
		delete(m, "subscription_key_file")
		migrated = true
	}
	if v, ok := m["subscription_server_url"]; ok {
		sub["server_url"] = v
		delete(m, "subscription_server_url")
		migrated = true
	}

	if migrated {
		m["subscription"] = sub
	}
}

func unmarshalConfigMap(raw map[string]any) {
	migrateLegacyConfigKeys(raw)
	data, err := json.Marshal(raw)
	if err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
	if err := json.Unmarshal(data, &Config); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
}

func isYAMLConfig(configPath string) bool {
	ext := strings.ToLower(filepath.Ext(configPath))
	return ext == ".yaml" || ext == ".yml"
}

func decodeConfigYAML(file io.Reader) {
	var raw any
	if err := yaml.NewDecoder(file).Decode(&raw); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
	data, err := json.Marshal(raw)
	if err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
	var configMap map[string]any
	if err := json.Unmarshal(data, &configMap); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
	unmarshalConfigMap(configMap)
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
		err = errors.New("BoltDB not supported")
		return
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
		fmt.Printf("BoltDB not supported\n")
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
		err = errors.New("BoltDB not supported")
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

	port = strings.TrimPrefix(port, ":")

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
