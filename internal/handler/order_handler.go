package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"frappuccino/internal/models"
	"frappuccino/internal/service"
	"frappuccino/internal/slog"
	"frappuccino/pkg/json"
)

type OrderHandler struct {
	OrderSvc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{
		OrderSvc: svc,
	}
}

func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	listOrders, err := h.OrderSvc.GetAllOrders(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error getting all orders: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, listOrders)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		Respond(w, http.StatusBadRequest, "content type is not application/json")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	item, err := json.UnmarshalJson[*models.Purchase](data)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to unmarshal request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.OrderSvc.CreateOrder(ctx, item)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error creating order: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusCreated, "Order created successfully")
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.PathValue("id")
	if id == "batch-process" {
		Respond(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	order, err := h.OrderSvc.GetOrderById(ctx, id)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error getting order by ID: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, order)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		Respond(w, http.StatusBadRequest, "content type is not application/json")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	item, err := json.UnmarshalJson[*models.Purchase](data)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to unmarshal request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.PathValue("id")

	if id == "batch-process" {
		Respond(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	err = h.OrderSvc.UpdateOrder(ctx, id, item)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error updating order: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.PathValue("id")
	if id == "batch-process" {
		Respond(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	err := h.OrderSvc.DeleteOrder(ctx, id)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error deleting order: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, "Order deleted successfully")
}

func (h *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.PathValue("id")
	err := h.OrderSvc.CloseOrder(ctx, id)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error closing order: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, "Order closed successfully")
}

func (h *OrderHandler) GetNumberOfOrderedItems(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var startDate, endDate *time.Time

	defaultStartDate := time.Now().AddDate(0, 0, -7)
	defaultEndDate := time.Now()

	startDateStr := r.URL.Query().Get("startDate")
	if startDateStr != "" {
		parsedStartDate, err := time.Parse("02.01.2006", startDateStr)
		if err != nil {
			Respond(w, http.StatusBadRequest, "invalid startDate format, must be DD.MM.YYYY")
			return
		}
		startDate = &parsedStartDate
	} else {
		startDate = &defaultStartDate
	}

	endDateStr := r.URL.Query().Get("endDate")
	if endDateStr != "" {
		parsedEndDate, err := time.Parse("02.01.2006", endDateStr)
		if err != nil {
			Respond(w, http.StatusBadRequest, "invalid endDate format, must be DD.MM.YYYY")
			return
		}
		endDate = &parsedEndDate
	} else {
		endDate = &defaultEndDate
	}

	numberOfOrderedItems, err := h.OrderSvc.GetNumberOfOrderedItems(ctx, startDate, endDate)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error getting number of ordered items: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, numberOfOrderedItems)
}

func (h *OrderHandler) BatchProcessOrders(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		Respond(w, http.StatusBadRequest, "content type is not application/json")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	batch, err := json.UnmarshalJson[*models.PurchaseBatch](data)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to unmarshal request body")
		return
	}

	res, err := h.OrderSvc.BatchProcessOrders(ctx, batch.Purchases)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}

		slog.Error("Error batch processing orders: %v", err)
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, res)
}
