package alerting

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
)

// Registry is the ordered set of channels the server supports.
type Registry struct {
	channels []Channel
	byType   map[db.AlertType]Channel
}

// NewRegistry returns the built-in channels. Register a new channel here.
func NewRegistry() *Registry {
	r := &Registry{byType: make(map[db.AlertType]Channel)}
	for _, ch := range []Channel{
		newTelegramChannel(),
		newSlackChannel(),
		newEmailChannel(),
		newTeamsChannel(),
		newRocketChatChannel(),
		newDingTalkChannel(),
		newGotifyChannel(),
	} {
		r.Register(ch)
	}
	return r
}

// Register adds a channel; a channel with the same type replaces the old one.
func (r *Registry) Register(ch Channel) {
	if _, exists := r.byType[ch.Type()]; exists {
		for i, existing := range r.channels {
			if existing.Type() == ch.Type() {
				r.channels[i] = ch
			}
		}
	} else {
		r.channels = append(r.channels, ch)
	}
	r.byType[ch.Type()] = ch
}

func (r *Registry) Get(typ db.AlertType) (Channel, error) {
	ch, ok := r.byType[typ]
	if !ok {
		return nil, common_errors.NewValidationError("unknown alert type: " + string(typ))
	}
	return ch, nil
}

func (r *Registry) List() []Channel {
	return append([]Channel(nil), r.channels...)
}

// ChannelInfo describes a channel to API clients so the UI can build the
// alert form and explain what the server has configured.
type ChannelInfo struct {
	Type          db.AlertType   `json:"type"`
	Title         string         `json:"title"`
	Icon          string         `json:"icon"`
	Fields        []Field        `json:"fields"`
	DefaultEvents db.AlertEvents `json:"default_events"`
	BodyFormat    BodyFormat     `json:"body_format"`
	DefaultBody   string         `json:"default_body"`
	// Secret describes the access key the channel may use instead of the
	// server-wide secret; nil for channels without secrets.
	Secret *SecretSpec `json:"secret,omitempty"`
	// ServerConfigured is true when config.json enables this channel
	// server-wide, i.e. projects with "alert" on receive it automatically.
	ServerConfigured bool `json:"server_configured"`
	// ServerEnabled is true when config.json enables the channel even if a
	// project still needs to provide data such as a Telegram chat ID.
	ServerEnabled bool `json:"server_enabled"`
	// Ready is true when the server-wide secret exists, so an alert may rely
	// on it instead of bringing its own; ReadyError explains what is missing.
	Ready      bool   `json:"ready"`
	ReadyError string `json:"ready_error,omitempty"`
}

// Infos describes every channel for the given project.
func (r *Registry) Infos(cfg *util.ConfigType, project db.Project) []ChannelInfo {
	out := make([]ChannelInfo, 0, len(r.channels))
	for _, ch := range r.channels {
		info := ChannelInfo{
			Type:          ch.Type(),
			Title:         ch.Title(),
			Icon:          ch.Icon(),
			Fields:        ch.Fields(),
			DefaultEvents: ch.DefaultEvents(),
			BodyFormat:    ch.BodyFormat(),
			DefaultBody:   ch.DefaultBody(),
			Secret:        ch.Secret(),
			Ready:         true,
		}
		if info.Fields == nil {
			info.Fields = []Field{}
		}
		info.ServerEnabled = ch.ServerEnabled(cfg)
		_, info.ServerConfigured = ch.InstanceDestination(cfg, project)
		if err := ch.ServerReady(cfg); err != nil {
			info.Ready = false
			info.ReadyError = err.Error()
		}
		out = append(out, info)
	}
	return out
}
