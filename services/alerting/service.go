package alerting

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// Logger receives per-delivery progress lines (the task log in production).
type Logger interface {
	Logf(format string, args ...any)
}

// Store is the persistence the service depends on.
type Store interface {
	db.AlertManager
	PayloadSource
	GetProject(projectID int) (db.Project, error)
	GetProjectUsers(projectID int, params db.RetrieveQueryParams) ([]db.UserWithProjectRole, error)
	GetAllAdmins() ([]db.User, error)
	GetAccessKey(projectID int, accessKeyID int) (db.AccessKey, error)
}

// SecretDecryptor decrypts access keys; server.AccessKeyEncryptionService
// satisfies it.
type SecretDecryptor interface {
	DeserializeSecret(key *db.AccessKey) error
}

// Service is the entry point used by the task pool and the API.
type Service interface {
	Registry() *Registry
	// Channels describes the channels for the alert form of a project.
	Channels(project db.Project) []ChannelInfo
	// ValidateAlert normalizes an alert and applies the channel rules.
	ValidateAlert(alert *db.Alert) error
	// Snapshot freezes the alerting decision on a task before it is stored.
	Snapshot(task *db.Task, template db.Template) error
	// Notify delivers the notifications a status change produces. It is
	// safe to call for every status: statuses without an event are ignored.
	Notify(ctx context.Context, task db.Task, template db.Template, status task_logger.TaskStatus, logger Logger) error
	// SendTest delivers a test message to one alert, ignoring its events
	// and enabled flag.
	SendTest(ctx context.Context, project db.Project, alert db.Alert) error
	// SendProjectTest delivers a test message to every destination the
	// project would use in default mode. It returns how many were tried.
	SendProjectTest(ctx context.Context, project db.Project) (int, error)
}

type service struct {
	store     Store
	registry  *Registry
	decryptor SecretDecryptor
	config    func() *util.ConfigType
}

// NewService wires the registry to the store. The server config is read on
// every send so the legacy config.json channels keep working exactly as
// before and never need a restart-time snapshot.
func NewService(store Store, registry *Registry, decryptor SecretDecryptor) Service {
	return NewServiceWithConfig(store, registry, decryptor, func() *util.ConfigType { return util.Config })
}

func NewServiceWithConfig(store Store, registry *Registry, decryptor SecretDecryptor, config func() *util.ConfigType) Service {
	return &service{store: store, registry: registry, decryptor: decryptor, config: config}
}

func (s *service) Registry() *Registry {
	return s.registry
}

func (s *service) Channels(project db.Project) []ChannelInfo {
	return s.registry.Infos(s.config(), project)
}

func (s *service) ValidateAlert(alert *db.Alert) error {
	alert.Normalize()
	if err := alert.Validate(); err != nil {
		return err
	}
	channel, err := s.registry.Get(alert.Type)
	if err != nil {
		return err
	}
	clearUnusedFields(channel, alert)

	dest, err := s.destinationForAlert(channel, *alert)
	if err != nil {
		return err
	}
	if err = channel.Validate(s.config(), dest); err != nil {
		return err
	}
	return ValidateBody(channel, strValue(alert.Body))
}

// clearUnusedFields drops destination values the channel does not declare so
// a type change in the UI never leaves stale data behind.
func clearUnusedFields(channel Channel, alert *db.Alert) {
	used := make(map[string]bool)
	for _, f := range channel.Fields() {
		used[f.Name] = true
	}
	if !used[FieldChatID] {
		alert.ChatID = nil
	}
	if !used[FieldThreadID] {
		alert.ThreadID = nil
	}
	if !used[FieldURL] {
		alert.URL = nil
	}
	if !used[FieldRecipients] {
		alert.Recipients = nil
	}
	if channel.Secret() == nil {
		alert.KeyID = nil
	}

	var params db.MapStringAnyField
	for name, value := range alert.Params {
		if !used[name] {
			continue
		}
		if params == nil {
			params = db.MapStringAnyField{}
		}
		params[name] = value
	}
	alert.Params = params
}

// loadSecret fetches and decrypts the access key an alert points at and
// checks it is of the type the channel expects.
func (s *service) loadSecret(channel Channel, alert db.Alert) (*db.AccessKey, error) {
	if alert.KeyID == nil {
		return nil, nil
	}
	spec := channel.Secret()
	if spec == nil {
		return nil, common_errors.NewValidationError(string(channel.Type()) + " alerts do not use an access key")
	}

	key, err := s.store.GetAccessKey(alert.ProjectID, *alert.KeyID)
	if err != nil {
		return nil, common_errors.NewValidationError("access key does not belong to this project")
	}
	if key.Type != spec.KeyType {
		return nil, common_errors.NewValidationError("access key must be of type " + string(spec.KeyType))
	}
	if s.decryptor != nil {
		if err = s.decryptor.DeserializeSecret(&key); err != nil {
			return nil, err
		}
	}
	return &key, nil
}

func (s *service) Snapshot(task *db.Task, template db.Template) error {
	project, err := s.store.GetProject(task.ProjectID)
	if err != nil {
		return err
	}

	var schedule *db.Schedule
	if task.ScheduleID != nil {
		sch, err := s.store.GetSchedule(task.ProjectID, *task.ScheduleID)
		if err != nil {
			return err
		}
		schedule = &sch
	}

	defaultIDs, err := s.store.GetDefaultAlertIDs(task.ProjectID)
	if err != nil {
		return err
	}

	snap := Resolve(project, template, schedule, defaultIDs)
	task.AlertSnapshot = &snap
	return nil
}

func (s *service) Notify(
	ctx context.Context,
	task db.Task,
	template db.Template,
	status task_logger.TaskStatus,
	logger Logger,
) error {
	event, ok := EventForStatus(status)
	if !ok {
		return nil
	}

	project, err := s.store.GetProject(task.ProjectID)
	if err != nil {
		return err
	}

	// Tasks created before the upgrade (or by code paths that skipped
	// Snapshot) are resolved on the fly with the current bindings.
	snapshot := task.AlertSnapshot
	if snapshot == nil {
		defaultIDs, err := s.store.GetDefaultAlertIDs(task.ProjectID)
		if err != nil {
			return err
		}
		resolved := Resolve(project, template, nil, defaultIDs)
		snapshot = &resolved
	}

	if !snapshot.Allows(event) || snapshot.IsEmpty() {
		return nil
	}

	cfg := s.config()
	payload := BuildPayload(cfg, s.store, project, template, task, status)

	var errs []error

	if snapshot.Instance {
		for _, channel := range s.registry.List() {
			dest, ok := channel.InstanceDestination(cfg, project)
			if !ok || !channel.DefaultEvents().Contains(event) {
				continue
			}
			key := "instance:" + string(channel.Type())
			dest = s.fillInstanceDestination(channel, project.ID, dest)
			if err := s.deliver(ctx, task.ID, key, channel.Title(), channel, dest, "", event, payload, logger); err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, alertID := range snapshot.AlertIDs {
		alert, err := s.store.GetAlert(task.ProjectID, alertID)
		if err != nil {
			logf(logger, "Alert %d skipped: %s", alertID, err.Error())
			continue
		}
		if !alert.Enabled {
			continue
		}
		channel, err := s.registry.Get(alert.Type)
		if err != nil {
			logf(logger, "Alert %s skipped: %s", alert.Name, err.Error())
			continue
		}
		events := alert.Events
		if len(events) == 0 {
			events = channel.DefaultEvents()
		}
		if !events.Contains(event) {
			continue
		}
		key := "alert:" + strconv.Itoa(alert.ID)
		dest, err := s.destinationForAlert(channel, alert)
		if err != nil {
			logf(logger, "Can't send alert %s: %s", alert.Name, err.Error())
			errs = append(errs, fmt.Errorf("alert %s: %w", alert.Name, err))
			continue
		}
		if err := s.deliver(ctx, task.ID, key, alert.Name, channel, dest, strValue(alert.Body), event, payload, logger); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// deliver claims the (task, destination, event) triple so that HA nodes do
// not double-send, renders the body and hands it to the channel.
func (s *service) deliver(
	ctx context.Context,
	taskID int,
	key string,
	name string,
	channel Channel,
	dest Destination,
	body string,
	event db.AlertEvent,
	payload Payload,
	logger Logger,
) error {
	if taskID > 0 {
		claimed, err := s.store.ClaimAlertSend(taskID, key, event)
		if err != nil {
			logf(logger, "Alert %s skipped: %s", name, err.Error())
			return err
		}
		if !claimed {
			return nil
		}
	}

	if err := s.send(ctx, channel, dest, body, payload); err != nil {
		logf(logger, "Can't send alert %s: %s", name, err.Error())
		return fmt.Errorf("alert %s: %w", name, err)
	}
	logf(logger, "Sent alert %s", name)
	return nil
}

func (s *service) send(ctx context.Context, channel Channel, dest Destination, body string, payload Payload) error {
	cfg := s.config()
	if cfg == nil {
		return common_errors.NewValidationError("server config is not loaded")
	}

	payload.Color = channel.StatusColor(payload.Task.Status)
	payload.Chat = PayloadChat{ID: dest.ChatID, ThreadID: dest.ThreadID}

	rendered, err := Render(channel, body, payload)
	if err != nil {
		return err
	}

	msg := Message{
		Subject: subjectFor(payload),
		Body:    rendered,
		Payload: payload,
	}
	return channel.Send(ctx, cfg, dest, msg)
}

// destinationForAlert attaches the decrypted secret and fills channel
// defaults the alert left empty. E-mail with no recipients goes to the
// project members who opted in, like the server-wide channel always did.
func (s *service) destinationForAlert(channel Channel, alert db.Alert) (Destination, error) {
	dest := DestinationFromAlert(alert)

	secret, err := s.loadSecret(channel, alert)
	if err != nil {
		return dest, err
	}
	dest.Secret = secret

	if wantsField(channel, FieldRecipients) && len(dest.Recipients) == 0 {
		dest.Recipients = s.defaultRecipients(alert.ProjectID)
	}
	return dest, nil
}

func (s *service) fillInstanceDestination(channel Channel, projectID int, dest Destination) Destination {
	if wantsField(channel, FieldRecipients) && len(dest.Recipients) == 0 {
		dest.Recipients = s.defaultRecipients(projectID)
	}
	return dest
}

// defaultRecipients are project members plus administrators who enabled
// alerts on their profile.
func (s *service) defaultRecipients(projectID int) []string {
	seen := make(map[int]bool)
	var out []string

	add := func(u db.User) {
		if seen[u.ID] || !u.Alert || u.Email == "" {
			return
		}
		seen[u.ID] = true
		out = append(out, u.Email)
	}

	if users, err := s.store.GetProjectUsers(projectID, db.RetrieveQueryParams{}); err == nil {
		for _, u := range users {
			add(u.User)
		}
	} else {
		log.WithError(err).Warn("alerting: can not load project users")
	}

	if admins, err := s.store.GetAllAdmins(); err == nil {
		for _, u := range admins {
			add(u)
		}
	} else {
		log.WithError(err).Warn("alerting: can not load admins")
	}

	return out
}

func (s *service) SendTest(ctx context.Context, project db.Project, alert db.Alert) error {
	channel, err := s.registry.Get(alert.Type)
	if err != nil {
		return err
	}
	payload := testPayload(s.config(), project)
	dest, err := s.destinationForAlert(channel, alert)
	if err != nil {
		return err
	}
	return s.send(ctx, channel, dest, strValue(alert.Body), payload)
}

func (s *service) SendProjectTest(ctx context.Context, project db.Project) (int, error) {
	cfg := s.config()
	payload := testPayload(cfg, project)

	var errs []error
	sent := 0

	if project.Alert {
		for _, channel := range s.registry.List() {
			dest, ok := channel.InstanceDestination(cfg, project)
			if !ok {
				continue
			}
			sent++
			dest = s.fillInstanceDestination(channel, project.ID, dest)
			if err := s.send(ctx, channel, dest, "", payload); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", channel.Title(), err))
			}
		}
	}

	alerts, err := s.store.GetAlerts(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		return sent, err
	}
	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}
		sent++
		if err := s.SendTest(ctx, project, alert); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", alert.Name, err))
		}
	}

	return sent, errors.Join(errs...)
}

func testPayload(cfg *util.ConfigType, project db.Project) Payload {
	return BuildPayload(
		cfg,
		nil,
		project,
		db.Template{ProjectID: project.ID, Name: "Test Notification", Type: db.TemplateTask},
		db.Task{ProjectID: project.ID, Status: task_logger.TaskSuccessStatus, Message: "This is a test notification"},
		task_logger.TaskSuccessStatus,
	)
}

func subjectFor(payload Payload) string {
	switch payload.Task.Status {
	case task_logger.TaskFailStatus:
		return fmt.Sprintf("Task '%s' failed", payload.Name)
	case task_logger.TaskSuccessStatus:
		return fmt.Sprintf("Task '%s' succeeded", payload.Name)
	case task_logger.TaskWaitingConfirmation:
		return fmt.Sprintf("Task '%s' is waiting for confirmation", payload.Name)
	default:
		return fmt.Sprintf("Task '%s' %s", payload.Name, payload.Task.Status)
	}
}

func wantsField(channel Channel, name string) bool {
	for _, f := range channel.Fields() {
		if f.Name == name {
			return true
		}
	}
	return false
}

func logf(logger Logger, format string, args ...any) {
	if logger == nil {
		log.Infof(format, args...)
		return
	}
	logger.Logf(format, args...)
}
