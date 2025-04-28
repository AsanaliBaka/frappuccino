package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"frappuccino/internal/models"
	repo "frappuccino/internal/repo"

	"github.com/lib/pq"
)

type OrderService interface {
	GetAllOrders(ctx context.Context) (listOrders []*models.Purchase, err error)
	CreateOrder(ctx context.Context, order *models.Purchase) error
	GetOrderById(ctx context.Context, id string) (order *models.Purchase, err error)
	UpdateOrder(ctx context.Context, id string, order *models.Purchase) error
	DeleteOrder(ctx context.Context, id string) error
	CloseOrder(ctx context.Context, id string) error
	BatchProcessOrders(ctx context.Context, listOrders []*models.Purchase) (*models.PurchaseResult, error)
	GetNumberOfOrderedItems(ctx context.Context, startDate, endDate *time.Time) (map[string]int, error)
}

type orderService struct {
	Repo *repo.Container
}

func NewOrderService(r *repo.Container) OrderService {
	return &orderService{
		Repo: r,
	}
}

func (s *orderService) GetAllOrders(ctx context.Context) (listOrders []*models.Purchase, err error) {
	listOrders, err = s.Repo.OrderRepo.GetAllOrders(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return listOrders, nil
		}
		return nil, err
	}
	return
}

func (s *orderService) CreateOrder(ctx context.Context, order *models.Purchase) error {
	if err := s.validateOrderInput(ctx, order); err != nil {
		return err
	}

	if err := s.checkInventoryAvailability(ctx, order); err != nil {
		return err
	}

	if err := s.calculateOrderPrices(ctx, order); err != nil {
		return err
	}

	order.Status = "open"

	return s.Repo.OrderRepo.CreateOrder(ctx, order)
}

func (s *orderService) GetOrderById(ctx context.Context, id string) (order *models.Purchase, err error) {
	order, err = s.Repo.OrderRepo.GetOrderByID(ctx, id)
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
				errors.New("order not found"),
			)
		}
		return nil, err
	}
	return
}

func (s *orderService) UpdateOrder(ctx context.Context, id string, order *models.Purchase) error {
	var err error

	err = s.validateOrderInput(ctx, order)
	if err != nil {
		return err
	}

	err = s.checkInventoryAvailability(ctx, order)
	if err != nil {
		return err
	}

	oldOrder, err := s.Repo.OrderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	if oldOrder.Status != "open" {
		return models.NewError(models.ErrInvalidInput, errors.New("order is not open"))
	}

	err = s.calculateOrderPrices(ctx, order)
	if err != nil {
		return err
	}

	return s.Repo.OrderRepo.UpdateOrder(ctx, id, order)
}

func (s *orderService) DeleteOrder(ctx context.Context, id string) error {
	err := s.Repo.OrderRepo.DeleteOrder(ctx, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "22P02" {
			return models.NewError(models.ErrInvalidInput, errors.New("invalid UUID format"))
		}
	}

	return err
}

func (s *orderService) CloseOrder(ctx context.Context, id string) error {
	order, err := s.GetOrderById(ctx, id)
	if err != nil {
		return err
	}

	if order.Status != "open" {
		return models.NewError(models.ErrInvalidInput, errors.New("order is not open"))
	}

	err = s.Repo.OrderRepo.CloseOrder(ctx, order)
	if err != nil {
		return err
	}

	return nil
}

func (s *orderService) BatchProcessOrders(ctx context.Context, listOrders []*models.Purchase) (*models.PurchaseResult, error) {
	var processedOrders []*models.Purchase
	var totalRevenue float64
	var rejectedCount, acceptedCount int

	inventoryChanges := make(map[string]*models.ProductComponent)

	for _, order := range listOrders {
		setRejected := func(err error) {
			rejectedCount++
			var str string
			if svcError, ok := err.(models.Error); ok {
				str = svcError.AppError().Error()
			} else {
				str = err.Error()
			}
			order.Note = &str
			order.Status = "rejected"
			processedOrders = append(processedOrders, order)
		}

		// Validate input
		if err := s.validateOrderInput(ctx, order); err != nil {
			setRejected(err)
			continue
		}

		if err := s.checkInventoryAvailability(ctx, order); err != nil {
			setRejected(err)
			continue
		}

		if err := s.calculateOrderPrices(ctx, order); err != nil {
			setRejected(err)
			continue
		}

		if err := s.Repo.OrderRepo.CreateOrder(ctx, order); err != nil {
			setRejected(err)
			continue
		}

		for _, item := range order.Positions {
			menu, err := s.Repo.MenuRepo.GetProductByID(ctx, item.ItemID)
			if err != nil {
				setRejected(err)
				continue
			}

			for _, ingredient := range menu.Components {
				ingredientID := ingredient.ComponentID
				requiredQty := roundFloat(ingredient.RequiredQty*float64(item.Count), 2)

				if change, exists := inventoryChanges[ingredientID]; !exists {
					inventoryChanges[ingredientID] = &models.ProductComponent{
						ComponentID: ingredientID,
						RequiredQty: requiredQty,
					}
				} else {
					change.RequiredQty = roundFloat(change.RequiredQty+requiredQty, 2)
				}
			}
		}

		if err := s.Repo.OrderRepo.CloseOrder(ctx, order); err != nil {
			setRejected(err)
			continue
		}

		order.Status = "accepted"
		order.Note = nil
		totalRevenue += *order.Amount
		acceptedCount++
		processedOrders = append(processedOrders, order)
	}

	for id, change := range inventoryChanges {
		invtItem, err := s.Repo.InventoryRepo.GetInventoryByID(ctx, id)
		if err != nil {
			continue
		}
		inventoryChanges[id].ComponentName = invtItem.Title
		inventoryChanges[id].RequiredQty = change.RequiredQty
		remaining := roundFloat(invtItem.UnitCost-change.RequiredQty, 2)
		change.InStock = &remaining
	}

	invtList := make([]*models.ProductComponent, 0, len(inventoryChanges))
	for _, v := range inventoryChanges {
		invtList = append(invtList, v)
	}

	return &models.PurchaseResult{
		Handled: processedOrders,
		Report: models.ResultReport{
			Total:       len(listOrders),
			Confirmed:   acceptedCount,
			Declined:    rejectedCount,
			Revenue:     totalRevenue,
			StockEvents: invtList,
		},
	}, nil
}

func (s *orderService) GetNumberOfOrderedItems(ctx context.Context, startDate, endDate *time.Time) (map[string]int, error) {
	results, err := s.Repo.OrderRepo.GetNumberOfOrderedItems(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return results, nil
}
