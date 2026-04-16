package api

import (
	"net/http"
	"runtime"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/util"
)

func getAdminInfo(w http.ResponseWriter, r *http.Request) {
	store := helpers.Store(r)

	// Database info
	dbInfo := map[string]any{
		"dialect": util.Config.Dialect,
	}

	// Authentication
	authInfo := map[string]any{
		"password_login_enabled": !util.Config.PasswordLoginDisable,
	}

	if util.Config.Auth != nil {
		if util.Config.Auth.Totp != nil {
			authInfo["totp_enabled"] = util.Config.Auth.Totp.Enabled
		} else {
			authInfo["totp_enabled"] = false
		}
		if util.Config.Auth.Email != nil {
			authInfo["email_otp_enabled"] = util.Config.Auth.Email.Enabled
		} else {
			authInfo["email_otp_enabled"] = false
		}
	}

	// LDAP
	authInfo["ldap_enabled"] = util.Config.LdapEnable

	// OpenID Connect providers
	oidcProviders := []string{}
	for name := range util.Config.OidcProviders {
		oidcProviders = append(oidcProviders, name)
	}
	authInfo["oidc_providers"] = oidcProviders

	// Notifications
	notifications := map[string]bool{
		"email":           util.Config.EmailAlert,
		"telegram":        util.Config.TelegramAlert,
		"slack":           util.Config.SlackAlert,
		"rocketchat":      util.Config.RocketChatAlert,
		"microsoft_teams": util.Config.MicrosoftTeamsAlert,
		"dingtalk":        util.Config.DingTalkAlert,
		"gotify":          util.Config.GotifyAlert,
	}

	// Cluster / HA
	haEnabled := util.HAEnabled()
	clusterInfo := map[string]any{
		"ha_enabled": haEnabled,
	}
	if haEnabled && util.Config.HA != nil {
		clusterInfo["node_id"] = util.Config.HA.NodeID
		nodeCount, err := store.GetNodeCount()
		if err == nil {
			clusterInfo["node_count"] = nodeCount
		}
		uiCount, err := store.GetUiCount()
		if err == nil {
			clusterInfo["ui_count"] = uiCount
		}
	}

	// Runners
	runnersInfo := map[string]any{
		"use_remote_runner": util.Config.UseRemoteRunner,
	}

	// Task settings
	taskSettings := map[string]any{
		"max_parallel_tasks":     util.Config.MaxParallelTasks,
		"max_task_duration_sec":  util.Config.MaxTaskDurationSec,
		"max_tasks_per_template": util.Config.MaxTasksPerTemplate,
	}

	// System
	systemInfo := map[string]any{
		"version":       util.Version(),
		"ansible":       util.AnsibleVersion(),
		"git_client":    util.Config.GitClientId,
		"go_version":    runtime.Version(),
		"go_arch":       runtime.GOARCH,
		"go_os":         runtime.GOOS,
		"tmp_path":      util.Config.TmpPath,
		"home_dir_mode": util.Config.HomeDirMode,
	}

	if util.Config.Schedule != nil && util.Config.Schedule.Timezone != "" {
		systemInfo["schedule_timezone"] = util.Config.Schedule.Timezone
	} else {
		systemInfo["schedule_timezone"] = "UTC"
	}

	// Feature flags
	features := map[string]any{
		"non_admin_can_create_project": util.Config.NonAdminCanCreateProject,
	}

	body := map[string]any{
		"system":        systemInfo,
		"database":      dbInfo,
		"auth":          authInfo,
		"notifications": notifications,
		"cluster":       clusterInfo,
		"runners":       runnersInfo,
		"task_settings": taskSettings,
		"features":      features,
	}

	helpers.WriteJSON(w, http.StatusOK, body)
}
