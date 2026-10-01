package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

type contextKey int

const (
	actorKey contextKey = iota
	requestKey
)

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}

func ActorFrom(ctx context.Context) Actor {
	if actor, ok := ctx.Value(actorKey).(Actor); ok {
		return actor
	}
	return AnonymousActor()
}

func UserActor(id int, username string, auth AuthMethod, tokenFingerprint string) Actor {
	return Actor{Type: ActorUser, ID: strconv.Itoa(id), Name: username, Auth: auth, TokenFingerprint: tokenFingerprint}
}

func SystemActor(component string) Actor {
	return Actor{Type: ActorSystem, Name: component}
}

func RunnerActor(id int, name string) Actor {
	return Actor{Type: ActorRunner, ID: strconv.Itoa(id), Name: name}
}

func IntegrationActor(id int, name string) Actor {
	return Actor{Type: ActorIntegration, ID: strconv.Itoa(id), Name: name}
}

func AnonymousActor() Actor {
	return Actor{Type: ActorAnonymous}
}

type RequestInfo struct {
	ID        string
	IP        string
	UserAgent string
}

func WithRequest(ctx context.Context, info RequestInfo) context.Context {
	return context.WithValue(ctx, requestKey, info)
}

func RequestFrom(ctx context.Context) (RequestInfo, bool) {
	info, ok := ctx.Value(requestKey).(RequestInfo)
	return info, ok
}

// The token is its own ID, so only its hash is recorded.
func TokenFingerprint(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])[:16]
}
