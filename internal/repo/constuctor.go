package internal

import (
	"database/sql"

	customer_repo "frappuccino/internal/repo/customer"
	inventory_repo "frappuccino/internal/repo/inventory"
	menu_repo "frappuccino/internal/repo/menu"
	order_repo "frappuccino/internal/repo/order"
	stats_repo "frappuccino/internal/repo/stats"
)

type Container struct {
	OrderRepo     order_repo.OrderRepo
	CustomerRepo  *customer_repo.CustomerRepo
	MenuRepo      menu_repo.MenuRepo
	InventoryRepo inventory_repo.InventoryRepo
	StatsRepo     stats_repo.StatsRepo
}

func New(db *sql.DB) *Container {
	return &Container{
		OrderRepo:     order_repo.NewOrderRepo(db),
		CustomerRepo:  customer_repo.NewCustomerRepo(db),
		MenuRepo:      menu_repo.NewMenuRepo(db),
		InventoryRepo: inventory_repo.NewInventoryRepo(db),
		StatsRepo:     stats_repo.NewStatsRepo(db),
	}
}
