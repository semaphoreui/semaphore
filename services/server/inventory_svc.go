package server

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

type InventoryService interface {
	GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error)
}

func NewInventoryService(
	accessKeyRepo db.AccessKeyManager,
	repositoryRepo db.RepositoryManager,
	inventoryRepo InventoryRepository,
	encryptionService AccessKeyEncryptionService,
) InventoryService {
	return &InventoryServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		repositoryRepo:    repositoryRepo,
		encryptionService: encryptionService,
		inventoryRepo:     inventoryRepo,
	}
}

// InventoryRepository reads inventories and the MCP servers of mcp ones.
type InventoryRepository interface {
	db.InventoryManager
	db.McpServerManager
}

type InventoryServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	repositoryRepo    db.RepositoryManager
	encryptionService AccessKeyEncryptionService
	inventoryRepo     InventoryRepository
}

func (s *InventoryServiceImpl) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	inventory, err = s.inventoryRepo.GetInventory(projectID, inventoryID)
	if err != nil {
		return
	}

	err = s.fillInventory(&inventory)
	return
}

func (s *InventoryServiceImpl) fillInventory(inventory *db.Inventory) (err error) {
	if inventory.Type == db.InventoryMcp {
		return s.fillMcpServers(inventory)
	}

	if inventory.SSHKeyID != nil {
		inventory.SSHKey, err = s.accessKeyRepo.GetAccessKey(inventory.ProjectID, *inventory.SSHKeyID)
	}

	if err != nil {
		return
	}

	if inventory.BecomeKeyID != nil {
		inventory.BecomeKey, err = s.accessKeyRepo.GetAccessKey(inventory.ProjectID, *inventory.BecomeKeyID)
	}

	if err != nil {
		return
	}

	if inventory.RepositoryID != nil {
		var repo db.Repository
		repo, err = s.repositoryRepo.GetRepository(inventory.ProjectID, *inventory.RepositoryID)
		if err != nil {
			return
		}

		err = s.encryptionService.DeserializeSecret(&repo.SSHKey)
		if err != nil {
			return
		}

		inventory.Repository = &repo
	}

	return
}

// fillMcpServers gives an mcp inventory its enabled servers, each with its
// decrypted token. It runs on the server for a task about to start.
// McpServer.Token is not serialized: a remote runner receives the token only
// as a decrypted access key of the runner state (RunnerController).
func (s *InventoryServiceImpl) fillMcpServers(inventory *db.Inventory) error {
	servers, err := s.inventoryRepo.GetMcpServers(inventory.ProjectID, inventory.ID)
	if err != nil {
		return err
	}

	keys, err := s.inventoryRepo.GetMcpServerTokens(inventory.ProjectID, inventory.ID)
	if err != nil {
		return err
	}

	tokens := make(map[int]db.AccessKey, len(keys))
	for _, key := range keys {
		if key.McpID == nil {
			continue
		}
		if err = s.encryptionService.DeserializeSecret(&key); err != nil {
			return fmt.Errorf("token of mcp server %s: %w", key.Name, err)
		}
		tokens[*key.McpID] = key
	}

	enabled := make([]db.McpServer, 0, len(servers))
	for _, server := range servers {
		if !server.Enabled {
			continue
		}
		if key, ok := tokens[server.ID]; ok {
			server.Token = key.String
			server.TokenKeyID = &key.ID
		}
		enabled = append(enabled, server)
	}

	inventory.McpServers = enabled
	return nil
}
