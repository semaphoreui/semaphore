package db

type EventParams struct {
	UserID    int
	ProjectID int
	Query     RetrieveQueryParams
}

type UserParams struct {
	UserID   int
	Password string
	Query    RetrieveQueryParams
}

type TokenParams struct {
	UserID  int
	TokenID string
}

type SessionParams struct {
	UserID    int
	SessionID string
}

type RunnerParams struct {
	RunnerID  int
	ProjectID int
}

type ProjectParams struct {
	Admin     bool
	UserID    int
	ProjectID int
}

type MemberParams struct {
	ProjectID int
	UserID    int
	Query     RetrieveQueryParams
}

type TemplateParams struct {
	ProjectID  int
	TemplateID int
	Filter     TemplateFilter
	Query      RetrieveQueryParams
}

type AccessKeyParams struct {
	ProjectID   int
	AccessKeyID int
	OldKey      string
	Query       RetrieveQueryParams
}

type EnvParams struct {
	ProjectID     int
	EnvironmentID int
	Query         RetrieveQueryParams
}

type InventoryParams struct {
	ProjectID   int
	InventoryID int
	Query       RetrieveQueryParams
}

type RepoParams struct {
	ProjectID    int
	RepositoryID int
	Query        RetrieveQueryParams
}

type ViewParams struct {
	ProjectID int
	ViewID    int
	Positions map[int]int
	Query     RetrieveQueryParams
}

type ScheduleParams struct {
	ProjectID  int
	ScheduleID int
	TemplateID int
	Hash       string
	Query      RetrieveQueryParams
}

type TaskParams struct {
	ProjectID  int
	TaskID     int
	TemplateID int
	Query      RetrieveQueryParams
}

type IntegrationParams struct {
	ProjectID     int
	IntegrationID int
	Alias         string
	Query         RetrieveQueryParams
}

type IntegrationExtractValueParams struct {
	ProjectID     int
	IntegrationID int
	ValueID       int
	Query         RetrieveQueryParams
}

type IntegrationMatcherParams struct {
	ProjectID     int
	IntegrationID int
	MatcherID     int
	Query         RetrieveQueryParams
}

type IntegrationAliasParams struct {
	ProjectID     int
	IntegrationID int
	AliasID       int
}
