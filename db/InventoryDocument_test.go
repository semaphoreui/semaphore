package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventoryVariablesText_IsStable(t *testing.T) {
	first := map[string]any{}
	first["b"] = map[string]any{"q": 2, "p": 1}
	first["a"] = "text"
	second := map[string]any{}
	second["a"] = "text"
	second["b"] = map[string]any{"p": 1, "q": 2}

	firstText, err := InventoryVariablesText(first)
	require.NoError(t, err)
	secondText, err := InventoryVariablesText(second)
	require.NoError(t, err)

	assert.Equal(t, "a: text\nb:\n    p: 1\n    q: 2\n", firstText)
	assert.Equal(t, firstText, secondText)
	_, err = ParseInventoryVariables(firstText)
	assert.NoError(t, err)
	empty, err := InventoryVariablesText(map[string]any{})
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestInventoryVariablesText_KeepsVault(t *testing.T) {
	vars := map[string]any{"pw": map[string]any{"__ansible_vault": "$ANSIBLE_VAULT;1.1;AES256\n61\n"}}

	text, err := InventoryVariablesText(vars)

	require.NoError(t, err)
	assert.Contains(t, text, "pw: !vault |")
	parsed, err := ParseInventoryVariables(text)
	require.NoError(t, err)
	assert.Equal(t, "$ANSIBLE_VAULT;1.1;AES256\n61\n", parsed["pw"])
}

func TestInventoryVariablesText_KeepsFloats(t *testing.T) {
	vars := map[string]any{"big": 2000000.0, "small": 0.00001, "one": 1.0, "ratio": 3.5, "negative": -0.5, "list": []any{1.0}}

	text, err := InventoryVariablesText(vars)

	require.NoError(t, err)
	assert.Equal(t, "big: 2000000.0\nlist:\n    - 1.0\nnegative: -0.5\none: 1.0\nratio: 3.5\nsmall: 0.00001\n", text)
	again, err := InventoryVariablesText(vars)
	require.NoError(t, err)
	assert.Equal(t, text, again)
	parsed, err := ParseInventoryVariables(text)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"big": 2000000.0, "small": 0.00001, "one": 1.0, "ratio": 3.5, "negative": -0.5, "list": []any{1.0}}, parsed)
	assert.Equal(t, 1.0, vars["one"])
}

func TestInventoryVariablesText_KeepsUnsafe(t *testing.T) {
	vars := map[string]any{"tpl": map[string]any{"__ansible_unsafe": "{{ not_a_template }}"}, "count": map[string]any{"__ansible_unsafe": "10"}}

	text, err := InventoryVariablesText(vars)

	require.NoError(t, err)
	assert.Equal(t, "count: \"10\"\ntpl: !unsafe \"{{ not_a_template }}\"\n", text)
}

func TestInventoryVariablesText_KeepsAMappingThatOnlyLooksLikeAWrapper(t *testing.T) {
	vars := map[string]any{"config": map[string]any{"__ansible_unsafe": false}, "limits": map[string]any{"__ansible_vault": 5}}

	text, err := InventoryVariablesText(vars)

	require.NoError(t, err)
	assert.Equal(t, "config:\n    __ansible_unsafe: false\nlimits:\n    __ansible_vault: 5\n", text)
}

func TestInventoryVariablesText_KeepsTopLevelVaultKey(t *testing.T) {
	text, err := InventoryVariablesText(map[string]any{"__ansible_vault": "x"})

	require.NoError(t, err)
	assert.Equal(t, "__ansible_vault: x\n", text)
}

func webDocument() InventoryDocument {
	return InventoryDocument{
		Variables: map[string]any{"env": "prod"},
		Hosts: []InventoryDocumentHost{
			{Name: "web2"},
			{Name: "web1", Variables: map[string]any{"port": 8080}},
		},
		Groups: []InventoryDocumentGroup{
			{Name: "web", Hosts: []string{"web2", "web1"}, Children: []string{"canary"}},
			{Name: "canary", Variables: map[string]any{"weight": 10}, Hosts: []string{"web2"}},
		},
	}
}

func TestPlanInventorySync_FirstImport(t *testing.T) {
	source := 7

	plan, err := PlanInventorySync(source, InventoryState{}, webDocument())

	require.NoError(t, err)
	assert.Equal(t, "env: prod\n", plan.Variables)
	assert.Equal(t, []InventoryHost{
		{Name: "web1", Variables: "port: 8080\n", Enabled: true, SourceID: &source},
		{Name: "web2", Enabled: true, SourceID: &source},
	}, plan.CreateHosts)
	assert.Equal(t, []InventoryGroup{
		{Name: "canary", Variables: "weight: 10\n", SourceID: &source},
		{Name: "web", SourceID: &source},
	}, plan.CreateGroups)
	assert.Equal(t, map[string][]string{"canary": {"web2"}, "web": {"web1", "web2"}}, plan.GroupHosts)
	assert.Equal(t, map[string][]string{"canary": {}, "web": {"canary"}}, plan.GroupChildren)
	assert.Equal(t, InventorySyncResult{HostsAdded: 2, GroupsAdded: 2}, plan.Result)
}

func TestPlanInventorySync_RepeatIsEmpty(t *testing.T) {
	source := 7
	state := InventoryState{
		Hosts: []InventoryHost{
			{ID: 1, Name: "web1", Variables: "port: 8080\n", Enabled: true, SourceID: &source},
			{ID: 2, Name: "web2", Enabled: true, SourceID: &source},
		},
		Groups: []InventoryGroup{
			{ID: 10, Name: "canary", Variables: "weight: 10\n", SourceID: &source},
			{ID: 11, Name: "web", SourceID: &source},
		},
		Members:  []InventoryGroupHost{{GroupID: 10, HostID: 2}, {GroupID: 11, HostID: 2}, {GroupID: 11, HostID: 1}},
		Children: map[int][]int{11: {10}},
	}

	plan, err := PlanInventorySync(source, state, webDocument())

	require.NoError(t, err)
	assert.False(t, plan.Result.Changed())
	assert.Empty(t, plan.CreateHosts)
	assert.Empty(t, plan.UpdateHosts)
	assert.Empty(t, plan.DeleteHosts)
	assert.Empty(t, plan.CreateGroups)
	assert.Empty(t, plan.UpdateGroups)
	assert.Empty(t, plan.DeleteGroups)
}

func TestPlanInventorySync_RemovesMissing(t *testing.T) {
	source := 7
	state := InventoryState{
		Hosts: []InventoryHost{
			{ID: 1, Name: "web1", SourceID: &source},
			{ID: 2, Name: "old1", SourceID: &source},
		},
		Groups:  []InventoryGroup{{ID: 10, Name: "old", SourceID: &source}},
		Members: []InventoryGroupHost{{GroupID: 10, HostID: 2}},
	}
	doc := InventoryDocument{Hosts: []InventoryDocumentHost{{Name: "web1"}}}

	plan, err := PlanInventorySync(source, state, doc)

	require.NoError(t, err)
	assert.Equal(t, []int{2}, plan.DeleteHosts)
	assert.Equal(t, []int{10}, plan.DeleteGroups)
	assert.Equal(t, InventorySyncResult{HostsRemoved: 1, GroupsRemoved: 1}, plan.Result)
}

func TestPlanInventorySync_AdoptsManual(t *testing.T) {
	source := 7
	state := InventoryState{
		Hosts: []InventoryHost{
			{ID: 1, Name: "web1", Description: "rack 4", Enabled: false, Variables: "port: 80\n"},
			{ID: 2, Name: "db1", Enabled: true},
		},
		Groups:  []InventoryGroup{{ID: 10, Name: "web", Description: "front"}},
		Members: []InventoryGroupHost{{GroupID: 10, HostID: 2}},
	}
	doc := InventoryDocument{
		Hosts:  []InventoryDocumentHost{{Name: "web1", Variables: map[string]any{"port": 8080}}},
		Groups: []InventoryDocumentGroup{{Name: "web", Hosts: []string{"web1"}}},
	}

	plan, err := PlanInventorySync(source, state, doc)

	require.NoError(t, err)
	assert.Equal(t, []InventoryHost{
		{ID: 1, Name: "web1", Description: "rack 4", Enabled: false, Variables: "port: 8080\n", SourceID: &source},
	}, plan.UpdateHosts)
	assert.Equal(t, []InventoryGroup{{ID: 10, Name: "web", Description: "front", SourceID: &source}}, plan.UpdateGroups)
	assert.Equal(t, map[string][]string{"web": {"web1"}}, plan.GroupHosts)
	assert.Equal(t, InventorySyncResult{HostsChanged: 1, GroupsChanged: 1, Adopted: 2}, plan.Result)
}

func TestPlanInventorySync_SkipsForeignOwner(t *testing.T) {
	source, other := 7, 8
	state := InventoryState{
		Hosts: []InventoryHost{{ID: 1, Name: "web1", Variables: "a: 1\n", SourceID: &other}},
	}
	doc := InventoryDocument{
		Hosts:  []InventoryDocumentHost{{Name: "web2"}, {Name: "web1", Variables: map[string]any{"a": 2}}},
		Groups: []InventoryDocumentGroup{{Name: "app", Hosts: []string{"web1", "web2"}}},
	}

	plan, err := PlanInventorySync(source, state, doc)

	require.NoError(t, err)
	assert.Empty(t, plan.UpdateHosts)
	assert.Equal(t, []InventoryHost{{Name: "web2", Enabled: true, SourceID: &source}}, plan.CreateHosts)
	assert.Equal(t, map[string][]string{"app": {"web1", "web2"}}, plan.GroupHosts)
	assert.Equal(t, []InventorySyncSkip{{Kind: "host", Name: "web1", Reason: "owned by another source"}}, plan.Result.Skipped)
	assert.Equal(t, 1, plan.Result.SkippedCount)
}

func TestPlanInventorySync_SkipsInvalidName(t *testing.T) {
	source := 7
	doc := InventoryDocument{
		Hosts: []InventoryDocumentHost{{Name: "web1"}, {Name: "fe80::1"}},
		Groups: []InventoryDocumentGroup{
			{Name: "all", Hosts: []string{"web1"}},
			{Name: "web", Hosts: []string{"web1", "fe80::1"}},
		},
	}

	plan, err := PlanInventorySync(source, InventoryState{}, doc)

	require.NoError(t, err)
	assert.Equal(t, []InventoryHost{{Name: "web1", Enabled: true, SourceID: &source}}, plan.CreateHosts)
	assert.Equal(t, []InventoryGroup{{Name: "web", SourceID: &source}}, plan.CreateGroups)
	assert.Equal(t, map[string][]string{"web": {"web1"}}, plan.GroupHosts)
	assert.Equal(t, []InventorySyncSkip{
		{Kind: "host", Name: "fe80::1", Reason: "invalid name"},
		{Kind: "group", Name: "all", Reason: "invalid name"},
	}, plan.Result.Skipped)
	assert.Equal(t, 2, plan.Result.SkippedCount)
}

func TestPlanInventorySync_LeavesOthersAlone(t *testing.T) {
	source, other := 7, 8
	state := InventoryState{
		Hosts: []InventoryHost{
			{ID: 1, Name: "manual1"},
			{ID: 2, Name: "foreign1", SourceID: &other},
			{ID: 3, Name: "web1", SourceID: &source},
		},
		Groups:  []InventoryGroup{{ID: 10, Name: "ops"}},
		Members: []InventoryGroupHost{{GroupID: 10, HostID: 1}, {GroupID: 10, HostID: 3}},
	}
	doc := InventoryDocument{Hosts: []InventoryDocumentHost{{Name: "web1"}}}

	plan, err := PlanInventorySync(source, state, doc)

	require.NoError(t, err)
	assert.False(t, plan.Result.Changed())
	assert.Empty(t, plan.CreateHosts)
	assert.Empty(t, plan.UpdateHosts)
	assert.Empty(t, plan.DeleteHosts)
	assert.Empty(t, plan.CreateGroups)
	assert.Empty(t, plan.UpdateGroups)
	assert.Empty(t, plan.DeleteGroups)
	assert.Empty(t, plan.GroupHosts)
}

func TestPlanInventorySync_SkippedGroupDropsLinks(t *testing.T) {
	source, other := 7, 8
	state := InventoryState{Groups: []InventoryGroup{{ID: 10, Name: "web", SourceID: &other}}}
	doc := InventoryDocument{
		Hosts: []InventoryDocumentHost{{Name: "web1"}},
		Groups: []InventoryDocumentGroup{
			{Name: "web", Variables: map[string]any{"a": 1}, Hosts: []string{"web1"}},
			{Name: "app", Children: []string{"web"}},
		},
	}

	plan, err := PlanInventorySync(source, state, doc)

	require.NoError(t, err)
	assert.Equal(t, []InventoryHost{{Name: "web1", Enabled: true, SourceID: &source}}, plan.CreateHosts)
	assert.Empty(t, plan.UpdateGroups)
	assert.NotContains(t, plan.GroupHosts, "web")
	assert.Equal(t, map[string][]string{"app": {"web"}}, plan.GroupChildren)
	assert.Equal(t, []InventorySyncSkip{{Kind: "group", Name: "web", Reason: "owned by another source"}}, plan.Result.Skipped)
}

func TestPlanInventorySync_RejectsCycle(t *testing.T) {
	source := 7
	state := InventoryState{
		Groups:   []InventoryGroup{{ID: 1, Name: "a"}, {ID: 2, Name: "b", SourceID: &source}},
		Children: map[int][]int{1: {2}},
	}
	doc := InventoryDocument{
		Hosts:  []InventoryDocumentHost{{Name: "web1"}},
		Groups: []InventoryDocumentGroup{{Name: "b", Children: []string{"a"}}},
	}

	plan, err := PlanInventorySync(source, state, doc)

	assert.EqualError(t, err, "group cycle: a, b, a")
	assert.Equal(t, InventorySyncPlan{}, plan)
}

func TestPlanInventorySync_RejectsEmpty(t *testing.T) {
	source := 7
	state := InventoryState{Hosts: []InventoryHost{{ID: 1, Name: "web1", SourceID: &source}}}
	doc := InventoryDocument{Groups: []InventoryDocumentGroup{{Name: "web"}}}

	plan, err := PlanInventorySync(source, state, doc)

	assert.ErrorIs(t, err, ErrInventorySourceEmpty)
	assert.Equal(t, InventorySyncPlan{}, plan)
}

func TestPlanInventorySync_RejectsOnlyInvalidNames(t *testing.T) {
	source := 7
	state := InventoryState{Hosts: []InventoryHost{{ID: 1, Name: "web1", SourceID: &source}}}
	doc := InventoryDocument{Hosts: []InventoryDocumentHost{{Name: "fe80::1"}}}

	plan, err := PlanInventorySync(source, state, doc)

	assert.ErrorIs(t, err, ErrInventorySourceEmpty)
	assert.Equal(t, InventorySyncPlan{}, plan)
}

func TestPlanInventorySync_RejectsOnlyForeignHosts(t *testing.T) {
	source, other := 7, 8
	state := InventoryState{Hosts: []InventoryHost{
		{ID: 1, Name: "old", SourceID: &source},
		{ID: 2, Name: "shared", SourceID: &other},
	}}
	doc := InventoryDocument{Hosts: []InventoryDocumentHost{{Name: "shared"}}}

	plan, err := PlanInventorySync(source, state, doc)

	assert.ErrorIs(t, err, ErrInventorySourceEmpty)
	assert.Equal(t, InventorySyncPlan{}, plan)
}

func TestPlanInventorySync_UnchangedOwnHostIsEnough(t *testing.T) {
	source, other := 7, 8
	state := InventoryState{Hosts: []InventoryHost{
		{ID: 1, Name: "old", SourceID: &source},
		{ID: 2, Name: "kept", SourceID: &source},
		{ID: 3, Name: "shared", SourceID: &other},
	}}
	doc := InventoryDocument{Hosts: []InventoryDocumentHost{{Name: "kept"}, {Name: "shared"}}}

	plan, err := PlanInventorySync(source, state, doc)

	require.NoError(t, err)
	assert.Equal(t, []int{1}, plan.DeleteHosts)
	assert.Empty(t, plan.UpdateHosts)
}
