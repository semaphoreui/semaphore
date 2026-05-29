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
bolt:
  host: /tmp/db.bolt
`
	decodeConfig(strings.NewReader(yamlBody), "config.yaml")

	assert.Equal(t, ":1337", Config.Port)
	assert.Equal(t, "abc", Config.CookieHash)
	assert.Equal(t, 7, Config.MaxParallelTasks)
	require.NotNil(t, Config.BoltDb)
	assert.Equal(t, "/tmp/db.bolt", Config.BoltDb.Hostname)
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
