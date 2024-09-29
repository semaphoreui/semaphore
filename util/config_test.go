package util

import (
	"fmt"
	"os"
	"reflect"
	"testing"
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

	err = loadEnvironmentToObject(&val)
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
}

func TestCastStringToInt(t *testing.T) {

	var errMsg = "Cast string => int failed"

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

	var errMsg = "Cast string => bool failed"

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

func TestGetConfigValue(t *testing.T) {

	Config = new(ConfigType)

	var testPort = "1337"
	var testCookieHash = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var testMaxParallelTasks = 5
	var testLdapNeedTls = true
	var testDbHost = "192.168.0.1"

	Config.Port = testPort
	Config.CookieHash = testCookieHash
	Config.MaxParallelTasks = testMaxParallelTasks
	Config.LdapNeedTLS = testLdapNeedTls
	Config.BoltDb = &DbConfig{
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

	if getConfigValue("BoltDb.Hostname") != fmt.Sprintf("%v", testDbHost) {
		t.Error("Could not get value for config attribute 'BoltDb.Hostname'!")
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

	var testPort = "1337"
	var testCookieHash = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var testMaxParallelTasks = 5
	var testLdapNeedTls = true
	//var testDbHost string = "192.168.0.1"
	var testEmailSecure = "1"
	var expectEmailSecure = true

	setConfigValue(configValue.FieldByName("Port"), testPort)
	setConfigValue(configValue.FieldByName("CookieHash"), testCookieHash)
	setConfigValue(configValue.FieldByName("MaxParallelTasks"), testMaxParallelTasks)
	setConfigValue(configValue.FieldByName("LdapNeedTLS"), testLdapNeedTls)
	//setConfigValue(configValue.FieldByName("BoltDb.Hostname"), testDbHost)
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
	//setConfigValue(configValue.FieldByName("Not.Existent"), "someValue")

}

func TestLoadConfigEnvironmet(t *testing.T) {

	Config = new(ConfigType)
	Config.BoltDb = &DbConfig{}
	Config.Dialect = DbDriverBolt

	var envPort = "1337"
	var envCookieHash = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var envAccessKeyEncryption = "1/wRYXQltDGwbzNZRP9ZfJb2IoWcn1hYrxA0vOdvVos="
	var envMaxParallelTasks = "5"
	var expectMaxParallelTasks = 5
	var expectLdapNeedTls = true
	var envLdapNeedTls = "1"
	var envDbHost = "192.168.0.1"

	os.Setenv("SEMAPHORE_PORT", envPort)
	os.Setenv("SEMAPHORE_COOKIE_HASH", envCookieHash)
	os.Setenv("SEMAPHORE_ACCESS_KEY_ENCRYPTION", envAccessKeyEncryption)
	os.Setenv("SEMAPHORE_MAX_PARALLEL_TASKS", envMaxParallelTasks)
	os.Setenv("SEMAPHORE_LDAP_NEEDTLS", envLdapNeedTls)
	os.Setenv("SEMAPHORE_DB_HOST", envDbHost)

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
	if Config.BoltDb.Hostname != envDbHost {
		t.Error("Setting 'BoltDb.Hostname' was not loaded from environment-vars!")
	}

	//if Config.MySQL.Hostname == envDbHost || Config.Postgres.Hostname == envDbHost {
	//	// inactive db-dialects could be set as they share the same env-vars; but should be ignored
	//	t.Error("DB-Hostname was loaded for inactive DB-dialects!")
	//}

}

func TestLoadConfigDefaults(t *testing.T) {

	Config = new(ConfigType)
	var errMsg = "Failed to load config-default"

	loadConfigDefaults()

	if Config.Port != ":3000" {
		t.Error(errMsg)
	}
	if Config.TmpPath != "/tmp/semaphore" {
		t.Error(errMsg)
	}
}

func ensureConfigValidationFailure(t *testing.T, attribute string, value interface{}) {

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
	//assert := assert.New(t)

	Config = new(ConfigType)

	var testPort = ":3000"
	var testDbDialect = DbDriverBolt
	var testCookieHash = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var testMaxParallelTasks = 0

	Config.Port = testPort
	Config.Dialect = testDbDialect
	Config.CookieHash = testCookieHash
	Config.MaxParallelTasks = testMaxParallelTasks
	Config.GitClientId = GoGitClientId
	Config.CookieEncryption = testCookieHash
	Config.AccessKeyEncryption = testCookieHash
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

	//Config.CookieHash = "\"0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ=\"" // invalid with quotes (can happen when supplied as env-var)
	//ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	//Config.CookieHash = "!)394340"
	//ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	//Config.CookieHash = ""
	//ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	//Config.CookieHash = "TQwjDZ5fIQtaIw==" // valid b64, but too small
	//ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)
	Config.CookieHash = testCookieHash

	Config.Dialect = "someOtherDB"
	ensureConfigValidationFailure(t, "Dialect", Config.Dialect)
	Config.Dialect = testDbDialect

}
