package server

import "github.com/semaphoreui/semaphore/db"

type InventoryService interface {
	GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error)
}

func NewInventoryService(
	accessKeyRepo db.AccessKeyManager,
	repositoryRepo db.RepositoryManager,
	inventoryRepo db.InventoryManager,
	encryptionService AccessKeyEncryptionService,
) InventoryService {
	return &InventoryServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		repositoryRepo:    repositoryRepo,
		encryptionService: encryptionService,
		inventoryRepo:     inventoryRepo,
	}
}

type InventoryServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	repositoryRepo    db.RepositoryManager
	encryptionService AccessKeyEncryptionService
	inventoryRepo     db.InventoryManager
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
