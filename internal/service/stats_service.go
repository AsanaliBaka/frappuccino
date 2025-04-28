package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"frappuccino/internal/models"
	repo "frappuccino/internal/repo"
)

type StatsService interface {
	GetPopularItem(ctx context.Context) ([]models.TopProduct, error)
	GetTotalSum(ctx context.Context) (*models.RevenueSummary, error)
	GetSearch(ctx context.Context, query, filter string, minPrice *float64, maxPrice *float64) (models.LookupResult, error)
	GetItemByPeriod(ctx context.Context, period string, month string, year int) (map[string]interface{}, error)
}

type statsService struct {
	Repo *repo.Container
}

func NewStatsService(r *repo.Container) StatsService {
	return &statsService{
		Repo: r,
	}
}

func (m *statsService) GetPopularItem(ctx context.Context) ([]models.TopProduct, error) {
	list, err := m.Repo.StatsRepo.GetPopularItem(ctx)
	if err != nil {
		return nil, models.NewError(nil, err)
	}
	return list, nil
}

func (m *statsService) GetTotalSum(ctx context.Context) (*models.RevenueSummary, error) {
	list, err := m.Repo.StatsRepo.GetTotalSum(ctx)
	if err != nil {
		return nil, models.NewError(nil, err)
	}
	return list, nil
}

func (m *statsService) GetSearch(ctx context.Context, query, filter string, minPrice *float64, maxPrice *float64) (models.LookupResult, error) {
	var response models.LookupResult

	filters := strings.Split(filter, ",")
	searchMenu := true
	searchOrders := true

	if filter != "" {
		searchMenu = false
		searchOrders = false
		for _, f := range filters {
			if f == "menu" {
				searchMenu = true
			}
			if f == "orders" {
				searchOrders = true
			}
		}
	}

	if searchMenu {
		menuItems, err := m.Repo.StatsRepo.SearchMenu(ctx, query, minPrice, maxPrice)
		if err != nil {
			return response, err
		}
		response.Products = menuItems
		response.MatchesFound += len(menuItems)
	}

	if searchOrders {
		orders, err := m.Repo.StatsRepo.SearchOrders(ctx, query, minPrice, maxPrice)
		if err != nil {
			return response, err
		}
		response.RecentOrders = orders
		response.MatchesFound += len(orders)
	}

	return response, nil
}

func (m *statsService) GetItemByPeriod(ctx context.Context, period string, month string, year int) (map[string]interface{}, error) {
	if period != "day" && period != "month" {
		return nil, errors.New("invalid period parameter")
	}

	if period == "day" && month == "" {
		return nil, errors.New("month is required when period is 'day'")
	}

	if year == 0 {
		year = time.Now().Year()
	}

	data, err := m.Repo.StatsRepo.GetItemByPeriod(ctx, period, month, year)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{
		"period": period,
	}

	if period == "day" {
		response["month"] = month
	}

	orderedItems := []map[string]int{}
	for _, item := range data {
		orderedItems = append(orderedItems, map[string]int{
			item.Date.Format("2"): item.Count,
		})
	}
	response["orderedItems"] = orderedItems

	return response, nil
}
