package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActorFromDefaultsToAnonymous(t *testing.T) {
	assert.Equal(t, Actor{Type: ActorAnonymous}, ActorFrom(context.Background()))
}

func TestActorRoundTrip(t *testing.T) {
	ctx := WithActor(context.Background(), UserActor(7, "alice", AuthAPIToken, "3f9a0c1d2e4b5a69"))
	assert.Equal(t, Actor{Type: ActorUser, ID: "7", Name: "alice", Auth: AuthAPIToken, TokenFingerprint: "3f9a0c1d2e4b5a69"}, ActorFrom(ctx))
}

func TestActorConstructors(t *testing.T) {
	assert.Equal(t, Actor{Type: ActorSystem, Name: ComponentScheduler}, SystemActor(ComponentScheduler))
	assert.Equal(t, Actor{Type: ActorRunner, ID: "3", Name: "r1"}, RunnerActor(3, "r1"))
	assert.Equal(t, Actor{Type: ActorIntegration, ID: "4", Name: "gh"}, IntegrationActor(4, "gh"))
}

func TestRequestInfoRoundTrip(t *testing.T) {
	_, ok := RequestFrom(context.Background())
	assert.False(t, ok)

	info := RequestInfo{ID: "id", IP: "10.0.0.5", UserAgent: "curl/8.5"}
	got, ok := RequestFrom(WithRequest(context.Background(), info))
	assert.True(t, ok)
	assert.Equal(t, info, got)
}

func TestTokenFingerprint(t *testing.T) {
	assert.Equal(t, "ba7816bf8f01cfea", TokenFingerprint("abc"))
}

func TestUserTarget(t *testing.T) {
	assert.Equal(t, &Target{Type: TargetUser, ID: "7", Name: "alice"}, UserTarget(7, "alice"))
}
