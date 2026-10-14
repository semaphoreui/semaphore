package audit

import "strings"

// Kind is "<event_code>/<action>".
type Kind string

func (k Kind) Code() string {
	code, _, _ := strings.Cut(string(k), "/")
	return code
}

func (k Kind) Action() string {
	_, action, _ := strings.Cut(string(k), "/")
	return action
}

func (k Kind) Category() string {
	category, _, _ := strings.Cut(k.Code(), ".")
	return category
}

const (
	AuthLogin             Kind = "auth.login/authenticate"
	AuthLogout            Kind = "auth.logout/terminate_session"
	AuthMFAVerifyTOTP     Kind = "auth.mfa/verify_totp"
	AuthMFAVerifyEmail    Kind = "auth.mfa/verify_email"
	AuthMFARecover        Kind = "auth.mfa/recover"
	AuthAPITokenReject    Kind = "auth.api_token/reject"
	AuthAuthorizationDeny Kind = "auth.authorization/deny"
	AuthCSRFBlock         Kind = "auth.csrf/block"

	IAMUserCreate                  Kind = "iam.user/create"
	IAMUserUpdate                  Kind = "iam.user/update"
	IAMUserDelete                  Kind = "iam.user/delete"
	IAMUserAutoProvision           Kind = "iam.user/auto_provision"
	IAMUserPasswordChange          Kind = "iam.user_password/change"
	IAMUserPasswordAdminReset      Kind = "iam.user_password/admin_reset"
	IAMMFAEnable                   Kind = "iam.mfa/enable"
	IAMMFADisable                  Kind = "iam.mfa/disable"
	IAMMFAViewQR                   Kind = "iam.mfa/view_qr"
	IAMExternalIdentityLink        Kind = "iam.external_identity/link"
	IAMExternalIdentityUnlink      Kind = "iam.external_identity/unlink"
	IAMAPITokenCreate              Kind = "iam.api_token/create"
	IAMAPITokenDelete              Kind = "iam.api_token/delete"
	IAMMembershipAdd               Kind = "iam.membership/add"
	IAMMembershipRemove            Kind = "iam.membership/remove"
	IAMProjectRoleChange           Kind = "iam.project_role/change"
	IAMRoleCreate                  Kind = "iam.role/create"
	IAMRoleUpdate                  Kind = "iam.role/update"
	IAMRoleDelete                  Kind = "iam.role/delete"
	IAMProjectRoleDefinitionCreate Kind = "iam.project_role_definition/create"
	IAMProjectRoleDefinitionUpdate Kind = "iam.project_role_definition/update"
	IAMProjectRoleDefinitionDelete Kind = "iam.project_role_definition/delete"
	IAMTemplatePermissionCreate    Kind = "iam.template_permission/create"
	IAMTemplatePermissionUpdate    Kind = "iam.template_permission/update"
	IAMTemplatePermissionDelete    Kind = "iam.template_permission/delete"

	SystemSettingsUpdate  Kind = "system.settings/update"
	SystemLicenseActivate Kind = "system.license/activate"

	AuditLifecycleStart  Kind = "audit.lifecycle/start"
	AuditRetentionDelete Kind = "audit.retention/delete"
	AuditLogExport       Kind = "audit.log/export"

	ResourceProjectCreate               Kind = "resource.project/create"
	ResourceProjectUpdate               Kind = "resource.project/update"
	ResourceProjectDelete               Kind = "resource.project/delete"
	ResourceProjectBackupExport         Kind = "resource.project_backup/export"
	ResourceProjectBackupRestore        Kind = "resource.project_backup/restore"
	ResourceInventoryCreate             Kind = "resource.inventory/create"
	ResourceInventoryUpdate             Kind = "resource.inventory/update"
	ResourceInventoryDelete             Kind = "resource.inventory/delete"
	ResourceRepositoryCreate            Kind = "resource.repository/create"
	ResourceRepositoryUpdate            Kind = "resource.repository/update"
	ResourceRepositoryDelete            Kind = "resource.repository/delete"
	ResourceTemplateCreate              Kind = "resource.template/create"
	ResourceTemplateUpdate              Kind = "resource.template/update"
	ResourceTemplateDelete              Kind = "resource.template/delete"
	ResourceTemplateAttachInventory     Kind = "resource.template/attach_inventory"
	ResourceTemplateDetachInventory     Kind = "resource.template/detach_inventory"
	ResourceTemplateSetDefaultInventory Kind = "resource.template/set_default_inventory"
	ResourceScheduleCreate              Kind = "resource.schedule/create"
	ResourceScheduleUpdate              Kind = "resource.schedule/update"
	ResourceScheduleDelete              Kind = "resource.schedule/delete"
	ResourceScheduleActivate            Kind = "resource.schedule/activate"
	ResourceScheduleDeactivate          Kind = "resource.schedule/deactivate"
	ResourceIntegrationCreate           Kind = "resource.integration/create"
	ResourceIntegrationUpdate           Kind = "resource.integration/update"
	ResourceIntegrationDelete           Kind = "resource.integration/delete"
	ResourceIntegrationMatcherCreate    Kind = "resource.integration_matcher/create"
	ResourceIntegrationMatcherUpdate    Kind = "resource.integration_matcher/update"
	ResourceIntegrationMatcherDelete    Kind = "resource.integration_matcher/delete"
	ResourceIntegrationExtractorCreate  Kind = "resource.integration_extractor/create"
	ResourceIntegrationExtractorUpdate  Kind = "resource.integration_extractor/update"
	ResourceIntegrationExtractorDelete  Kind = "resource.integration_extractor/delete"
	ResourceIntegrationAliasCreate      Kind = "resource.integration_alias/create"
	ResourceIntegrationAliasDelete      Kind = "resource.integration_alias/delete"
	ResourceHostConfigCreate            Kind = "resource.host_config/create"
	ResourceHostConfigUpdate            Kind = "resource.host_config/update"
	ResourceHostConfigDelete            Kind = "resource.host_config/delete"
	ResourceWorkflowCreate              Kind = "resource.workflow/create"
	ResourceWorkflowUpdate              Kind = "resource.workflow/update"
	ResourceWorkflowDelete              Kind = "resource.workflow/delete"
	ResourceEnvironmentCreate           Kind = "resource.environment/create"
	ResourceEnvironmentUpdate           Kind = "resource.environment/update"
	ResourceEnvironmentDelete           Kind = "resource.environment/delete"
	ResourceEnvironmentSync             Kind = "resource.environment/sync"

	SecretCredentialCreate Kind = "secret.credential/create"
	SecretCredentialUpdate Kind = "secret.credential/update"
	SecretCredentialDelete Kind = "secret.credential/delete"
	SecretStorageCreate    Kind = "secret.storage/create"
	SecretStorageUpdate    Kind = "secret.storage/update"
	SecretStorageDelete    Kind = "secret.storage/delete"
	SecretStorageSync      Kind = "secret.storage/sync"

	TaskExecutionCreate   Kind = "task.execution/create"
	TaskExecutionComplete Kind = "task.execution/complete"
	TaskApprovalRequest   Kind = "task.approval/request"
	TaskApprovalApprove   Kind = "task.approval/approve"
	TaskApprovalReject    Kind = "task.approval/reject"
	TaskControlStop       Kind = "task.control/stop"
	TaskControlForceStop  Kind = "task.control/force_stop"
	TaskControlStopAll    Kind = "task.control/stop_all"
	TaskHistoryDelete     Kind = "task.history/delete"

	RunnerLifecycleCreate     Kind = "runner.lifecycle/create"
	RunnerLifecycleUpdate     Kind = "runner.lifecycle/update"
	RunnerLifecycleDelete     Kind = "runner.lifecycle/delete"
	RunnerLifecycleEnable     Kind = "runner.lifecycle/enable"
	RunnerLifecycleDisable    Kind = "runner.lifecycle/disable"
	RunnerLifecycleRegister   Kind = "runner.lifecycle/register"
	RunnerLifecycleUnregister Kind = "runner.lifecycle/unregister"
	RunnerCredentialRotate    Kind = "runner.credential/rotate_registration_token"
	RunnerCacheClear          Kind = "runner.cache/clear"
	RunnerProgressReject      Kind = "runner.progress/reject"
)
