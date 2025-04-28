package service

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"frappuccino/internal/models"

	"github.com/lib/pq"
)

func (s *orderService) validateOrderInput(ctx context.Context, order *models.Purchase) error {
	if err := order.Validate(); err != nil {
		return models.NewError(models.ErrInvalidInput, err)
	}

	// Check if customer exists
	if _, err := s.Repo.CustomerRepo.GetCustomerByID(ctx, order.CustomerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.NewError(models.ErrInvalidInput, errors.New("customer not found"))
		}
		return models.NewError(models.ErrInternal, errors.New("failed to get customer"))
	}

	return nil
}

func (s *orderService) calculateOrderPrices(ctx context.Context, order *models.Purchase) error {
	ids := order.GetItemIDs()
	menus, err := s.Repo.MenuRepo.FetchProductsByIDs(ctx, ids)
	if err != nil {
		return err
	}

	if len(menus) != len(ids) {
		return models.NewError(models.ErrInvalidInput, errors.New("product not found"))
	}

	menuMap := make(map[string]float64, len(ids))
	for _, menu := range menus {
		menuMap[menu.ProductID] = menu.UnitPrice
	}

	if order.Amount == nil {
		total := 0.0
		order.Amount = &total
	}

	for _, item := range order.Positions {
		if price, ok := menuMap[item.ItemID]; ok {
			item.UnitPrice = price
			*order.Amount += float64(item.Count) * price
		}
	}

	*order.Amount = math.Round(*order.Amount)
	return nil
}

func (s *orderService) checkInventoryAvailability(ctx context.Context, order *models.Purchase) error {
	hasIngredients, err := s.Repo.OrderRepo.CheckInventoryForOrder(ctx, order)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "22P02" {
			return models.NewError(models.ErrInvalidInput, errors.New("invalid UUID format"))
		}

		if errors.Is(err, sql.ErrNoRows) {
			return models.NewError(models.ErrInvalidInput, errors.New("product not found"))
		}
		return err
	}

	if !hasIngredients {
		return models.NewError(models.ErrInvalidInput, models.ErrInventoryNotAvailable)
	}

	return nil
}

func roundFloat(f float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(f*factor) / factor
}
