package handler

import (
	"net/http"

	"frappuccino/internal/service"
)

type Handler struct {
	InvHandler   *InventoryHandler
	MenuHandler  *MenuHandler
	OrderHandler *OrderHandler
	StatsHandler *StatsHandler
}

func NewHandler(invSvc service.InventoryService, menuSvc service.MenuService, orderSvc service.OrderService, statsSvc service.StatsService) *Handler {
	return &Handler{
		InvHandler:   NewInventoryHandler(invSvc),
		MenuHandler:  NewMenuHandler(menuSvc),
		OrderHandler: NewOrderHandler(orderSvc),
		StatsHandler: NewStatsHandler(statsSvc),
	}
}

func (h *Handler) newRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /inventory", h.InvHandler.GetInventoryItems)
	router.HandleFunc("POST /inventory", h.InvHandler.AddInventoryItem)
	router.HandleFunc("GET /inventory/{id}", h.InvHandler.GetInventoryItemId)
	router.HandleFunc("PUT /inventory/{id}", h.InvHandler.UpdateInventoryItem)
	router.HandleFunc("DELETE /inventory/{id}", h.InvHandler.DeleteInventoryItem)
	router.HandleFunc("GET /inventory/list", h.InvHandler.GetInventoryList)

	router.HandleFunc("GET /menu", h.MenuHandler.GetAllMenus)
	router.HandleFunc("POST /menu", h.MenuHandler.CreateMenu)
	router.HandleFunc("GET /menu/{id}", h.MenuHandler.GetMenuByID)
	router.HandleFunc("PUT /menu/{id}", h.MenuHandler.UpdateMenu)
	router.HandleFunc("DELETE /menu/{id}", h.MenuHandler.DeleteMenu)

	router.HandleFunc("GET /orders", h.OrderHandler.GetAllOrders)
	router.HandleFunc("POST /orders", h.OrderHandler.CreateOrder)
	router.HandleFunc("POST /orders/batch", h.OrderHandler.BatchProcessOrders)
	router.HandleFunc("GET /orders/{id}", h.OrderHandler.GetOrderByID)
	router.HandleFunc("PUT /orders/{id}", h.OrderHandler.UpdateOrder)
	router.HandleFunc("DELETE /orders/{id}", h.OrderHandler.DeleteOrder)
	router.HandleFunc("POST /orders/{id}/close", h.OrderHandler.CloseOrder)
	router.HandleFunc("GET /orders/number", h.OrderHandler.GetNumberOfOrderedItems)

	router.HandleFunc("GET /stats/total-sales", h.StatsHandler.GetTotalSum)
	router.HandleFunc("GET /stats/popular-items", h.StatsHandler.GetPopularItem)
	router.HandleFunc("GET /stats/search", h.StatsHandler.GetSearch)
	router.HandleFunc("GET /stats/orderedItemsByPeriod", h.StatsHandler.GetItemByPeriod)

	return router
}
