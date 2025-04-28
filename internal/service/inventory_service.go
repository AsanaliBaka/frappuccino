package service

import (
	"context"
	"database/sql"
	"errors"

	repo "frappuccino/internal/repo"

	"frappuccino/internal/models"

	"github.com/lib/pq"
)

type InventoryService interface {
	GetAllInventory(ctx context.Context) ([]*models.InventoryItem, error)
	CreateInventory(ctx context.Context, item *models.InventoryItem) error
	GetInventoryByID(ctx context.Context, id string) (*models.InventoryItem, error)
	UpdateInventoryByID(ctx context.Context, id string, item *models.InventoryItem) error
	DeleteInventoryByID(ctx context.Context, id string) error
	GetInventoryList(ctx context.Context, sortBy string, page, pageSize int) ([]models.InventoryItem, int, bool, int, error)
}

type inventoryService struct {
	repo *repo.Container
}

func NewInventoryService(r *repo.Container) *inventoryService {
	return &inventoryService{repo: r}
}

func (s *inventoryService) GetAllInventory(ctx context.Context) ([]*models.InventoryItem, error) {
	items, err := s.repo.InventoryRepo.GetAllInventory(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return items, nil
		}
		return nil, err
	}
	return items, nil
}

func (s *inventoryService) CreateInventory(ctx context.Context, item *models.InventoryItem) error {
	if err := item.Validate(); err != nil {
		return models.NewError(models.ErrInvalidInput, errors.New("invalid input"))
	}

	if err := s.repo.InventoryRepo.CreateInventory(ctx, item); err != nil {
		return err
	}
	return nil
}

func (s *inventoryService) GetInventoryByID(ctx context.Context, id string) (*models.InventoryItem, error) {
	item, err := s.repo.InventoryRepo.GetInventoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewError(models.ErrNotFound, errors.New("inventory item not found"))
		}
		return nil, err
	}
	return item, nil
}

func (s *inventoryService) UpdateInventoryByID(ctx context.Context, id string, item *models.InventoryItem) error {
	if err := item.Validate(); err != nil {
		return models.NewError(models.ErrInvalidInput, err)
	}

	if _, err := s.GetInventoryByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.InventoryRepo.UpdateInventoryByID(ctx, id, item); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return models.NewError(models.ErrElemExist, err)
		}
		return models.NewError(models.ErrInternal, err)
	}
	return nil
}

func (s *inventoryService) DeleteInventoryByID(ctx context.Context, id string) error {
	if _, err := s.GetInventoryByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.InventoryRepo.DeleteInventoryByID(ctx, id); err != nil {
		return models.NewError(models.ErrInternal, err)
	}
	return nil
}

func (s *inventoryService) GetInventoryList(ctx context.Context, sortBy string, page, pageSize int) ([]models.InventoryItem, int, bool, int, error) {
	if sortBy != "price" && sortBy != "quantity" {
		sortBy = "quantity"
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	items, currentPage, hasNext, totalPages, err := s.repo.InventoryRepo.GetInventoryList(ctx, sortBy, page, pageSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.InventoryItem{}, page, false, 0, nil
		}
		return nil, 0, false, 0, err
	}

	return items, currentPage, hasNext, totalPages, nil
}
