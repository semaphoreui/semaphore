package db

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// InventoryDocument is the hosts and groups a source produced, Variables belong to group all
// Numbers in Variables are int64, uint64 or float64, never strings
type InventoryDocument struct {
	Variables map[string]any
	Hosts     []InventoryDocumentHost
	Groups    []InventoryDocumentGroup
}

type InventoryDocumentHost struct {
	Name      string
	Variables map[string]any
}

type InventoryDocumentGroup struct {
	Name      string
	Variables map[string]any
	Hosts     []string
	Children  []string
}

// InventoryState is an inventory before a sync, Children maps a group to its child groups
type InventoryState struct {
	Hosts    []InventoryHost
	Groups   []InventoryGroup
	Members  []InventoryGroupHost
	Children map[int][]int
}

// InventoryContent is an inventory at one moment: the hosts, the groups with their children, the sources by id
type InventoryContent struct {
	Hosts   []InventoryHost
	Groups  []InventoryGroup
	Members []InventoryGroupHost
	Sources []InventorySource
}

// InventorySyncPlan is what a sync writes, GroupHosts and GroupChildren hold every group of the source by name
type InventorySyncPlan struct {
	Variables     string
	CreateHosts   []InventoryHost
	UpdateHosts   []InventoryHost
	DeleteHosts   []int
	CreateGroups  []InventoryGroup
	UpdateGroups  []InventoryGroup
	DeleteGroups  []int
	GroupHosts    map[string][]string
	GroupChildren map[string][]string
	Result        InventorySyncResult
}

var ErrInventorySourceEmpty = errors.New("source returned no hosts")

// InventoryVariablesText writes variables as YAML with sorted keys, an empty map gives an empty text
func InventoryVariablesText(vars map[string]any) (string, error) {
	if len(vars) == 0 {
		return "", nil
	}

	var node yaml.Node
	err := node.Encode(floatScalars(vars))
	if err != nil {
		return "", err
	}

	// ansible-inventory writes a vault or an unsafe value as a mapping with the single key __ansible_vault or __ansible_unsafe
	// The top mapping is not a value
	pending := slices.Clone(node.Content)
	for len(pending) > 0 {
		next := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if next.Kind == yaml.MappingNode && len(next.Content) == 2 && next.Content[1].Kind == yaml.ScalarNode && next.Content[1].Tag == "!!str" {
			switch next.Content[0].Value {
			case "__ansible_vault":
				*next = yaml.Node{Kind: yaml.ScalarNode, Tag: "!vault", Style: yaml.LiteralStyle, Value: next.Content[1].Value}
				continue
			case "__ansible_unsafe":
				// Ansible 2.19 and later read an !unsafe scalar without its quotes, "10" would become a number
				// Only a value Ansible could template needs the tag
				value := next.Content[1].Value
				*next = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Style: yaml.DoubleQuotedStyle, Value: value}
				if strings.HasPrefix(value, "#jinja2:") || strings.Contains(value, "{{") || strings.Contains(value, "{%") || strings.Contains(value, "{#") {
					next.Tag = "!unsafe"
				}
				continue
			}
		}
		pending = append(pending, next.Content...)
	}

	text, err := yaml.Marshal(&node)
	return string(text), err
}

// floatScalars copies the value with each float64 as a decimal with a dot, PyYAML reads 2e+06 as a string and 1 as an int
// An integer past 64 bits is written whole
func floatScalars(value any) any {
	switch value := value.(type) {
	case map[string]any:
		copied := make(map[string]any, len(value))
		for key, item := range value {
			copied[key] = floatScalars(item)
		}
		return copied
	case []any:
		copied := make([]any, len(value))
		for i, item := range value {
			copied[i] = floatScalars(item)
		}
		return copied
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return value
		}
		text := strconv.FormatFloat(value, 'f', -1, 64)
		if !strings.Contains(text, ".") {
			text += ".0"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: text}
	case *big.Int:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: value.String()}
	}
	return value
}

// PlanInventorySync compares the document of a source with the inventory and lists the changes
func PlanInventorySync(sourceID int, state InventoryState, doc InventoryDocument) (InventorySyncPlan, error) {
	variables, err := InventoryVariablesText(doc.Variables)
	if err != nil {
		return InventorySyncPlan{}, err
	}

	plan := InventorySyncPlan{
		Variables:     variables,
		GroupHosts:    map[string][]string{},
		GroupChildren: map[string][]string{},
	}

	hostExists, accepted, err := plan.planHosts(sourceID, state.Hosts, doc.Hosts)
	if err != nil {
		return InventorySyncPlan{}, err
	}
	// A document that gives the source no host would remove every host it has: an unparsed file, a wrong path
	if accepted == 0 {
		return InventorySyncPlan{}, ErrInventorySourceEmpty
	}

	err = plan.planGroups(sourceID, state, doc.Groups, hostExists)
	if err != nil {
		return InventorySyncPlan{}, err
	}

	plan.Result.SkippedCount = len(plan.Result.Skipped)
	return plan, nil
}

// planHosts fills the host lists and returns the host names that exist after the sync and how many document hosts the source took
func (p *InventorySyncPlan) planHosts(sourceID int, hosts []InventoryHost, docHosts []InventoryDocumentHost) (map[string]bool, int, error) {
	inDocument := map[string]bool{}
	for _, docHost := range docHosts {
		inDocument[docHost.Name] = true
	}

	existing := map[string]InventoryHost{}
	exists := map[string]bool{}
	for _, host := range hosts {
		if host.SourceID != nil && *host.SourceID == sourceID && !inDocument[host.Name] {
			p.DeleteHosts = append(p.DeleteHosts, host.ID)
			continue
		}
		existing[host.Name] = host
		exists[host.Name] = true
	}
	p.Result.HostsRemoved = len(p.DeleteHosts)

	sorted := slices.SortedFunc(slices.Values(docHosts), func(a, b InventoryDocumentHost) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, docHost := range sorted {
		vars, err := InventoryVariablesText(docHost.Variables)
		if err != nil {
			return nil, 0, err
		}

		host, found := existing[docHost.Name]
		switch {
		case InventoryHost{Name: docHost.Name}.Validate() != nil:
			p.Result.Skipped = append(p.Result.Skipped, InventorySyncSkip{Kind: "host", Name: docHost.Name, Reason: "invalid name"})
		case !found:
			p.CreateHosts = append(p.CreateHosts, InventoryHost{Name: docHost.Name, Variables: vars, Enabled: true, SourceID: &sourceID})
			p.Result.HostsAdded++
			exists[docHost.Name] = true
		case host.SourceID == nil:
			host.Variables = vars
			host.SourceID = &sourceID
			p.UpdateHosts = append(p.UpdateHosts, host)
			p.Result.HostsChanged++
			p.Result.Adopted++
		case *host.SourceID != sourceID:
			p.Result.Skipped = append(p.Result.Skipped, InventorySyncSkip{Kind: "host", Name: docHost.Name, Reason: "owned by another source"})
		case host.Variables != vars:
			host.Variables = vars
			p.UpdateHosts = append(p.UpdateHosts, host)
			p.Result.HostsChanged++
		}
	}

	accepted := len(docHosts) - len(p.Result.Skipped)
	return exists, accepted, nil
}

// planGroups fills the group lists and the links of the source groups, a cycle of groups is an error
func (p *InventorySyncPlan) planGroups(sourceID int, state InventoryState, docGroups []InventoryDocumentGroup, hostExists map[string]bool) error {
	inDocument := map[string]bool{}
	for _, docGroup := range docGroups {
		inDocument[docGroup.Name] = true
	}

	hostNames := map[int]string{}
	for _, host := range state.Hosts {
		hostNames[host.ID] = host.Name
	}
	groupNames := map[int]string{}
	for _, group := range state.Groups {
		groupNames[group.ID] = group.Name
	}

	oldHosts := map[int][]string{}
	for _, member := range state.Members {
		oldHosts[member.GroupID] = append(oldHosts[member.GroupID], hostNames[member.HostID])
	}
	oldChildren := map[int][]string{}
	for id, children := range state.Children {
		for _, child := range children {
			oldChildren[id] = append(oldChildren[id], groupNames[child])
		}
		slices.Sort(oldChildren[id])
	}
	for _, names := range oldHosts {
		slices.Sort(names)
	}

	existing := map[string]InventoryGroup{}
	exists := map[string]bool{}
	for _, group := range state.Groups {
		if group.SourceID != nil && *group.SourceID == sourceID && !inDocument[group.Name] {
			p.DeleteGroups = append(p.DeleteGroups, group.ID)
			continue
		}
		existing[group.Name] = group
		exists[group.Name] = true
	}
	p.Result.GroupsRemoved = len(p.DeleteGroups)

	for _, docGroup := range docGroups {
		if (InventoryGroup{Name: docGroup.Name}).Validate() == nil {
			exists[docGroup.Name] = true
		}
	}

	sorted := slices.SortedFunc(slices.Values(docGroups), func(a, b InventoryDocumentGroup) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, docGroup := range sorted {
		vars, err := InventoryVariablesText(docGroup.Variables)
		if err != nil {
			return err
		}

		group, found := existing[docGroup.Name]
		if (InventoryGroup{Name: docGroup.Name}).Validate() != nil {
			p.Result.Skipped = append(p.Result.Skipped, InventorySyncSkip{Kind: "group", Name: docGroup.Name, Reason: "invalid name"})
			continue
		}
		if found && group.SourceID != nil && *group.SourceID != sourceID {
			p.Result.Skipped = append(p.Result.Skipped, InventorySyncSkip{Kind: "group", Name: docGroup.Name, Reason: "owned by another source"})
			continue
		}

		hosts := []string{}
		for _, name := range docGroup.Hosts {
			if hostExists[name] {
				hosts = append(hosts, name)
			}
		}
		slices.Sort(hosts)
		hosts = slices.Compact(hosts)
		p.GroupHosts[docGroup.Name] = hosts

		children := []string{}
		for _, name := range docGroup.Children {
			if exists[name] {
				children = append(children, name)
			}
		}
		slices.Sort(children)
		children = slices.Compact(children)
		p.GroupChildren[docGroup.Name] = children

		switch {
		case !found:
			p.CreateGroups = append(p.CreateGroups, InventoryGroup{Name: docGroup.Name, Variables: vars, SourceID: &sourceID})
			p.Result.GroupsAdded++
		case group.SourceID == nil:
			group.Variables = vars
			group.SourceID = &sourceID
			p.UpdateGroups = append(p.UpdateGroups, group)
			p.Result.GroupsChanged++
			p.Result.Adopted++
		case group.Variables != vars || !slices.Equal(hosts, oldHosts[group.ID]) || !slices.Equal(children, oldChildren[group.ID]):
			group.Variables = vars
			p.UpdateGroups = append(p.UpdateGroups, group)
			p.Result.GroupsChanged++
		}
	}

	// New groups get negative ids for the cycle check only
	ids := map[string]int{}
	for _, group := range state.Groups {
		ids[group.Name] = group.ID
	}
	for i, group := range p.CreateGroups {
		ids[group.Name] = -1 - i
		groupNames[-1-i] = group.Name
	}

	links := map[int][]int{}
	for _, group := range state.Groups {
		if exists[group.Name] {
			links[group.ID] = state.Children[group.ID]
		}
	}
	for name, children := range p.GroupChildren {
		links[ids[name]] = nil
		for _, child := range children {
			links[ids[name]] = append(links[ids[name]], ids[child])
		}
	}

	if cycle := InventoryGroupCycle(links); cycle != nil {
		names := make([]string, 0, len(cycle)+1)
		for _, id := range append(cycle, cycle[0]) {
			names = append(names, groupNames[id])
		}
		return fmt.Errorf("group cycle: %s", strings.Join(names, ", "))
	}

	return nil
}
