package project

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func getEntryByName[T BackupEntry](name *string, items []T) *T {
	if name == nil {
		return nil
	}
	for _, o := range items {
		if o.GetName() == *name {
			return &o
		}
	}
	return nil
}

func verifyDuplicate[T BackupEntry](name string, items []T) error {
	n := 0
	for _, o := range items {
		if o.GetName() == name {
			n++
		}
		if n > 2 {
			return fmt.Errorf("%s is duplicate", name)
		}
	}
	return nil
}

func (e BackupSecretStorage) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupSecretStorage](e.Name, backup.SecretStorages)
}

func (e BackupSecretStorage) Restore(b *BackupDB) error {
	st := e.SecretStorage
	st.ProjectID = b.meta.ID
	newStorage, err := b.store.CreateSecretStorage(st)
	if err != nil {
		return err
	}
	b.secretStorages = append(b.secretStorages, newStorage)
	return nil
}

// Verify skips the slug format: an upgraded instance keeps old slugs like "my role", and its export must restore.
func (e BackupRole) Verify(backup *BackupFormat) error {
	role := e.Role
	role.Permissions &= db.KnownRolePermissions
	if err := db.ValidateRole(role, false); err != nil {
		return err
	}
	if n := len([]rune(role.Slug)); n == 0 || n > 100 {
		return &common_errors.ValidationError{Message: "Role slug must be 1 to 100 characters"}
	}
	if db.IsReservedRoleSlug(role.Slug) {
		return &common_errors.ValidationError{Message: "Role slug is reserved: " + role.Slug}
	}
	n := 0
	for _, other := range backup.Roles {
		if strings.EqualFold(other.Slug, role.Slug) {
			n++
		}
	}
	if n > 1 {
		return &common_errors.ValidationError{Message: "Role slug is duplicate: " + role.Slug}
	}
	return nil
}

func (e BackupRole) Restore(b *BackupDB) error {
	role := e.Role
	role.ProjectID = &b.meta.ID
	// Unknown bits are never checked today but would turn on with a future permission.
	role.Permissions &= db.KnownRolePermissions
	if b.version < 3 {
		role.Permissions = db.LiftPermissions(role.Permissions)
	}
	if b.version < 4 {
		role.Permissions |= db.CanReadProject
	}
	_, err := b.store.GetSharedRoleBySlug(role.Slug)
	if err == nil {
		taken := b.unassignedRole(role.Slug, ReasonSlugTaken)
		taken.Name, taken.Permissions = role.Name, role.Permissions
		return nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return err
	}
	created, err := b.store.CreateRole(role)
	if err != nil {
		return err
	}
	b.roles = append(b.roles, created)
	return nil
}

// unassignedRole finds or adds the report entry for a slug. Use the pointer at once, a later add may move it.
func (b *BackupDB) unassignedRole(slug, reason string) *UnassignedRole {
	for i := range b.unassigned {
		if b.unassigned[i].Slug == slug {
			return &b.unassigned[i]
		}
	}
	b.unassigned = append(b.unassigned, UnassignedRole{Slug: slug, Reason: reason, Templates: []UnassignedTemplate{}, Workflows: []UnassignedWorkflow{}})
	return &b.unassigned[len(b.unassigned)-1]
}

func (e BackupEnvironment) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupEnvironment](e.Name, backup.Environments)
}

func (e BackupEnvironment) Restore(b *BackupDB) error {
	env := e.Environment
	env.ProjectID = b.meta.ID
	newEnv, err := b.store.CreateEnvironment(env)
	if err != nil {
		return err
	}
	b.environments = append(b.environments, newEnv)
	return nil
}

func (e BackupView) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupView](e.Title, backup.Views)
}

func (e BackupView) Restore(b *BackupDB) error {
	v := e.View
	v.ProjectID = b.meta.ID
	newView, err := b.store.CreateView(v)
	if err != nil {
		return err
	}
	b.views = append(b.views, newView)
	return nil
}

func (e BackupSchedule) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupSchedule](e.Name, backup.Schedules)
}

func (e BackupSchedule) Restore(b *BackupDB) error {
	v := e.Schedule
	v.ProjectID = b.meta.ID
	v.TemplateID = nil
	v.WorkflowID = nil

	switch {
	case e.Template != nil && e.Workflow != nil:
		return fmt.Errorf("schedule runs either a template or a workflow, not both")
	case e.Template != nil:
		tpl := findEntityByName[db.Template](e.Template, b.templates)
		if tpl == nil {
			return fmt.Errorf("template does not exist in templates[].name")
		}
		v.TemplateID = &tpl.ID
	case e.Workflow != nil:
		// Workflows are restored before schedules, so b.workflows is complete.
		workflow := findEntityByName[db.WorkflowTemplate](e.Workflow, b.workflows)
		if workflow == nil {
			return fmt.Errorf("workflow does not exist in workflows[].name")
		}
		v.WorkflowID = &workflow.ID
	default:
		return fmt.Errorf("schedule has neither a template nor a workflow")
	}

	if e.CheckableRepository != nil {
		repo := findEntityByName[db.Repository](e.CheckableRepository, b.repositories)
		if repo == nil {
			return fmt.Errorf("repo does not exist in repositories[].name")
		}
		v.RepositoryID = &repo.ID
	}

	if e.TaskParams != nil {
		inv := findEntityByName[db.Inventory](e.TaskParams.InventoryName, b.inventories)
		if inv != nil {
			v.TaskParams.InventoryID = &inv.ID
		}
	}

	newSchedule, err := b.store.CreateSchedule(v)
	if err != nil {
		return err
	}

	b.schedules = append(b.schedules, newSchedule)
	return nil
}

func (e BackupAccessKey) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupAccessKey](e.Name, backup.Keys)
}

func (e BackupAccessKey) Restore(b *BackupDB) error {

	key := e.AccessKey
	key.ProjectID = &b.meta.ID

	if e.Storage != nil {
		storage := findEntityByName[db.SecretStorage](e.Storage, b.secretStorages)
		if storage == nil {
			return fmt.Errorf("secret storage does not exist in secret_storage[].name")
		}
		key.StorageID = &storage.ID
	}

	if e.SourceStorage != nil {
		sourceStorage := findEntityByName[db.SecretStorage](e.SourceStorage, b.secretStorages)
		if sourceStorage == nil {
			return fmt.Errorf("secret storage does not exist in secret_storage[].name")
		}
		key.SourceStorageID = &sourceStorage.ID
	}

	newKey, err := b.store.CreateAccessKey(key)

	if err != nil {
		return err
	}
	b.keys = append(b.keys, newKey)
	return nil
}

func (e BackupInventory) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupInventory](e.Name, backup.Inventories); err != nil {
		return err
	}
	if e.SSHKey != nil && getEntryByName[BackupAccessKey](e.SSHKey, backup.Keys) == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	}
	if e.BecomeKey != nil && getEntryByName[BackupAccessKey](e.BecomeKey, backup.Keys) == nil {
		return fmt.Errorf("BecomeKey does not exist in keys[].Name")
	}
	if len(e.Hosts) > 0 && e.Type != db.InventoryManaged {
		return fmt.Errorf("inventory %q is not managed but has hosts", e.Name)
	}
	if e.VariablesText() != "" && e.Type != db.InventoryManaged {
		return fmt.Errorf("inventory %q is not managed but has variables", e.Name)
	}
	if _, err := db.ParseInventoryVariables(e.VariablesText()); err != nil {
		return fmt.Errorf("variables of inventory %q: %w", e.Name, err)
	}
	if len(e.Sources) > 0 && e.Type != db.InventoryManaged {
		return fmt.Errorf("inventory %q is not managed but has sources", e.Name)
	}
	sources := make(map[string]bool)
	for _, s := range e.Sources {
		if sources[s.Name] {
			return fmt.Errorf("source %q of inventory %q is duplicated", s.Name, e.Name)
		}
		sources[s.Name] = true
		if s.Repository != nil && getEntryByName[BackupRepository](s.Repository, backup.Repositories) == nil {
			return fmt.Errorf("source %q of inventory %q names unknown repository %q", s.Name, e.Name, *s.Repository)
		}
		if s.Key != nil && getEntryByName[BackupAccessKey](s.Key, backup.Keys) == nil {
			return fmt.Errorf("source %q of inventory %q names unknown key %q", s.Name, e.Name, *s.Key)
		}
		source := db.InventorySource{Name: s.Name, Type: s.Type, Params: s.Params}
		if err := source.Validate(); err != nil {
			return fmt.Errorf("source %q of inventory %q: %w", s.Name, e.Name, err)
		}
		if _, err := db.ParseInventoryVariables(s.Variables); err != nil {
			return fmt.Errorf("variables of source %q of inventory %q: %w", s.Name, e.Name, err)
		}
	}
	names := make(map[string]bool)
	for i, h := range e.Hosts {
		if names[h.Name] {
			return fmt.Errorf("host %q of inventory %q is duplicated", h.Name, e.Name)
		}
		if h.Source != nil && !sources[*h.Source] {
			return fmt.Errorf("host %q of inventory %q names unknown source %q", h.Name, e.Name, *h.Source)
		}
		names[h.Name] = true
		host := db.InventoryHost{Name: h.Name, Variables: h.Variables}
		if err := host.Validate(); err != nil {
			return fmt.Errorf("host %q of inventory %q at hosts[%d]: %w", h.Name, e.Name, i, err)
		}
	}
	if len(e.Groups) > 0 && e.Type != db.InventoryManaged {
		return fmt.Errorf("inventory %q is not managed but has groups", e.Name)
	}
	groupIndex := make(map[string]int)
	for i, g := range e.Groups {
		if _, ok := groupIndex[g.Name]; ok {
			return fmt.Errorf("group %q of inventory %q is duplicated", g.Name, e.Name)
		}
		groupIndex[g.Name] = i
		if g.Source != nil && !sources[*g.Source] {
			return fmt.Errorf("group %q of inventory %q names unknown source %q", g.Name, e.Name, *g.Source)
		}
		group := db.InventoryGroup{Name: g.Name, Variables: g.Variables}
		if err := group.Validate(); err != nil {
			return fmt.Errorf("group %q of inventory %q at groups[%d]: %w", g.Name, e.Name, i, err)
		}
		listed := make(map[string]bool)
		for _, h := range g.Hosts {
			if !names[h] {
				return fmt.Errorf("group %q of inventory %q names unknown host %q", g.Name, e.Name, h)
			}
			if listed[h] {
				return fmt.Errorf("group %q of inventory %q lists host %q twice", g.Name, e.Name, h)
			}
			listed[h] = true
		}
	}
	children := make(map[int][]int)
	for i, g := range e.Groups {
		listed := make(map[string]bool)
		for _, c := range g.Children {
			child, ok := groupIndex[c]
			if !ok {
				return fmt.Errorf("group %q of inventory %q names unknown child group %q", g.Name, e.Name, c)
			}
			if listed[c] {
				return fmt.Errorf("group %q of inventory %q lists child group %q twice", g.Name, e.Name, c)
			}
			listed[c] = true
			children[i] = append(children[i], child)
		}
	}
	if cycle := db.InventoryGroupCycle(children); cycle != nil {
		return fmt.Errorf("group %q of inventory %q is its own ancestor", e.Groups[cycle[0]].Name, e.Name)
	}
	return nil
}

func (e BackupInventory) Restore(b *BackupDB) error {
	var SSHKeyID *int
	if e.SSHKey == nil {
		SSHKeyID = nil
	} else if k := findEntityByName[db.AccessKey](e.SSHKey, b.keys); k == nil {
		SSHKeyID = nil
	} else {
		SSHKeyID = &((*k).ID)
	}
	var BecomeKeyID *int
	if e.BecomeKey == nil {
		BecomeKeyID = nil
	} else if k := findEntityByName[db.AccessKey](e.BecomeKey, b.keys); k == nil {
		BecomeKeyID = nil
	} else {
		BecomeKeyID = &((*k).ID)
	}

	inv := e.Inventory
	inv.ProjectID = b.meta.ID
	inv.SSHKeyID = SSHKeyID
	inv.BecomeKeyID = BecomeKeyID

	newInventory, err := b.store.CreateInventory(inv)
	if err != nil {
		return err
	}
	b.inventories = append(b.inventories, newInventory)

	sourceIDs := make(map[string]*int)
	for _, s := range e.Sources {
		var repositoryID *int
		if repo := findEntityByName[db.Repository](s.Repository, b.repositories); repo != nil {
			repositoryID = &repo.ID
		}
		var keyID *int
		if s.Key != nil {
			if key := findEntityByName[db.AccessKey](s.Key, b.keys); key != nil {
				keyID = &key.ID
			}
		}
		source, err := b.store.CreateInventorySource(db.InventorySource{
			ProjectID:     b.meta.ID,
			InventoryID:   newInventory.ID,
			Name:          s.Name,
			Type:          s.Type,
			RepositoryID:  repositoryID,
			KeyID:         keyID,
			Params:        s.Params,
			Variables:     s.Variables,
			VariableCount: db.CountInventoryVariables(s.Variables),
		})
		if err != nil {
			return fmt.Errorf("source %q of inventory %q: %w", s.Name, e.Name, err)
		}
		sourceIDs[s.Name] = &source.ID
	}

	hostIDs := make(map[string]int)
	for _, h := range e.Hosts {
		var sourceID *int
		if h.Source != nil {
			sourceID = sourceIDs[*h.Source]
		}
		host, err := b.store.CreateInventoryHost(db.InventoryHost{
			ProjectID:   b.meta.ID,
			InventoryID: newInventory.ID,
			Name:        h.Name,
			Description: h.Description,
			Enabled:     h.Enabled,
			Variables:   h.Variables,
			SourceID:    sourceID,
		}, nil)
		if err != nil {
			return fmt.Errorf("host %q of inventory %q: %w", h.Name, e.Name, err)
		}
		hostIDs[h.Name] = host.ID
	}

	groups := make(map[string]db.InventoryGroup)
	for _, g := range e.Groups {
		var sourceID *int
		if g.Source != nil {
			sourceID = sourceIDs[*g.Source]
		}
		group, err := b.store.CreateInventoryGroup(db.InventoryGroup{
			ProjectID:   b.meta.ID,
			InventoryID: newInventory.ID,
			Name:        g.Name,
			Description: g.Description,
			Variables:   g.Variables,
			SourceID:    sourceID,
		}, nil)
		if err != nil {
			return fmt.Errorf("group %q of inventory %q: %w", g.Name, e.Name, err)
		}
		groups[g.Name] = group
	}

	for _, g := range e.Groups {
		group := groups[g.Name]
		if len(g.Children) > 0 {
			var children []int
			for _, c := range g.Children {
				children = append(children, groups[c].ID)
			}
			if _, _, err = b.store.UpdateInventoryGroup(group, children, nil); err != nil {
				return fmt.Errorf("group %q of inventory %q: %w", g.Name, e.Name, err)
			}
		}

		var members []int
		for _, h := range g.Hosts {
			members = append(members, hostIDs[h])
		}
		// Batches keep the host id list of one call within database parameter limits
		for batch := range slices.Chunk(members, 1000) {
			if _, _, err = b.store.ChangeInventoryHostGroups(b.meta.ID, newInventory.ID, batch, []int{group.ID}, nil, false); err != nil {
				return fmt.Errorf("group %q of inventory %q: %w", g.Name, e.Name, err)
			}
		}
	}
	return nil
}

func (e BackupRepository) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupRepository](e.Name, backup.Repositories); err != nil {
		return err
	}
	if e.SSHKey != nil && getEntryByName[BackupAccessKey](e.SSHKey, backup.Keys) == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	}
	return nil
}

func (e BackupRepository) Restore(b *BackupDB) error {
	var SSHKeyID int
	if k := findEntityByName[db.AccessKey](e.SSHKey, b.keys); k == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	} else {
		SSHKeyID = (*k).ID
	}

	repo := e.Repository
	repo.ProjectID = b.meta.ID
	repo.SSHKeyID = SSHKeyID

	newRepo, err := b.store.CreateRepository(repo)
	if err != nil {
		return err
	}
	b.repositories = append(b.repositories, newRepo)
	return nil
}

func (e BackupHostConfig) GetName() string {
	return e.Name
}

func (e BackupHostConfig) Verify(backup *BackupFormat) error {
	// The credential is remapped during the restore, so only what the mapping
	// points at can be checked here — but it must be, because Restore runs after
	// the project and its keys already exist. It also trims the name, which the
	// duplicate check below compares against.
	if err := e.ValidateMapping(); err != nil {
		return err
	}

	// Not verifyDuplicate: a mapping is unique on (type, name), and the table
	// enforces it, so a duplicate must fail here rather than half way through
	// the restore with the project already created. The stored name is trimmed,
	// so two entries differing only in surrounding whitespace are one mapping.
	duplicates := 0
	for _, other := range backup.HostConfigs {
		if other.Type == e.Type && strings.TrimSpace(other.Name) == e.Name {
			duplicates++
		}
	}
	if duplicates > 1 {
		return fmt.Errorf("%s is duplicate", e.Name)
	}

	if e.SSHKey == nil {
		return fmt.Errorf("SSHKey can not be empty")
	}
	key := getEntryByName[BackupAccessKey](e.SSHKey, backup.Keys)
	if key == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	}

	// The same rule ValidateHostConfig applies, checked before anything is
	// written.
	if err := e.ValidateCredential(key.Type); err != nil {
		return err
	}

	return nil
}

func (e BackupHostConfig) Restore(b *BackupDB) error {
	k := findEntityByName[db.AccessKey](e.SSHKey, b.keys)
	if k == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	}

	hostConfig := e.HostConfig
	hostConfig.ProjectID = b.meta.ID
	hostConfig.SSHKeyID = k.ID

	// A backup is a user supplied file, so a mapping restored from one goes
	// through the same checks as one created through the API: its host reaches a
	// generated ssh config, and its key must belong to the project.
	if err := hostConfig.Validate(); err != nil {
		return err
	}

	if err := db.ValidateHostConfig(b.store, &hostConfig); err != nil {
		return err
	}

	newHostConfig, err := b.store.CreateHostConfig(hostConfig)
	if err != nil {
		return err
	}

	b.hostConfigs = append(b.hostConfigs, newHostConfig)
	return nil
}

func (e BackupTemplate) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupTemplate](e.Name, backup.Templates); err != nil {
		return err
	}

	if getEntryByName[BackupRepository](&e.Repository, backup.Repositories) == nil {
		return fmt.Errorf("repository does not exist in repositories[].name")
	}

	if e.Inventory != nil && getEntryByName[BackupInventory](e.Inventory, backup.Inventories) == nil {
		return fmt.Errorf("inventory does not exist in inventories[].name")
	}

	if e.VaultKey != nil && getEntryByName[BackupAccessKey](e.VaultKey, backup.Keys) == nil {
		return fmt.Errorf("vault_key does not exist in keys[].name")
	}

	for _, grant := range e.Roles {
		if grant.Slug == "" {
			return fmt.Errorf("role grant has no slug")
		}
	}

	if e.Vaults != nil {
		for _, vault := range e.Vaults {
			if vault.VaultKey != nil {
				if getEntryByName[BackupAccessKey](vault.VaultKey, backup.Keys) == nil {
					return fmt.Errorf("vaults[].vaultKey does not exist in keys[].name")
				}
			}
		}
	}

	if e.View != nil && getEntryByName[BackupView](e.View, backup.Views) == nil {
		return fmt.Errorf("view does not exist in views[].name")
	}

	if buildTemplate := getEntryByName[BackupTemplate](e.BuildTemplate, backup.Templates); string(e.Type) == "deploy" && buildTemplate == nil {
		return fmt.Errorf("deploy is build but build_template does not exist in templates[].name")
	}

	return nil
}

func (e BackupTemplate) Restore(b *BackupDB) error {
	var InventoryID *int
	if e.Inventory != nil {
		if k := findEntityByName[db.Inventory](e.Inventory, b.inventories); k == nil {
			return fmt.Errorf("inventory does not exist in inventories[].name")
		} else {
			id := k.GetID()
			InventoryID = &id
		}
	}

	var EnvironmentIDs []int
	for i := range e.Environments {
		envName := &e.Environments[i]
		k := findEntityByName[db.Environment](envName, b.environments)
		if k == nil {
			return fmt.Errorf("environment does not exist in environments[].name")
		}
		EnvironmentIDs = append(EnvironmentIDs, k.GetID())
	}

	var RepositoryID int
	if k := findEntityByName[db.Repository](&e.Repository, b.repositories); k == nil {
		return fmt.Errorf("repository does not exist in repositories[].name")
	} else {
		RepositoryID = k.GetID()
	}

	var BuildTemplateID *int
	if string(e.Type) != "deploy" {
		BuildTemplateID = nil
	} else if k := findEntityByName[db.Template](e.BuildTemplate, b.templates); k == nil {
		BuildTemplateID = nil
	} else {
		BuildTemplateID = &(k.ID)
	}

	var ViewID *int
	if k := findEntityByName[db.View](e.View, b.views); k == nil {
		ViewID = nil
	} else {
		ViewID = &k.ID
	}

	template := e.Template
	template.ProjectID = b.meta.ID
	template.RepositoryID = RepositoryID
	template.EnvironmentIDs = EnvironmentIDs
	template.InventoryID = InventoryID
	template.ViewID = ViewID
	template.BuildTemplateID = BuildTemplateID

	newTemplate, err := b.store.CreateTemplate(template)
	if err != nil {
		return err
	}
	b.templates = append(b.templates, newTemplate)

	if e.Vaults != nil {
		for _, vault := range e.Vaults {
			var VaultKeyID *int

			if vault.VaultKey != nil {
				if k := findEntityByName[db.AccessKey](vault.VaultKey, b.keys); k == nil {
					return fmt.Errorf("vaults[].vaultKey does not exist in keys[].name")
				} else {
					VaultKeyID = &k.ID
				}
			}

			tplVault := vault.TemplateVault
			tplVault.ProjectID = b.meta.ID
			tplVault.TemplateID = newTemplate.ID
			tplVault.VaultKeyID = VaultKeyID

			_, err := b.store.CreateTemplateVault(tplVault)
			if err != nil {
				return err
			}
		}
	}

	for _, grant := range e.Roles {
		permissions := grant.Permissions & db.TemplateGrantPermissions
		if b.version < 4 && permissions != 0 {
			permissions |= db.CanReadProject
		}
		role, err := b.fileRole(grant.Slug)
		if errors.Is(err, db.ErrNotFound) {
			unassigned := b.unassignedRole(grant.Slug, ReasonNotInFile)
			unassigned.Templates = append(unassigned.Templates, UnassignedTemplate{ID: newTemplate.ID, Name: newTemplate.Name, Permissions: permissions})
			continue
		}
		if err != nil {
			return err
		}
		_, err = b.store.CreateTemplateRole(db.TemplateRolePerm{
			TemplateID:  newTemplate.ID,
			RoleID:      role.ID,
			ProjectID:   b.meta.ID,
			Permissions: permissions,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// fileRole is a role from the file or a builtin one; any other role is the user's choice, never a slug match.
func (b *BackupDB) fileRole(slug string) (db.Role, error) {
	for _, role := range b.roles {
		if role.Slug == slug {
			return role, nil
		}
	}
	role, err := b.store.GetSharedRoleBySlug(slug)
	if err == nil && !role.Builtin {
		return db.Role{}, db.ErrNotFound
	}
	return role, err
}

func (e BackupEndpoint) Verify(backup *BackupFormat) error {
	if e.Name == "" {
		return fmt.Errorf("endpoint has no name")
	}
	return verifyDuplicate[BackupEndpoint](e.Name, backup.Endpoints)
}

func (e BackupEndpoint) Restore(b *BackupDB) error {
	endpoint := e.Endpoint
	endpoint.ProjectID = b.meta.ID
	endpoint.AuthSecretID = nil
	if k := findEntityByName[db.AccessKey](e.AuthSecretName, b.keys); k != nil {
		endpoint.AuthSecretID = &k.ID
	}

	newEndpoint, err := b.store.CreateEndpoint(endpoint)
	if err != nil {
		return err
	}
	b.endpoints = append(b.endpoints, newEndpoint)

	// An alias is unique across the instance: when the project the backup was
	// taken from still holds it, the restored endpoint goes without it and the
	// user adds a new one, as with the integration aliases before 2.21.11.
	for _, a := range e.Aliases {
		_, _ = b.store.CreateEndpointAlias(db.EndpointAlias{Alias: a, ProjectID: b.meta.ID, EndpointID: newEndpoint.ID})
	}

	return nil
}

func (e BackupIntegrationGroup) Verify(backup *BackupFormat) error {
	if e.Name == "" {
		return fmt.Errorf("integration group has no name")
	}
	if getEntryByName[BackupEndpoint](&e.Endpoint, backup.Endpoints) == nil {
		return fmt.Errorf("endpoint does not exist in endpoints[].name")
	}
	return verifyDuplicate[BackupIntegrationGroup](e.Name, backup.IntegrationGroups)
}

func (e BackupIntegrationGroup) Restore(b *BackupDB) error {
	endpoint := findEntityByName[db.Endpoint](&e.Endpoint, b.endpoints)
	if endpoint == nil {
		return fmt.Errorf("endpoint does not exist in endpoints[].name")
	}

	group := e.IntegrationGroup
	group.ProjectID = b.meta.ID
	group.EndpointID = endpoint.ID
	group.AuthSecretID = nil
	if k := findEntityByName[db.AccessKey](e.AuthSecretName, b.keys); k != nil {
		group.AuthSecretID = &k.ID
	}

	newGroup, err := b.store.CreateIntegrationGroup(group)
	if err != nil {
		return err
	}
	b.integrationGroups = append(b.integrationGroups, newGroup)

	return nil
}

func (e BackupIntegration) Verify(backup *BackupFormat) error {
	if e.Group == nil {
		return fmt.Errorf("integration has no group")
	}
	if getEntryByName[BackupIntegrationGroup](e.Group, backup.IntegrationGroups) == nil {
		return fmt.Errorf("integration group does not exist in integration_groups[].name")
	}

	switch {
	case e.Template != nil && e.Workflow != nil:
		return fmt.Errorf("integration runs either a template or a workflow, not both")
	case e.Template != nil:
		if getEntryByName[BackupTemplate](e.Template, backup.Templates) == nil {
			return fmt.Errorf("template does not exist in templates[].name")
		}
	case e.Workflow != nil:
		if getEntryByName[BackupWorkflow](e.Workflow, backup.Workflows) == nil {
			return fmt.Errorf("workflow does not exist in workflows[].name")
		}
	default:
		return fmt.Errorf("integration has neither a template nor a workflow")
	}
	return nil
}

func (e BackupIntegration) Restore(b *BackupDB) error {
	// Groups are restored before integrations, so b.integrationGroups is complete.
	group := findEntityByName[db.IntegrationGroup](e.Group, b.integrationGroups)
	if group == nil {
		return fmt.Errorf("integration group does not exist in integration_groups[].name")
	}

	integration := e.Integration
	integration.ProjectID = b.meta.ID
	integration.GroupID = group.ID
	integration.TemplateID = nil
	integration.WorkflowID = nil

	switch {
	case e.Template != nil:
		tpl := findEntityByName[db.Template](e.Template, b.templates)
		if tpl == nil {
			return fmt.Errorf("template does not exist in templates[].name")
		}
		integration.TemplateID = &tpl.ID
	case e.Workflow != nil:
		// Workflows are restored before integrations, so b.workflows is complete.
		workflow := findEntityByName[db.WorkflowTemplate](e.Workflow, b.workflows)
		if workflow == nil {
			return fmt.Errorf("workflow does not exist in workflows[].name")
		}
		integration.WorkflowID = &workflow.ID
	default:
		return fmt.Errorf("integration has neither a template nor a workflow")
	}

	if integration.TaskParams != nil {
		inv := findEntityByName[db.Inventory](e.TaskParams.InventoryName, b.inventories)
		if inv != nil {
			integration.TaskParams.InventoryID = &inv.ID
		}
	}

	newIntegration, err := b.store.CreateIntegration(integration)
	if err != nil {
		return err
	}
	b.integrations = append(b.integrations, newIntegration)

	for _, m := range e.Matchers {
		m.IntegrationID = newIntegration.ID
		_, _ = b.store.CreateIntegrationMatcher(b.meta.ID, m)
	}

	for _, v := range e.ExtractValues {
		v.IntegrationID = newIntegration.ID
		_, _ = b.store.CreateIntegrationExtractValue(b.meta.ID, v)
	}

	return nil
}

func (e BackupRunner) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupRunner](e.Name, backup.Runners)
}

func (e BackupRunner) Restore(b *BackupDB) error {
	runner := e.Runner
	runner.ProjectID = &b.meta.ID
	runner.Token = "" // Unregistered runner
	newRunner, err := b.store.CreateRunner(runner)
	if err != nil {
		return err
	}
	b.runners = append(b.runners, newRunner)
	return nil
}

func (e BackupWorkflow) Verify(backup *BackupFormat) error {
	for _, grant := range e.Roles {
		if grant.Slug == "" {
			return fmt.Errorf("role grant has no slug")
		}
	}
	if err := verifyDuplicate[BackupWorkflow](e.Name, backup.Workflows); err != nil {
		return err
	}

	for _, n := range e.Nodes {
		if n.Template != nil && getEntryByName[BackupTemplate](n.Template, backup.Templates) == nil {
			return fmt.Errorf("template does not exist in templates[].name")
		}
	}

	return nil
}

func (e BackupWorkflow) Restore(b *BackupDB) error {
	workflow := e.WorkflowTemplate
	workflow.ID = 0
	workflow.ProjectID = b.meta.ID

	nodes := make([]db.WorkflowNode, len(e.Nodes))
	for i, bn := range e.Nodes {
		// Keep the node ID (used by edges) but clear all project-scoped
		// references; they are re-resolved from names below.
		node := bn.WorkflowNode
		node.WorkflowTemplateID = 0
		node.TemplateID = 0
		node.TaskParamsID = nil

		if bn.Template != nil {
			tpl := findEntityByName[db.Template](bn.Template, b.templates)
			if tpl == nil {
				return fmt.Errorf("template does not exist in templates[].name")
			}
			node.TemplateID = tpl.ID
		}

		if node.TaskParams != nil {
			node.TaskParams.InventoryID = nil
			inv := findEntityByName[db.Inventory](node.TaskParams.InventoryName, b.inventories)
			if inv != nil {
				node.TaskParams.InventoryID = &inv.ID
			}
		}

		nodes[i] = node
	}

	workflow.Nodes = nodes
	// workflow.Edges is carried by the embedded WorkflowTemplate and references
	// nodes by the IDs preserved above; writeWorkflowGraph remaps them.

	newWorkflow, err := b.workflowStore.CreateWorkflowTemplate(workflow)
	if err != nil {
		return err
	}

	for _, grant := range e.Roles {
		permissions := grant.Permissions & db.WorkflowGrantPermissions
		if b.version < 4 && permissions != 0 {
			permissions |= db.CanReadProject
		}
		role, err := b.fileRole(grant.Slug)
		if errors.Is(err, db.ErrNotFound) {
			unassigned := b.unassignedRole(grant.Slug, ReasonNotInFile)
			unassigned.Workflows = append(unassigned.Workflows, UnassignedWorkflow{ID: newWorkflow.ID, Name: newWorkflow.Name, Permissions: permissions})
			continue
		}
		if err != nil {
			return err
		}
		_, err = b.store.CreateWorkflowRole(db.WorkflowRolePerm{
			WorkflowID:  newWorkflow.ID,
			RoleID:      role.ID,
			ProjectID:   b.meta.ID,
			Permissions: permissions,
		})
		if err != nil {
			return err
		}
	}

	b.workflows = append(b.workflows, newWorkflow)
	return nil
}

func (backup *BackupFormat) Verify() error {
	grants := 0
	for _, t := range backup.Templates {
		grants += len(t.Roles)
	}
	// Roles got slugs in version 2, so an older file without them still reads.
	if backup.Version < 2 && (len(backup.Roles) > 0 || grants > 0) {
		return fmt.Errorf("backup was written by an older version of Semaphore, export the project again")
	}
	if backup.Version > BackupFormatVersion {
		return fmt.Errorf("backup was written by a newer version of Semaphore")
	}
	for i, o := range backup.Environments {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at environments[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Views {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at views[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Schedules {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at schedules[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Keys {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at keys[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Repositories {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at repositories[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.HostConfigs {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at host_configs[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Inventories {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at inventories[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.SecretStorages {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at secret storage[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Templates {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at templates[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Roles {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at roles[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Runners {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at runners[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Workflows {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at workflows[%d]: %s", i, err.Error())
		}
	}
	backup.upgradeLegacyIntegrations()
	for i, o := range backup.Endpoints {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at endpoints[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.IntegrationGroups {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at integration_groups[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Integration {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at integrations[%d]: %s", i, err.Error())
		}
	}

	return nil
}

// Restore creates the project whole or not at all: a failed entry removes it again.
func (backup *BackupFormat) Restore(user db.User, store db.Store, workflowStore db.WorkflowManager) (restored *RestoredProject, err error) {
	var b = BackupDB{store: store, workflowStore: workflowStore, unassigned: []UnassignedRole{}, version: backup.Version}
	project := backup.Meta.Project

	// Prevent importing a project with a name that already exists
	existingProjects, err := b.store.GetAllProjects()
	if err == nil {
		for _, p := range existingProjects {
			if p.Name == project.Name { // exact name match
				return nil, common_errors.NewValidationError(fmt.Sprintf("project with name '%s' already exists", project.Name))
			}
		}
	}

	ownerRole, err := b.store.GetSharedRoleBySlug(string(db.ProjectOwner))
	if err != nil {
		return nil, err
	}

	newProject, err := b.store.CreateProject(project)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			return
		}
		if deleteErr := b.store.DeleteProject(newProject.ID); deleteErr != nil {
			err = &CleanupError{ProjectID: newProject.ID, Err: err, DeleteErr: deleteErr}
		}
	}()

	if err = b.store.CreateProjectUser(newProject.ID, user.ID, []int{ownerRole.ID}); err != nil {
		return nil, err
	}

	b.meta = newProject
	for i, o := range backup.SecretStorages {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at secret storage[%d]: %w", i, err)
		}
	}

	for i, o := range backup.Roles {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at roles[%d]: %w", i, err)
		}
	}

	for i, o := range backup.Environments {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at environments[%d]: %w", i, err)
		}
	}

	for i, o := range backup.Views {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at views[%d]: %w", i, err)
		}
	}

	for i, o := range backup.Keys {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at keys[%d]: %w", i, err)
		}
	}

	for i, o := range backup.Repositories {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at repositories[%d]: %w", i, err)
		}
	}

	for i, o := range backup.HostConfigs {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at host_configs[%d]: %w", i, err)
		}
	}

	for i, o := range backup.Inventories {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at inventories[%d]: %w", i, err)
		}
	}

	deployTemplates := make([]int, 0)
	for i, o := range backup.Templates {
		if string(o.Type) == "deploy" {
			deployTemplates = append(deployTemplates, i)
			continue
		}
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at templates[%d]: %w", i, err)
		}
	}

	for _, i := range deployTemplates {
		o := backup.Templates[i]
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at templates[%d]: %w", i, err)
		}
	}

	for i, o := range backup.Runners {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at runners[%d]: %w", i, err)
		}
	}

	// Workflows come after the templates, inventories and environments their
	// nodes reference by name, all of which must already exist in the project.
	for i, o := range backup.Workflows {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at workflows[%d]: %w", i, err)
		}
	}

	// Schedules come after workflows: one may start a workflow, and nothing
	// restored above refers to a schedule.
	for i, o := range backup.Schedules {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at schedules[%d]: %w", i, err)
		}
	}

	// Endpoints need only the keys, groups need their endpoint; both come
	// before the integrations that belong to them.
	backup.upgradeLegacyIntegrations()
	for i, o := range backup.Endpoints {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at endpoints[%d]: %w", i, err)
		}
	}
	for i, o := range backup.IntegrationGroups {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at integration_groups[%d]: %w", i, err)
		}
	}

	// Integrations come after workflows: one may start a workflow, and nothing
	// restored above refers to an integration.
	for i, o := range backup.Integration {
		if err := o.Restore(&b); err != nil {
			return nil, fmt.Errorf("error at integrations[%d]: %w", i, err)
		}
	}

	return &RestoredProject{Project: newProject, UnassignedRoles: b.unassigned}, nil
}
