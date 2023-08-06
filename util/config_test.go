package util

import (
	"fmt"
	"os"
	"testing"
)

func mockError(msg string) {
	panic(msg)
}

func TestCastStringToInt(t *testing.T) {

	var errMsg string = "Cast string => int failed"

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

	var errMsg string = "Cast string => bool failed"

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

	var testPort string = "1337"
	var testCookieHash string = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var testMaxParallelTasks int = 5
	var testLdapNeedTls bool = true
	var testDbHost string = "192.168.0.1"

	Config.Port = testPort
	Config.CookieHash = testCookieHash
	Config.MaxParallelTasks = testMaxParallelTasks
	Config.LdapNeedTLS = testLdapNeedTls
	Config.BoltDb.Hostname = testDbHost

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

	var testPort string = "1337"
	var testCookieHash string = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var testMaxParallelTasks int = 5
	var testLdapNeedTls bool = true
	var testDbHost string = "192.168.0.1"
	var testEmailSecure string = "1"
	var expectEmailSecure bool = true

	setConfigValue("Port", testPort)
	setConfigValue("CookieHash", testCookieHash)
	setConfigValue("MaxParallelTasks", testMaxParallelTasks)
	setConfigValue("LdapNeedTLS", testLdapNeedTls)
	setConfigValue("BoltDb.Hostname", testDbHost)
	setConfigValue("EmailSecure", testEmailSecure)

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
	if Config.BoltDb.Hostname != testDbHost {
		t.Error("Could not set value for config attribute 'BoltDb.Hostname'!")
	}
	if Config.EmailSecure != expectEmailSecure {
		t.Error("Could not set value for config attribute 'EmailSecure'!")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not fail on non-existent config attribute!")
		}
	}()
	setConfigValue("NotExistent", "someValue")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not fail on non-existent config attribute!")
		}
	}()
	setConfigValue("Not.Existent", "someValue")

}

func TestLoadConfigEnvironmet(t *testing.T) {

	Config = new(ConfigType)
	Config.Dialect = DbDriverBolt

	var envPort string = "1337"
	var envCookieHash string = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var envAccessKeyEncryption string = "1/wRYXQltDGwbzNZRP9ZfJb2IoWcn1hYrxA0vOdvVos="
	var envMaxParallelTasks string = "5"
	var expectMaxParallelTasks int = 5
	var expectLdapNeedTls bool = true
	var envLdapNeedTls string = "1"
	var envDbHost string = "192.168.0.1"

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
	if Config.MySQL.Hostname == envDbHost || Config.Postgres.Hostname == envDbHost {
		// inactive db-dialects could be set as they share the same env-vars; but should be ignored
		t.Error("DB-Hostname was loaded for inactive DB-dialects!")
	}

}

func TestLoadConfigDefaults(t *testing.T) {

	Config = new(ConfigType)
	var errMsg string = "Failed to load config-default"

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
	validateConfig(mockError)

}

func TestValidateConfig(t *testing.T) {

	Config = new(ConfigType)

	var testPort string = ":3000"
	var testDbDialect DbDriver = DbDriverBolt
	var testCookieHash string = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	var testMaxParallelTasks int = 0

	Config.Port = testPort
	Config.Dialect = testDbDialect
	Config.CookieHash = testCookieHash
	Config.MaxParallelTasks = testMaxParallelTasks
	Config.GitClientId = GoGitClientId
	Config.CookieEncryption = testCookieHash
	Config.AccessKeyEncryption = testCookieHash
	validateConfig(mockError)

	Config.Port = "INVALID"
	ensureConfigValidationFailure(t, "Port", Config.Port)

	Config.Port = ":100000"
	ensureConfigValidationFailure(t, "Port", Config.Port)
	Config.Port = testPort

	Config.MaxParallelTasks = -1
	ensureConfigValidationFailure(t, "MaxParallelTasks", Config.MaxParallelTasks)
	Config.MaxParallelTasks = testMaxParallelTasks

	Config.CookieHash = "\"0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ=\"" // invalid with quotes (can happen when supplied as env-var)
	ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	Config.CookieHash = "!)394340"
	ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	Config.CookieHash = ""
	ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)

	Config.CookieHash = "TQwjDZ5fIQtaIw==" // valid b64, but too small
	ensureConfigValidationFailure(t, "CookieHash", Config.CookieHash)
	Config.CookieHash = testCookieHash

	Config.Dialect = "someOtherDB"
	ensureConfigValidationFailure(t, "Dialect", Config.Dialect)
	Config.Dialect = testDbDialect

}
