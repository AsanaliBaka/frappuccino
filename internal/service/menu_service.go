package service

import (
	"context"
	"database/sql"
	"errors"

	"frappuccino/internal/models"
	repo "frappuccino/internal/repo"

	"github.com/lib/pq"
)

type MenuService interface {
	GetAllMenus(ctx context.Context) (listMenu []*models.Product, err error)
	CreateMenu(ctx context.Context, item *models.Product) (err error)
	GetMenuByID(ctx context.Context, id string) (item *models.Product, err error)
	UpdateMenu(ctx context.Context, id string, item *models.Product) (err error)
	DeleteMenu(ctx context.Context, id string) (err error)
}

type menuService struct {
	repo *repo.Container
}

func NewMenuService(r *repo.Container) MenuService {
	return &menuService{
		repo: r,
	}
}

func (m *menuService) GetAllMenus(ctx context.Context) (listMenu []*models.Product, err error) {
	listMenu, err = m.repo.MenuRepo.GetAllProducts(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return listMenu, nil
		}
		return nil, err
	}
	return
}

func (m *menuService) CreateMenu(ctx context.Context, item *models.Product) (err error) {
	if err = item.CheckRequiredFields(); err != nil {
		return models.NewError(models.ErrInvalidInput, err)
	}

	for _, ingredient := range item.Components {
		if _, err := m.repo.InventoryRepo.GetInventoryByID(ctx, ingredient.ComponentID); err != nil {
			return models.NewError(models.ErrNotFound, err)
		}
	}

	err = m.repo.MenuRepo.CreateProduct(ctx, item)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return models.NewError(models.ErrElemExist, err)
		}
		return models.NewError(models.ErrInternal, err)
	}
	return
}

func (m *menuService) GetMenuByID(ctx context.Context, id string) (item *models.Product, err error) {
	item, err = m.repo.MenuRepo.GetProductByID(ctx, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "22P02" {
			return nil, models.NewError(
				models.ErrInvalidInput,
				errors.New("invalid UUID format"),
			)
		}

		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewError(
				models.ErrNotFound,
				errors.New("menu item not found"),
			)
		}
		return nil, err
	}
	return
}

func (m *menuService) UpdateMenu(ctx context.Context, id string, item *models.Product) (err error) {
	if err = item.CheckRequiredFields(); err != nil {
		return models.NewError(models.ErrInvalidInput, err)
	}

	oldItem, err := m.GetMenuByID(ctx, id)
	if err != nil {
		return err
	}

	for _, ingredient := range item.Components {
		if _, err := m.repo.MenuRepo.GetProductByID(ctx, ingredient.ComponentID); err != nil {
			return models.NewError(models.ErrNotFound, err)
		}
	}

	if oldItem.UnitPrice != item.UnitPrice {
		priceHistory := &models.PriceHistory{
			MenuItemID: id,
			OldPrice:   oldItem.UnitPrice,
			NewPrice:   item.UnitPrice,
		}

		err = m.repo.MenuRepo.UpdatePrice(ctx, priceHistory)
		if err != nil {
			return models.NewError(models.ErrInternal, err)
		}
	}

	err = m.repo.MenuRepo.UpdateProduct(ctx, id, item)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return models.NewError(models.ErrElemExist, err)
		}
		return models.NewError(models.ErrInternal, err)
	}
	return
}

func (m *menuService) DeleteMenu(ctx context.Context, id string) (err error) {
	if _, err = m.GetMenuByID(ctx, id); err != nil {
		return err
	}

	err = m.repo.MenuRepo.DeleteProduct(ctx, id)
	if err != nil {
		return models.NewError(models.ErrInternal, err)
	}
	return
}
