package tasks

import (
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"gopkg.in/yaml.v3"
)

// variableNode validates the variables text and keeps its keys and scalars as written
func variableNode(text string) (*yaml.Node, error) {
	_, err := db.ParseInventoryVariables(text)
	if err != nil {
		return nil, err
	}

	var doc yaml.Node
	err = yaml.Unmarshal([]byte(text), &doc)
	if err != nil {
		return nil, err
	}

	if len(doc.Content) == 1 && doc.Content[0].Kind == yaml.MappingNode {
		return doc.Content[0], nil
	}

	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}, nil
}

// buildManagedInventory renders the inventory and source variables, enabled hosts and groups as a static YAML inventory
func buildManagedInventory(variables string, sourceVariables []string, hosts []db.InventoryHost, groups []db.InventoryGroup, members []db.InventoryGroupHost) (string, error) {
	inventoryVars, err := variableNode(variables)
	if err != nil {
		return "", fmt.Errorf("inventory variables: %w", err)
	}

	keys := map[string]bool{}
	for i := 0; i < len(inventoryVars.Content); i += 2 {
		keys[inventoryVars.Content[i].Value] = true
	}
	for _, text := range sourceVariables {
		sourceVars, err := variableNode(text)
		if err != nil {
			return "", fmt.Errorf("source variables: %w", err)
		}
		for i := 0; i < len(sourceVars.Content); i += 2 {
			if !keys[sourceVars.Content[i].Value] {
				keys[sourceVars.Content[i].Value] = true
				inventoryVars.Content = append(inventoryVars.Content, sourceVars.Content[i], sourceVars.Content[i+1])
			}
		}
	}

	hostsMap := map[string]any{}
	for _, host := range hosts {
		hostVars, err := variableNode(host.Variables)
		if err != nil {
			return "", fmt.Errorf("host %s: %w", host.Name, err)
		}
		hostsMap[host.Name] = hostVars
	}

	all := map[string]any{"hosts": hostsMap}
	if len(inventoryVars.Content) > 0 {
		all["vars"] = inventoryVars
	}

	inventory := map[string]any{"all": all}

	hostNames := map[int]string{}
	for _, host := range hosts {
		hostNames[host.ID] = host.Name
	}
	groupNames := map[int]string{}
	children := map[int][]int{}
	for _, group := range groups {
		groupNames[group.ID] = group.Name
		children[group.ID] = group.Children
	}

	if cycle := db.InventoryGroupCycle(children); cycle != nil {
		names := make([]string, 0, len(cycle)+1)
		for _, id := range append(cycle, cycle[0]) {
			names = append(names, groupNames[id])
		}
		return "", fmt.Errorf("group cycle: %s", strings.Join(names, ", "))
	}

	groupHosts := map[int]map[string]any{}
	for _, member := range members {
		name, ok := hostNames[member.HostID]
		if !ok {
			continue
		}
		if groupHosts[member.GroupID] == nil {
			groupHosts[member.GroupID] = map[string]any{}
		}
		groupHosts[member.GroupID][name] = map[string]any{}
	}

	// Groups go under all, a top-level key named plugin would stop Ansible from reading the file as an inventory
	allChildren := map[string]any{}
	for _, group := range groups {
		groupVars, err := variableNode(group.Variables)
		if err != nil {
			return "", fmt.Errorf("group %s: %w", group.Name, err)
		}

		entry := map[string]any{}
		if len(groupVars.Content) > 0 {
			entry["vars"] = groupVars
		}
		if len(groupHosts[group.ID]) > 0 {
			entry["hosts"] = groupHosts[group.ID]
		}
		if len(group.Children) > 0 {
			groupChildren := map[string]any{}
			for _, id := range group.Children {
				if name, ok := groupNames[id]; ok {
					groupChildren[name] = map[string]any{}
				}
			}
			entry["children"] = groupChildren
		}
		allChildren[group.Name] = entry
	}
	if len(allChildren) > 0 {
		all["children"] = allChildren
	}

	text, err := yaml.Marshal(inventory)
	if err != nil {
		return "", err
	}

	return string(text), nil
}
