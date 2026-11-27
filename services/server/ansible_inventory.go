package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// maxUngroupedCopyBytes limits the size of the variables of the group ungrouped times its hosts
const maxUngroupedCopyBytes = 16 << 20

// ParseAnsibleInventory reads the output of ansible-inventory --list --export
func ParseAnsibleInventory(output []byte) (db.InventoryDocument, error) {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(output, &parsed); err != nil {
		return db.InventoryDocument{}, err
	}

	var meta struct {
		Hostvars map[string]map[string]any `json:"hostvars"`
	}
	if raw, ok := parsed["_meta"]; ok {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&meta); err != nil {
			return db.InventoryDocument{}, err
		}
	}
	for _, vars := range meta.Hostvars {
		exactNumbers(vars)
	}

	doc := db.InventoryDocument{}
	childNames := []string{}
	hostNames := map[string]bool{}
	var ungrouped map[string]any
	var ungroupedHosts []string
	for name := range meta.Hostvars {
		hostNames[name] = true
	}

	for name, raw := range parsed {
		if name == "_meta" {
			continue
		}

		var group struct {
			Hosts    []string       `json:"hosts"`
			Children []string       `json:"children"`
			Vars     map[string]any `json:"vars"`
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&group); err != nil {
			return db.InventoryDocument{}, err
		}
		exactNumbers(group.Vars)

		childNames = append(childNames, group.Children...)

		for _, host := range group.Hosts {
			hostNames[host] = true
		}

		if name == "all" {
			doc.Variables = group.Vars
			continue
		}
		if name == "ungrouped" {
			ungrouped, ungroupedHosts = group.Vars, group.Hosts
			continue
		}

		hosts := slices.Clone(group.Hosts)
		slices.Sort(hosts)
		children := slices.Clone(group.Children)
		slices.Sort(children)
		doc.Groups = append(doc.Groups, db.InventoryDocumentGroup{
			Name:      name,
			Variables: group.Vars,
			Hosts:     slices.Compact(hosts),
			Children:  slices.Compact(children),
		})
	}

	listed := map[string]bool{}
	for _, group := range doc.Groups {
		listed[group.Name] = true
	}
	for _, child := range childNames {
		if !listed[child] && child != "all" && child != "ungrouped" {
			listed[child] = true
			doc.Groups = append(doc.Groups, db.InventoryDocumentGroup{Name: child})
		}
	}

	// The managed inventory has no ungrouped group, its variables go to its hosts
	hostVars := map[string]map[string]any{}
	maps.Copy(hostVars, meta.Hostvars)
	if len(ungrouped) > 0 {
		// A small file can ask for a copy of large variables to thousands of hosts
		size, err := json.Marshal(ungrouped)
		if err != nil {
			return db.InventoryDocument{}, err
		}
		// int is 32 bits on some release targets
		if int64(len(size))*int64(len(ungroupedHosts)) > maxUngroupedCopyBytes {
			return db.InventoryDocument{}, fmt.Errorf("the variables of the group ungrouped are too large to copy to its %d hosts", len(ungroupedHosts))
		}
		for _, name := range ungroupedHosts {
			vars := maps.Clone(ungrouped)
			maps.Copy(vars, hostVars[name])
			hostVars[name] = vars
		}
	}

	for name := range hostNames {
		doc.Hosts = append(doc.Hosts, db.InventoryDocumentHost{Name: name, Variables: hostVars[name]})
	}

	slices.SortFunc(doc.Hosts, func(a, b db.InventoryDocumentHost) int { return strings.Compare(a.Name, b.Name) })
	slices.SortFunc(doc.Groups, func(a, b db.InventoryDocumentGroup) int { return strings.Compare(a.Name, b.Name) })

	return doc, nil
}

// exactNumbers replaces each json.Number in place with an integer when it is one, else with a float
func exactNumbers(value any) any {
	switch value := value.(type) {
	case map[string]any:
		for key, item := range value {
			value[key] = exactNumbers(item)
		}
	case []any:
		for i, item := range value {
			value[i] = exactNumbers(item)
		}
	case json.Number:
		if n, err := value.Int64(); err == nil {
			return n
		}
		if n, err := strconv.ParseUint(value.String(), 10, 64); err == nil {
			return n
		}
		// A float would round an integer past 64 bits
		if n, ok := new(big.Int).SetString(value.String(), 10); ok {
			return n
		}
		n, _ := value.Float64()
		return n
	}
	return value
}
