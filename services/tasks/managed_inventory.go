package tasks

import (
	"fmt"

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

// buildManagedInventory renders the inventory variables and enabled hosts as a static YAML inventory
func buildManagedInventory(variables string, hosts []db.InventoryHost) (string, error) {
	inventoryVars, err := variableNode(variables)
	if err != nil {
		return "", fmt.Errorf("inventory variables: %w", err)
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

	text, err := yaml.Marshal(map[string]any{"all": all})
	if err != nil {
		return "", err
	}

	return string(text), nil
}
