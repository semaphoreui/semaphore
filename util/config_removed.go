package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// The `process` section of the configuration (user, uid, gid, chroot,
// no_new_privs, app_namespaces) restricted the processes Semaphore starts for a
// task. Semaphore no longer applies any of it. A configuration that still
// carries the section would load without complaint, because the decoders ignore
// unknown keys and nothing reads the variables, and tasks would then run with
// the privileges of the Semaphore process itself. This file detects such a
// configuration so that startup can refuse it instead.

// removedProcessEnvVars lists the environment variables of the removed
// `process` section.
func removedProcessEnvVars() []string {
	return []string{
		"SEMAPHORE_PROCESS_USER",
		"SEMAPHORE_PROCESS_UID",
		"SEMAPHORE_PROCESS_GID",
		"SEMAPHORE_PROCESS_CHROOT",
		"SEMAPHORE_PROCESS_NO_NEW_PRIVS",
		"SEMAPHORE_PROCESS_APP_NS_USER",
		"SEMAPHORE_PROCESS_APP_NS_MOUNT",
		"SEMAPHORE_PROCESS_APP_NS_PID",
		"SEMAPHORE_PROCESS_APP_NS_IPC",
		"SEMAPHORE_PROCESS_APP_NS_UTS",
	}
}

// checkRemovedProcessSettings returns an error naming every removed `process`
// setting the configuration still switches on, in the config file at
// configPath (nil when no file was loaded) and in the environment. It returns
// nil when there is none: an empty `process` block, or one holding only empty,
// false and zero values, restricted nothing and is accepted.
func checkRemovedProcessSettings(configPath *string) error {
	var found []string

	if configPath != nil {
		inFile, err := removedProcessSettingsInFile(*configPath)
		if err != nil {
			return fmt.Errorf("cannot check %s for the removed \"process\" settings: %w", *configPath, err)
		}
		for _, name := range inFile {
			found = append(found, fmt.Sprintf("%s (config file %s)", name, *configPath))
		}
	}

	for _, name := range removedProcessSettingsInEnv() {
		found = append(found, name+" (environment variable)")
	}

	if len(found) == 0 {
		return nil
	}

	return fmt.Errorf(`the "process" settings are no longer supported, but this configuration still sets:

  %s

Semaphore no longer switches the user, group or root directory, sets no_new_privs
or creates namespaces for the processes it starts. Ansible, Terraform/OpenTofu,
shell tasks and git now run with the privileges of the Semaphore process itself.

To keep tasks restricted, apply the restrictions to the Semaphore service (server
or runner) instead: for example User=, Group= and NoNewPrivileges= in the systemd
unit, or a non-root user for the container. Then remove the settings listed above
and start Semaphore again`, strings.Join(found, "\n  "))
}

// removedProcessSettingsInFile lists the settings switched on under the
// `process` key of the config file, read the way decodeConfig reads it: YAML by
// extension, JSON otherwise, and only the first document of the file.
func removedProcessSettingsInFile(configPath string) ([]string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if isYAMLConfig(configPath) {
		var raw any
		if err = yaml.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		if data, err = json.Marshal(raw); err != nil {
			return nil, err
		}
	}

	// A struct field, not a map lookup: encoding/json matches a key to a field
	// without regard to case, so "Process" reached the old section as well.
	var document struct {
		Process any `json:"process"`
	}
	if err = json.NewDecoder(bytes.NewReader(data)).Decode(&document); err != nil {
		return nil, err
	}

	return enabledSettings("process", document.Process), nil
}

// enabledSettings walks a decoded JSON value and returns the dotted names of
// the values in it that switch something on: everything except null, an empty
// string, false, zero and an empty object or array.
func enabledSettings(name string, value any) (names []string) {
	switch v := value.(type) {
	case nil:
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			names = append(names, enabledSettings(name+"."+key, v[key])...)
		}
	case []any:
		if len(v) > 0 {
			names = append(names, name)
		}
	case string:
		if v != "" {
			names = append(names, name)
		}
	case bool:
		if v {
			names = append(names, name)
		}
	case float64:
		if v != 0 {
			names = append(names, name)
		}
	default:
		names = append(names, name)
	}
	return
}

// removedProcessSettingsInEnv lists the removed variables that are set to a
// value other than empty, 0 or false.
func removedProcessSettingsInEnv() (names []string) {
	for _, name := range removedProcessEnvVars() {
		value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
		if value != "" && value != "0" && value != "false" {
			names = append(names, name)
		}
	}
	return
}
