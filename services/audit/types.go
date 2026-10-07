package audit

// Type is the ECS event.type.
type Type string

const (
	TypeCreation Type = "creation"
	TypeChange   Type = "change"
	TypeDeletion Type = "deletion"
	TypeAccess   Type = "access"
	TypeStart    Type = "start"
	TypeEnd      Type = "end"
	TypeDenied   Type = "denied"
	TypeInfo     Type = "info"
)

func validTypes() []Type {
	return []Type{TypeCreation, TypeChange, TypeDeletion, TypeAccess, TypeStart, TypeEnd, TypeDenied, TypeInfo}
}

func validCategories() []string {
	return []string{"auth", "iam", "resource", "secret", "task", "runner", "system", "audit"}
}

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

// Reason is an enum, never error text.
type Reason string

const (
	ReasonNone                     Reason = ""
	ReasonInvalidCredentials       Reason = "invalid_credentials"
	ReasonUserNotFound             Reason = "user_not_found"
	ReasonMethodDisabled           Reason = "method_disabled"
	ReasonProviderError            Reason = "provider_error"
	ReasonInvalidState             Reason = "invalid_state"
	ReasonInternalError            Reason = "internal_error"
	ReasonInvalidPasscode          Reason = "invalid_passcode"
	ReasonInvalidRecoveryCode      Reason = "invalid_recovery_code"
	ReasonCodeExpired              Reason = "code_expired"
	ReasonTooManyAttempts          Reason = "too_many_attempts"
	ReasonTokenUnknown             Reason = "token_unknown"
	ReasonTokenExpired             Reason = "token_expired"
	ReasonForbidden                Reason = "forbidden"
	ReasonCrossOrigin              Reason = "cross_origin"
	ReasonInvalidCurrentPassword   Reason = "invalid_current_password"
	ReasonOwnerSelfChange          Reason = "owner_self_change"
	ReasonActivationFailed         Reason = "activation_failed"
	ReasonSecretFailed             Reason = "secret_failed"
	ReasonInventoryFailed          Reason = "inventory_failed"
	ReasonSetupFailed              Reason = "setup_failed"
	ReasonRestoreFailed            Reason = "restore_failed"
	ReasonInvalidRegistrationToken Reason = "invalid_registration_token"
	ReasonInvalidStatus            Reason = "invalid_status"
)

func knownReasons() []Reason {
	return []Reason{
		ReasonInvalidCredentials, ReasonUserNotFound, ReasonMethodDisabled, ReasonProviderError,
		ReasonInvalidState, ReasonInternalError, ReasonInvalidPasscode, ReasonInvalidRecoveryCode,
		ReasonCodeExpired, ReasonTooManyAttempts, ReasonTokenUnknown, ReasonTokenExpired,
		ReasonForbidden, ReasonCrossOrigin, ReasonInvalidCurrentPassword, ReasonOwnerSelfChange,
		ReasonActivationFailed, ReasonSecretFailed, ReasonInventoryFailed, ReasonSetupFailed, ReasonRestoreFailed,
		ReasonInvalidRegistrationToken, ReasonInvalidStatus,
	}
}

type ActorType string

const (
	ActorUser        ActorType = "user"
	ActorAnonymous   ActorType = "anonymous"
	ActorSystem      ActorType = "system"
	ActorRunner      ActorType = "runner"
	ActorIntegration ActorType = "integration"
)

type AuthMethod string

const (
	AuthSession  AuthMethod = "session"
	AuthAPIToken AuthMethod = "api_token"
)

const (
	TargetUser                  = "user"
	TargetAPIToken              = "api_token"
	TargetRoute                 = "route"
	TargetRole                  = "role"
	TargetProjectRoleDefinition = "project_role_definition"
	TargetTemplatePermission    = "template_permission"
	TargetLicense               = "license"
	TargetProject               = "project"
	TargetInventory             = "inventory"
	TargetRepository            = "repository"
	TargetTemplate              = "template"
	TargetSchedule              = "schedule"
	TargetIntegration           = "integration"
	TargetIntegrationMatcher    = "integration_matcher"
	TargetIntegrationExtractor  = "integration_extractor"
	TargetIntegrationAlias      = "integration_alias"
	TargetHostConfig            = "host_config"
	TargetWorkflow              = "workflow"
	TargetEnvironment           = "environment"
	TargetCredential            = "credential"
	TargetSecretStorage         = "secret_storage"
	TargetTask                  = "task"
	TargetRunner                = "runner"
)

const (
	ComponentScheduler  = "scheduler"
	ComponentTaskRunner = "task_runner"
	ComponentReconciler = "reconciler"
	ComponentServer     = "server"
)

// Equal to db.IdentityTypeLdap and db.IdentityTypeOidc.
const (
	LoginMethodPassword = "password"
	LoginMethodLDAP     = "ldap"
	LoginMethodOIDC     = "oidc"
)

// Values of metadata.trigger of task.execution/create.
const (
	TriggerAPI         = "api"
	TriggerSchedule    = "schedule"
	TriggerIntegration = "integration"
	TriggerAutorun     = "autorun"
	TriggerWorkflow    = "workflow"
)

// Values of metadata.end_reason of task.execution/complete.
const (
	EndReasonTimeout    = "timeout"
	EndReasonRunnerLost = "runner_lost"
)

// Values of metadata.token of runner.lifecycle/register.
const (
	RunnerTokenOneTime = "one_time"
	RunnerTokenGlobal  = "global"
)
