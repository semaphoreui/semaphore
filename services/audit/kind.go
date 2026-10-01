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

	AuditLifecycleStart Kind = "audit.lifecycle/start"
)
