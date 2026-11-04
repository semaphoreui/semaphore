package db

import "strconv"

// Option keys of the server-wide MCP settings. util.Config.MCP is closed to the
// database (db:"-"), so FillConfigFromDB never lays these keys over the config:
// they are read from the option table on every request instead, and a change
// applies at once on every node.
const (
	MCPAllowRunOption          = "mcp.allow_run"
	MCPAllowDestroyOption      = "mcp.allow_destroy"
	MCPAdminAllowWriteOption   = "mcp.admin.allow_write"
	MCPAdminAllowDestroyOption = "mcp.admin.allow_destroy"
)

// MCPSettingsOptions lists the option keys of the MCP settings in a stable order.
func MCPSettingsOptions() []string {
	return []string{MCPAllowRunOption, MCPAllowDestroyOption, MCPAdminAllowWriteOption, MCPAdminAllowDestroyOption}
}

// MCPSettings is the ceiling an administrator sets for the MCP tools of the
// whole server, one pair of switches per tool set. It limits the MCP tools
// only: the same token still works on the REST API.
type MCPSettings struct {
	// AllowRun allows the tools of the project set that run and control tasks.
	AllowRun bool `json:"allow_run"`
	// AllowDestroy allows the delete_* tools of the project set.
	AllowDestroy bool `json:"allow_destroy"`
	// AdminAllowWrite allows the save_* tools and the other changes of the admin set.
	AdminAllowWrite bool `json:"admin_allow_write"`
	// AdminAllowDestroy allows the delete_* tools of the admin set.
	AdminAllowDestroy bool `json:"admin_allow_destroy"`
}

// DefaultMCPSettings is what a server without stored settings has: every
// group is off until an administrator turns it on.
func DefaultMCPSettings() MCPSettings {
	return MCPSettings{AllowRun: false, AllowDestroy: false, AdminAllowWrite: false, AdminAllowDestroy: false}
}

// fields pairs each option key with its value in the settings, in the order of MCPSettingsOptions.
func (s *MCPSettings) fields() []struct {
	key   string
	value *bool
} {
	return []struct {
		key   string
		value *bool
	}{
		{MCPAllowRunOption, &s.AllowRun},
		{MCPAllowDestroyOption, &s.AllowDestroy},
		{MCPAdminAllowWriteOption, &s.AdminAllowWrite},
		{MCPAdminAllowDestroyOption, &s.AdminAllowDestroy},
	}
}

// MCPSettingsReader is the part of the store ReadMCPSettings needs.
type MCPSettingsReader interface {
	GetOption(key string) (string, error)
}

// MCPSettingsWriter is the part of the store WriteMCPSettings needs.
type MCPSettingsWriter interface {
	SetOption(key string, value string) error
}

// ReadMCPSettings reads the settings from the store. A missing row, or a value
// that is not a boolean, gives the default of that setting: off.
func ReadMCPSettings(store MCPSettingsReader) (MCPSettings, error) {
	settings := DefaultMCPSettings()
	for _, field := range settings.fields() {
		value, err := readMCPOption(store, field.key, *field.value)
		if err != nil {
			return MCPSettings{}, err
		}
		*field.value = value
	}
	return settings, nil
}

// WriteMCPSettings stores every setting as "true" or "false".
func WriteMCPSettings(store MCPSettingsWriter, settings MCPSettings) error {
	for _, field := range settings.fields() {
		if err := store.SetOption(field.key, strconv.FormatBool(*field.value)); err != nil {
			return err
		}
	}
	return nil
}

func readMCPOption(store MCPSettingsReader, key string, fallback bool) (bool, error) {
	value, err := store.GetOption(key)
	if err != nil {
		return false, err
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback, nil
	}
	return parsed, nil
}
