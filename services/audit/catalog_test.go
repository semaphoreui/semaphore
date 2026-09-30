package audit

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogIsWellFormed(t *testing.T) {
	codeRE := regexp.MustCompile(`^[a-z]+\.[a-z_]+$`)
	actionRE := regexp.MustCompile(`^[a-z_]+$`)
	seen := map[Kind]bool{}

	for _, entry := range Catalog() {
		t.Run(string(entry.Kind), func(t *testing.T) {
			assert.False(t, seen[entry.Kind], "duplicate kind")
			seen[entry.Kind] = true

			assert.Regexp(t, codeRE, entry.Kind.Code())
			assert.LessOrEqual(t, len(entry.Kind.Code()), 32, "event_code is the Syslog MSGID")
			assert.Regexp(t, actionRE, entry.Kind.Action())
			assert.Contains(t, validCategories(), entry.Kind.Category())
			assert.Contains(t, validTypes(), entry.Type)
			if entry.Type == TypeDenied {
				assert.NotEmpty(t, entry.Reasons, "a denial is always a failure")
			}
			for _, reason := range entry.Reasons {
				assert.Contains(t, knownReasons(), reason)
			}
			if entry.Metadata != nil {
				assert.Equal(t, reflect.Struct, reflect.TypeOf(entry.Metadata).Kind())
				_, err := json.Marshal(entry.Metadata)
				assert.NoError(t, err)
			}
		})
	}
}

func declaredKinds(t *testing.T) []Kind {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "kind.go", nil, 0)
	require.NoError(t, err)

	var kinds []Kind
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			if ident, ok := value.Type.(*ast.Ident); !ok || ident.Name != "Kind" {
				continue
			}
			for _, expr := range value.Values {
				kind, err := strconv.Unquote(expr.(*ast.BasicLit).Value)
				require.NoError(t, err)
				kinds = append(kinds, Kind(kind))
			}
		}
	}
	return kinds
}

func TestCatalogHasEveryKind(t *testing.T) {
	kinds := declaredKinds(t)
	require.NotEmpty(t, kinds)
	for _, kind := range kinds {
		assert.True(t, Known(kind), kind)
	}
	assert.Equal(t, len(kinds), len(Catalog()))
}

func TestKindParts(t *testing.T) {
	assert.Equal(t, "iam.api_token", IAMAPITokenCreate.Code())
	assert.Equal(t, "create", IAMAPITokenCreate.Action())
	assert.Equal(t, "iam", IAMAPITokenCreate.Category())
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{"success by default", Event{Kind: IAMUserDelete}, false},
		{"unknown kind", Event{Kind: Kind("iam.nothing/create")}, true},
		{"failure on success-only kind", Event{Kind: IAMUserDelete, Outcome: OutcomeFailure, Reason: ReasonForbidden}, true},
		{"failure without reason", Event{Kind: AuthLogin, Outcome: OutcomeFailure, Metadata: AuthMethodMetadata{}}, true},
		{"failure with reason", Event{Kind: AuthLogin, Outcome: OutcomeFailure, Reason: ReasonInvalidCredentials, Metadata: AuthMethodMetadata{}}, false},
		{"unknown reason", Event{Kind: AuthLogin, Outcome: OutcomeFailure, Reason: Reason("boom"), Metadata: AuthMethodMetadata{}}, true},
		{"reason of another kind", Event{Kind: AuthLogin, Outcome: OutcomeFailure, Reason: ReasonCrossOrigin, Metadata: AuthMethodMetadata{}}, true},
		{"reason on success", Event{Kind: AuthLogin, Reason: ReasonInvalidCredentials, Metadata: AuthMethodMetadata{}}, true},
		{"success of a denial", Event{Kind: AuthCSRFBlock, Metadata: DenyMetadata{}}, true},
		{"invalid outcome", Event{Kind: IAMUserDelete, Outcome: Outcome("maybe")}, true},
		{"missing metadata", Event{Kind: AuthLogin}, true},
		{"wrong metadata type", Event{Kind: AuthLogin, Metadata: DenyMetadata{}}, true},
		{"pointer metadata", Event{Kind: AuthLogin, Metadata: &AuthMethodMetadata{}}, true},
		{"metadata on kind without metadata", Event{Kind: IAMUserDelete, Metadata: DenyMetadata{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKnown(t *testing.T) {
	assert.True(t, Known(AuditLifecycleStart))
	assert.False(t, Known(Kind("audit.lifecycle/pause")))
}

func TestKnownReasonsExcludeEmpty(t *testing.T) {
	assert.False(t, slices.Contains(knownReasons(), ReasonNone))
}

func TestPermissionNames(t *testing.T) {
	assert.Equal(t, []string{}, PermissionNames(0))
	assert.Equal(t, []string{"run_tasks", "manage_users"}, PermissionNames(db.CanRunProjectTasks|db.CanManageProjectUsers))
	assert.Equal(t, []string{"run_tasks", "update_project", "manage_resources", "manage_users"}, PermissionNames(db.ProjectOwner.GetPermissions()))
}

func TestTruncateName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"short", "alice", 64, "alice"},
		{"ascii cut", "abcdef", 3, "abc"},
		{"never splits a rune", "ab€", 4, "ab"},
		{"invalid utf8 replaced", "a\xffb", 64, "a�b"},
		{"nul removed", "a\x00b", 64, "ab"},
		{"nul only", "\x00", 64, ""},
		{"nul at the cut boundary", "abc\x00def", 4, "abcd"},
		{"controls removed, tab kept", "a\x01\tb\x1f\x7fc\n", 64, "a\tb\x7fc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TruncateName(tt.in, tt.max))
		})
	}
}

func TestCatalogPartialReasons(t *testing.T) {
	for _, entry := range Catalog() {
		t.Run(string(entry.Kind), func(t *testing.T) {
			for _, reason := range entry.Partial {
				assert.Contains(t, knownReasons(), reason)
				assert.NotContains(t, entry.Reasons, reason, "a reason is either a failure or a partial success")
			}
			if len(entry.Partial) > 0 {
				assert.NotEqual(t, TypeDenied, entry.Type)
			}
		})
	}
}

func TestValidatePartialSuccess(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{"partial reason on success", Event{Kind: ResourceEnvironmentCreate, Reason: ReasonSecretFailed, Metadata: EnvironmentMetadata{Partial: true}}, false},
		{"partial reason as a failure", Event{Kind: ResourceEnvironmentCreate, Outcome: OutcomeFailure, Reason: ReasonSecretFailed, Metadata: EnvironmentMetadata{}}, true},
		{"partial reason of another kind", Event{Kind: ResourceInventoryCreate, Reason: ReasonSecretFailed}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceTarget(t *testing.T) {
	assert.Equal(t, &Target{Type: TargetInventory, ID: "5", Name: "prod"}, ResourceTarget(TargetInventory, 5, "prod"))
}
