package util

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockError(msg string) {
	panic(msg)
}

func TestValidate(t *testing.T) {
	var val struct {
		Test string `rule:"^\\d+$"`
	}
	val.Test = "45243524"

	err := validate(val)
	if err != nil {
		t.Error(err)
	}
}

func TestLoadEnvironmentToObject(t *testing.T) {
	var val struct {
		Flag     bool   `env:"TEST_FLAG"`
		Test     string `env:"TEST_ENV_VAR"`
		Subfield struct {
			Value string `env:"TEST_VALUE_ENV_VAR"`
		}
		StringArr []string `env:"TEST_STRING_ARR"`
	}

	err := os.Setenv("TEST_FLAG", "yes")
	if err != nil {
		panic(err)
	}

	err = os.Setenv("TEST_ENV_VAR", "758478")
	if err != nil {
		panic(err)
	}

	err = os.Setenv("TEST_VALUE_ENV_VAR", "test_value")
	if err != nil {
		panic(err)
	}

	err = os.Setenv("TEST_STRING_ARR", "[\"test1\",\"test2\"]")
	if err != nil {
		panic(err)
	}

	_, err = loadEnvironmentToObject(&val)
	if err != nil {
		t.Error(err)
	}

	if val.Flag != true {
		t.Error("Invalid value")
	}

	if val.Test != "758478" {
		t.Error("Invalid value")
	}

	if val.Subfield.Value != "test_value" {
		t.Error("Invalid value")
	}

	if val.StringArr == nil {
		t.Error("Invalid array value")
	}

	if val.StringArr[0] != "test1" {
		t.Error("Invalid array item value")
	}

	if val.StringArr[1] != "test2" {
		t.Error("Invalid array item value")
	}
}

func TestLoadEnvironmentToObject_Arr(t *testing.T) {
	var val struct {
		StringArr []string `env:"TEST_STRING_ARR"`
	}

	err := os.Setenv("TEST_STRING_ARR", "[\"test1\",\"test2\"]")
	if err != nil {
		panic(err)
	}

	_, err = loadEnvironmentToObject(&val)
	if err != nil {
		t.Error(err)
	}

	if val.StringArr == nil {
		t.Error("Invalid array value")
	}

	if val.StringArr[0] != "test1" {
		t.Error("Invalid array item value")
	}

	if val.StringArr[1] != "test2" {
		t.Error("Invalid array item value")
	}
}

func TestLoadEnvironmentToObject_Map(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var val struct {
		Users map[string]User `env:"TEST_USERS"`
	}

	err := os.Setenv("TEST_USERS", "{\"test\":{\"name\":\"test\",\"age\":5}}")
	if err != nil {
		panic(err)
	}

	_, err = loadEnvironmentToObject(&val)
	if err != nil {
		panic(err)
	}

	if val.Users["test"].Name != "test" {
		t.Error("Invalid field value")
	}
}

func TestLoadEnvironmentToObject_RunnerExecutor(t *testing.T) {
	var val RunnerConfig

	require.NoError(t, os.Setenv("SEMAPHORE_RUNNER_EXECUTOR", `{"type":"docker","docker":{"image":"example.com/job:1"}}`))
	require.NoError(t, os.Setenv("SEMAPHORE_RUNNER_DOCKER_NETWORK", "host"))
	defer os.Unsetenv("SEMAPHORE_RUNNER_EXECUTOR")
	defer os.Unsetenv("SEMAPHORE_RUNNER_DOCKER_NETWORK")

	_, err := loadEnvironmentToObject(&val)
	require.NoError(t, err)

	require.NotNil(t, val.Executor)
	assert.Equal(t, ExecutorTypeDocker, val.Executor.Type)
	assert.Equal(t, "example.com/job:1", val.Executor.Docker.Image)
	// individual executor env vars still override fields of the JSON object
	assert.Equal(t, "host", val.Executor.Docker.Network)
}

func TestLoadEnvironmentToObject_SensitiveEnvs(t *testing.T) {
	type sub struct {
		Field string `json:"field"`
	}
	var val struct {
		Secret    *sub `env:"TEST_SECRET_SUB,sensitive"`
		NotSecret *sub `env:"TEST_PUBLIC_SUB"`
	}

	require.NoError(t, os.Setenv("TEST_SECRET_SUB", `{"field":"s"}`))
	require.NoError(t, os.Setenv("TEST_PUBLIC_SUB", `{"field":"p"}`))
	defer os.Unsetenv("TEST_SECRET_SUB")
	defer os.Unsetenv("TEST_PUBLIC_SUB")

	sensitive, err := loadEnvironmentToObject(&val)
	require.NoError(t, err)

	assert.Contains(t, sensitive, "TEST_SECRET_SUB")
	assert.NotContains(t, sensitive, "TEST_PUBLIC_SUB")
}

func TestLoadEnvironmentToObject_SensitiveEnvs_NoDuplicates(t *testing.T) {
	type sub struct {
		Field string `json:"field"`
	}
	var val struct {
		A *sub `env:"TEST_SHARED_SECRET,sensitive"`
		B *sub `env:"TEST_SHARED_SECRET,sensitive"`
	}

	require.NoError(t, os.Setenv("TEST_SHARED_SECRET", `{"field":"x"}`))
	defer os.Unsetenv("TEST_SHARED_SECRET")

	sensitive, err := loadEnvironmentToObject(&val)
	require.NoError(t, err)

	count := 0
	for _, env := range sensitive {
		if env == "TEST_SHARED_SECRET" {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate sensitive env names must be deduplicated")
	assert.Equal(t, &sub{Field: "x"}, val.A)
	assert.Equal(t, &sub{Field: "x"}, val.B)
}

func TestLoadEnvironmentToObject_SensitiveEnvs_Empty(t *testing.T) {
	var val struct {
		Plain string `env:"TEST_PLAIN_VAR"`
	}

	require.NoError(t, os.Setenv("TEST_PLAIN_VAR", "value"))
	defer os.Unsetenv("TEST_PLAIN_VAR")

	sensitive, err := loadEnvironmentToObject(&val)
	require.NoError(t, err)
	assert.Empty(t, sensitive)
}

func TestCastStringToInt(t *testing.T) {
	errMsg := "Cast string => int failed"

	if castStringToInt("5") != 5 {
		t.Error(errMsg)
	}
	if castStringToInt("0") != 0 {
		t.Error(errMsg)
	}
	if castStringToInt("-1") != -1 {
		t.Error(errMsg)
	}
	if castStringToInt("999") != 999 {
		t.Error(errMsg)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Cast string => int did not panic on invalid input")
		}
	}()
	castStringToInt("xxx")
}

func TestCastStringToBool(t *testing.T) {
	errMsg := "Cast string => bool failed"

	if castStringToBool("1") != true {
		t.Error(errMsg)
	}
	if castStringToBool("0") != false {
		t.Error(errMsg)
	}
	if castStringToBool("true") != true {
		t.Error(errMsg)
	}
	if castStringToBool("false") != false {
		t.Error(errMsg)
	}
	if castStringToBool("xxx") != false {
		t.Error(errMsg)
	}
	if castStringToBool("") != false {
		t.Error(errMsg)
	}
}

func TestConfigInitialization(t *testing.T) {
	testLdapMappingsUID := "uid"

	Config = NewConfigType()

	// should not panic
	Config.LdapMappings.UID = testLdapMappingsUID
}

func TestGetConfigValue(t *testing.T) {
	Config = NewConfigType()

	testPort := "1337"
	testCookieHash := "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	testMaxParallelTasks := 5
	testLdapNeedTls := true
	testDbHost := "192.168.0.1"

	Config.Port = testPort
	Config.CookieHash = testCookieHash
	Config.MaxParallelTasks = testMaxParallelTasks
	Config.LdapNeedTLS = testLdapNeedTls
	Config.SQLite = &DbConfig{
		Hostname: testDbHost,
	}

	if getConfigValue("Port") != testPort {
		t.Error("Could not get value for config attribute 'Port'!")
	}
	if getConfigValue("CookieHash") != testCookieHash {
		t.Error("Could not get value for config attribute 'CookieHash'!")
	}
	if getConfigValue("MaxParallelTasks") != fmt.Sprintf("%v", testMaxParallelTasks) {
		t.Error("Could not get value for config attribute 'MaxParallelTasks'!")
	}
	if getConfigValue("LdapNeedTLS") != fmt.Sprintf("%v", testLdapNeedTls) {
		t.Error("Could not get value for config attribute 'LdapNeedTLS'!")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not fail on non-existent config attribute!")
		}
	}()
	getConfigValue("NotExistent")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not fail on non-existent config attribute!")
		}
	}()
	getConfigValue("Not.Existent")
}

func TestSetConfigValue(t *testing.T) {
	Config = new(ConfigType)

	configValue := reflect.ValueOf(Config).Elem()

	testPort := "1337"
	testCookieHash := "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	testMaxParallelTasks := 5
	testLdapNeedTls := true
	// var testDbHost string = "192.168.0.1"
	testEmailSecure := "1"
	expectEmailSecure := true

	setConfigValue(configValue.FieldByName("Port"), testPort)
	setConfigValue(configValue.FieldByName("CookieHash"), testCookieHash)
	setConfigValue(configValue.FieldByName("MaxParallelTasks"), strconv.Itoa(testMaxParallelTasks))
	setConfigValue(configValue.FieldByName("LdapNeedTLS"), "true")
	// setConfigValue(configValue.FieldByName("BoltDb.Hostname"), testDbHost)
	setConfigValue(configValue.FieldByName("EmailSecure"), testEmailSecure)

	if Config.Port != testPort {
		t.Error("Could not set value for config attribute 'Port'!")
	}
	if Config.CookieHash != testCookieHash {
		t.Error("Could not set value for config attribute 'CookieHash'!")
	}
	if Config.MaxParallelTasks != testMaxParallelTasks {
		t.Error("Could not set value for config attribute 'MaxParallelTasks'!")
	}
	if Config.LdapNeedTLS != testLdapNeedTls {
		t.Error("Could not set value for config attribute 'LdapNeedTls'!")
	}
	//if Config.BoltDb.Hostname != testDbHost {
	//	t.Error("Could not set value for config attribute 'BoltDb.Hostname'!")
	//}
	if Config.EmailSecure != expectEmailSecure {
		t.Error("Could not set value for config attribute 'EmailSecure'!")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not fail on non-existent config attribute!")
		}
	}()
	setConfigValue(configValue.FieldByName("NotExistent"), "someValue")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not fail on non-existent config attribute!")
		}
	}()
	// setConfigValue(configValue.FieldByName("Not.Existent"), "someValue")

}

func TestLoadConfigEnvironmet(t *testing.T) {
	Config = new(ConfigType)
	Config.SQLite = &DbConfig{}
	Config.Dialect = DbDriverSQLite

	envPort := "1337"
	envCookieHash := "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	envAccessKeyEncryption := "1/wRYXQltDGwbzNZRP9ZfJb2IoWcn1hYrxA0vOdvVos="
	envMaxParallelTasks := "5"
	expectMaxParallelTasks := 5
	expectLdapNeedTls := true
	envLdapNeedTls := "1"
	envDbHost := "192.168.0.1"

	os.Setenv("SEMAPHORE_PORT", envPort)                                 //nolint:errcheck
	os.Setenv("SEMAPHORE_COOKIE_HASH", envCookieHash)                    //nolint:errcheck
	os.Setenv("SEMAPHORE_ACCESS_KEY_ENCRYPTION", envAccessKeyEncryption) //nolint:errcheck
	os.Setenv("SEMAPHORE_MAX_PARALLEL_TASKS", envMaxParallelTasks)       //nolint:errcheck
	os.Setenv("SEMAPHORE_LDAP_NEEDTLS", envLdapNeedTls)                  //nolint:errcheck
	os.Setenv("SEMAPHORE_DB_HOST", envDbHost)                            //nolint:errcheck

	loadConfigEnvironment()

	if Config.Port != envPort {
		t.Error("Setting 'Port' was not loaded from environment-vars!")
	}
	if Config.CookieHash != envCookieHash {
		t.Error("Setting 'CookieHash' was not loaded from environment-vars!")
	}
	if Config.AccessKeyEncryption != envAccessKeyEncryption {
		t.Error("Setting 'AccessKeyEncryption' was not loaded from environment-vars!")
	}
	if Config.MaxParallelTasks != expectMaxParallelTasks {
		t.Error("Setting 'MaxParallelTasks' was not loaded from environment-vars!")
	}
	if Config.LdapNeedTLS != expectLdapNeedTls {
		t.Error("Setting 'LdapNeedTLS' was not loaded from environment-vars!")
	}
	if Config.SQLite.Hostname != envDbHost {
		t.Error("Setting 'SQLite.Hostname' was not loaded from environment-vars!")
	}

	//if Config.MySQL.Hostname == envDbHost || Config.Postgres.Hostname == envDbHost {
	//	// inactive db-dialects could be set as they share the same env-vars; but should be ignored
	//	t.Error("DB-Hostname was loaded for inactive DB-dialects!")
	//}
}

func TestIsYAMLConfig(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"config.yaml", true},
		{"config.yml", true},
		{"/etc/semaphore/config.YAML", true},
		{"config.json", false},
		{"config", false},
		{"", false},
		{"config.yaml.bak", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, isYAMLConfig(tt.path))
		})
	}
}

func TestDecodeConfig_JSON(t *testing.T) {
	Config = new(ConfigType)

	jsonBody := `{"port":":1337","cookie_hash":"abc","max_parallel_tasks":7}`
	decodeConfig(strings.NewReader(jsonBody), "config.json")

	assert.Equal(t, ":1337", Config.Port)
	assert.Equal(t, "abc", Config.CookieHash)
	assert.Equal(t, 7, Config.MaxParallelTasks)
}

func TestDecodeConfig_YAML(t *testing.T) {
	Config = new(ConfigType)

	yamlBody := `
port: ":1337"
cookie_hash: abc
max_parallel_tasks: 7
sqlite:
  host: /tmp/db.bolt
`
	decodeConfig(strings.NewReader(yamlBody), "config.yaml")

	assert.Equal(t, ":1337", Config.Port)
	assert.Equal(t, "abc", Config.CookieHash)
	assert.Equal(t, 7, Config.MaxParallelTasks)
	require.NotNil(t, Config.SQLite)
	assert.Equal(t, "/tmp/db.bolt", Config.SQLite.Hostname)
}

func TestDecodeConfig_YAML_YmlExtension(t *testing.T) {
	Config = new(ConfigType)

	yamlBody := "port: \":4242\"\n"
	decodeConfig(strings.NewReader(yamlBody), "config.yml")

	assert.Equal(t, ":4242", Config.Port)
}

func TestLoadConfigFile_YAML(t *testing.T) {
	Config = new(ConfigType)

	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(p, []byte("port: \":9999\"\ncookie_hash: yaml-hash\n"), 0600))

	used := loadConfigFile(p)
	require.NotNil(t, used)
	assert.Equal(t, p, *used)
	assert.Equal(t, ":9999", Config.Port)
	assert.Equal(t, "yaml-hash", Config.CookieHash)
}

func TestLoadConfigFile_JSON(t *testing.T) {
	Config = new(ConfigType)

	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(p, []byte(`{"port":":8888"}`), 0600))

	used := loadConfigFile(p)
	require.NotNil(t, used)
	assert.Equal(t, p, *used)
	assert.Equal(t, ":8888", Config.Port)
}

func TestLoadConfigDefaults(t *testing.T) {
	Config = new(ConfigType)
	errMsg := "Failed to load config-default"

	loadConfigDefaults()

	if Config.Port != ":3000" {
		t.Error(errMsg)
	}
	if Config.TmpPath != "/tmp/semaphore" {
		t.Error(errMsg)
	}
}

func ensureConfigValidationFailure(t *testing.T, attribute string, value any) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(
				"Config validation for attribute '%v' did not fail! (value '%v')",
				attribute, value,
			)
		}
	}()
	validateConfig()
}

func TestValidateConfig(t *testing.T) {
	// assert := assert.New(t)

	Config = new(ConfigType)

	testPort := ":3000"
	testDbDialect := DbDriverSQLite
	testCookieHash := "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	testMaxParallelTasks := 0
	testEmailTlsMinVersion := "1.2"

	Config.Port = testPort
	Config.Dialect = testDbDialect
	Config.CookieHash = testCookieHash
	Config.MaxParallelTasks = testMaxParallelTasks
	Config.GitClientId = GoGitClientId
	Config.CookieEncryption = testCookieHash
	Config.AccessKeyEncryption = testCookieHash
	Config.EmailTlsMinVersion = testEmailTlsMinVersion
	validateConfig()

	Config.Port = "INVALID"
	ensureConfigValidationFailure(t, "Port", Config.Port)

	Config.Port = ":100000"
	ensureConfigValidationFailure(t, "Port", Config.Port)
	Config.Port = testPort

	Config.MaxParallelTasks = -1
	ensureConfigValidationFailure(t, "MaxParallelTasks", Config.MaxParallelTasks)

	ensureConfigValidationFailure(t, "MaxParallelTasks", Config.MaxParallelTasks)
	Config.MaxParallelTasks = testMaxParallelTasks

	// Config.CookieHash = "\"0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ=\"" // invalid with quotes (can happen when supplied as env-var)
	// ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	// Config.CookieHash = "!)394340"
	// ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	// Config.CookieHash = ""
	// ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	// Config.CookieHash = "TQwjDZ5fIQtaIw==" // valid b64, but too small
	// ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)
	Config.CookieHash = testCookieHash

	Config.Dialect = "someOtherDB"
	ensureConfigValidationFailure(t, "Dialect", Config.Dialect)
	Config.Dialect = testDbDialect

	// AccessKeyEncryption: empty is allowed (no encryption)
	Config.AccessKeyEncryption = ""
	validateConfig()

	// AccessKeyEncryption: valid 32-byte key
	Config.AccessKeyEncryption = testCookieHash
	validateConfig()

	// AccessKeyEncryption: invalid base64
	Config.AccessKeyEncryption = "not-valid-base64!!!"
	ensureConfigValidationFailure(t, "AccessKeyEncryption", Config.AccessKeyEncryption)

	// AccessKeyEncryption: valid base64 but wrong size (48 bytes)
	Config.AccessKeyEncryption = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	ensureConfigValidationFailure(t, "AccessKeyEncryption", Config.AccessKeyEncryption)
	Config.AccessKeyEncryption = testCookieHash
}


func setTestEnv(t *testing.T, key, val string) {
	orig, existed := os.LookupEnv(key)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, orig)
		} else {
			_ = os.Unsetenv(key)
		}
	})
	_ = os.Setenv(key, val)
}

func unsetTestEnv(t *testing.T, key string) {
	orig, existed := os.LookupEnv(key)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, orig)
		} else {
			_ = os.Unsetenv(key)
		}
	})
	_ = os.Unsetenv(key)
}

func TestLoadEnvironmentToObject_TLS_HTTPRedirectPort(t *testing.T) {
	Config = NewConfigType()
	setTestEnv(t, "SEMAPHORE_TLS_ENABLED", "true")
	setTestEnv(t, "SEMAPHORE_TLS_CERT_FILE", "/path/to/cert.pem")
	setTestEnv(t, "SEMAPHORE_TLS_KEY_FILE", "/path/to/key.pem")
	setTestEnv(t, "SEMAPHORE_TLS_HTTP_REDIRECT_PORT", "8080")
	unsetTestEnv(t, "SEMAPHORE_TLS_HTTP_REDIRECT_ADDR")

	loadConfigEnvironment()

	require.NotNil(t, Config.TLS)
	assert.True(t, Config.TLS.Enabled)
	assert.Equal(t, "/path/to/cert.pem", Config.TLS.CertFile)
	assert.Equal(t, "/path/to/key.pem", Config.TLS.KeyFile)
	require.NotNil(t, Config.TLS.HTTPRedirectPort)
	assert.Equal(t, 8080, *Config.TLS.HTTPRedirectPort)
	assert.Empty(t, Config.TLS.HTTPRedirectAddr)
}

func TestLoadEnvironmentToObject_TLS_HTTPRedirectAddr(t *testing.T) {
	Config = NewConfigType()
	setTestEnv(t, "SEMAPHORE_TLS_ENABLED", "true")
	setTestEnv(t, "SEMAPHORE_TLS_HTTP_REDIRECT_ADDR", "0.0.0.0:8080")
	unsetTestEnv(t, "SEMAPHORE_TLS_HTTP_REDIRECT_PORT")

	loadConfigEnvironment()

	require.NotNil(t, Config.TLS)
	assert.True(t, Config.TLS.Enabled)
	assert.Equal(t, "0.0.0.0:8080", Config.TLS.HTTPRedirectAddr)
	assert.Nil(t, Config.TLS.HTTPRedirectPort)
}

func TestLoadEnvironmentToObject_PrimitivePointers(t *testing.T) {
	var val struct {
		IntPtr    *int            `env:"TEST_INT_PTR"`
		StringPtr *string         `env:"TEST_STR_PTR"`
		BoolPtr   *bool           `env:"TEST_BOOL_PTR"`
		SliceInt  *[]int          `env:"TEST_SLICE_INT"`
		SliceStr  *[]string       `env:"TEST_SLICE_STR"`
		MapPtr    *map[string]int `env:"TEST_MAP_PTR"`
	}

	setTestEnv(t, "TEST_INT_PTR", "42")
	setTestEnv(t, "TEST_STR_PTR", "hello_world")
	setTestEnv(t, "TEST_BOOL_PTR", "true")
	setTestEnv(t, "TEST_SLICE_INT", "[10,20,30]")
	setTestEnv(t, "TEST_SLICE_STR", `["foo","bar"]`)
	setTestEnv(t, "TEST_MAP_PTR", `{"a":1,"b":2}`)

	_, err := loadEnvironmentToObject(&val)
	require.NoError(t, err)

	require.NotNil(t, val.IntPtr)
	assert.Equal(t, 42, *val.IntPtr)

	require.NotNil(t, val.StringPtr)
	assert.Equal(t, "hello_world", *val.StringPtr)

	require.NotNil(t, val.BoolPtr)
	assert.True(t, *val.BoolPtr)

	require.NotNil(t, val.SliceInt)
	assert.Equal(t, []int{10, 20, 30}, *val.SliceInt)

	require.NotNil(t, val.SliceStr)
	assert.Equal(t, []string{"foo", "bar"}, *val.SliceStr)

	require.NotNil(t, val.MapPtr)
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, *val.MapPtr)
}

func TestLoadEnvironmentToObject_ConfigProcess_UID_GID(t *testing.T) {
	Config = NewConfigType()
	setTestEnv(t, "SEMAPHORE_PROCESS_USER", "semaphore")
	setTestEnv(t, "SEMAPHORE_PROCESS_UID", "1001")
	setTestEnv(t, "SEMAPHORE_PROCESS_GID", "1002")

	loadConfigEnvironment()

	assert.Equal(t, "semaphore", Config.Process.User)
	require.NotNil(t, Config.Process.UID)
	assert.Equal(t, uint32(1001), *Config.Process.UID)
	require.NotNil(t, Config.Process.GID)
	assert.Equal(t, uint32(1002), *Config.Process.GID)
}

func TestCastValueToKind_IntegerBoundaries(t *testing.T) {
	// Signed boundaries
	v, ok := CastValueToKind("-128", reflect.Int8)
	assert.True(t, ok)
	assert.Equal(t, int8(-128), v)

	v, ok = CastValueToKind("127", reflect.Int8)
	assert.True(t, ok)
	assert.Equal(t, int8(127), v)

	assert.Panics(t, func() { CastValueToKind("128", reflect.Int8) })
	assert.Panics(t, func() { CastValueToKind("-129", reflect.Int8) })

	v, ok = CastValueToKind("-32768", reflect.Int16)
	assert.True(t, ok)
	assert.Equal(t, int16(-32768), v)

	v, ok = CastValueToKind("32767", reflect.Int16)
	assert.True(t, ok)
	assert.Equal(t, int16(32767), v)

	assert.Panics(t, func() { CastValueToKind("32768", reflect.Int16) })

	v, ok = CastValueToKind("-2147483648", reflect.Int32)
	assert.True(t, ok)
	assert.Equal(t, int32(-2147483648), v)

	v, ok = CastValueToKind("2147483647", reflect.Int32)
	assert.True(t, ok)
	assert.Equal(t, int32(2147483647), v)

	assert.Panics(t, func() { CastValueToKind("2147483648", reflect.Int32) })

	v, ok = CastValueToKind("9223372036854775807", reflect.Int64)
	assert.True(t, ok)
	assert.Equal(t, int64(9223372036854775807), v)

	assert.Panics(t, func() { CastValueToKind("9223372036854775808", reflect.Int64) })

	// Unsigned boundaries
	v, ok = CastValueToKind("0", reflect.Uint8)
	assert.True(t, ok)
	assert.Equal(t, uint8(0), v)

	v, ok = CastValueToKind("255", reflect.Uint8)
	assert.True(t, ok)
	assert.Equal(t, uint8(255), v)

	assert.Panics(t, func() { CastValueToKind("256", reflect.Uint8) })
	assert.Panics(t, func() { CastValueToKind("-1", reflect.Uint8) })

	v, ok = CastValueToKind("65535", reflect.Uint16)
	assert.True(t, ok)
	assert.Equal(t, uint16(65535), v)

	assert.Panics(t, func() { CastValueToKind("65536", reflect.Uint16) })

	v, ok = CastValueToKind("4294967295", reflect.Uint32)
	assert.True(t, ok)
	assert.Equal(t, uint32(4294967295), v)

	assert.Panics(t, func() { CastValueToKind("4294967296", reflect.Uint32) })

	v, ok = CastValueToKind("18446744073709551615", reflect.Uint64)
	assert.True(t, ok)
	assert.Equal(t, uint64(18446744073709551615), v)

	assert.Panics(t, func() { CastValueToKind("18446744073709551616", reflect.Uint64) })
}
