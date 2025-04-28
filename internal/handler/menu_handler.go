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

type MenuHandler struct {
	MenuSvc service.MenuService
}

func NewMenuHandler(svc service.MenuService) *MenuHandler {
	return &MenuHandler{
		MenuSvc: svc,
	}
}

func (h *MenuHandler) GetAllMenus(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling CreateNewMenuItem request: %s", r.URL.Path)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	listMenu, err := h.MenuSvc.GetAllMenus(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timed out")
			return
		}

		slog.Error("Failed to get all menu items: %s", err.Error())
		Err := FromError(err)
		Respond(w, Err.Status, "Failed to get all menu items")
		return
	}

	Respond(w, http.StatusOK, listMenu)
}

func (h *MenuHandler) CreateMenu(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		Respond(w, http.StatusUnsupportedMediaType, "content type is not application/json")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	item, err := json.UnmarshalJson[*models.Product](data)
	if err != nil {
		Respond(w, http.StatusBadRequest, "Failed to unmarshal JSON data")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.MenuSvc.CreateMenu(ctx, item)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusGatewayTimeout, "Request timed out")
			return
		}

		slog.Error("Failed to create new menu item: %s", err.Error())
		Err := FromError(err)
		Respond(w, Err.Status, "Failed to create new menu item")
		return
	}

	Respond(w, http.StatusCreated, "Menu item created successfully")
}

func (h *MenuHandler) GetMenuByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.PathValue("id")
	item, err := h.MenuSvc.GetMenuByID(ctx, id)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusGatewayTimeout, "Request timed out")
			return
		}

		slog.Error("Failed to get menu item by ID: %s", err.Error())
		Err := FromError(err)
		Respond(w, Err.Status, "Failed to get menu item by ID")
		return
	}

	Respond(w, http.StatusOK, item)
}

func (h *MenuHandler) UpdateMenu(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		Respond(w, http.StatusUnsupportedMediaType, "content type is not application/json")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		Respond(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	item, err := json.UnmarshalJson[*models.Product](data)
	if err != nil {
		Respond(w, http.StatusBadRequest, "Failed to unmarshal JSON data")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.PathValue("id")
	err = h.MenuSvc.UpdateMenu(ctx, id, item)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusGatewayTimeout, "Request timed out")
			return
		}

		slog.Error("Failed to update menu item: %s", err.Error())
		Err := FromError(err)
		Respond(w, Err.Status, "Failed to update menu item")
		return
	}

	Respond(w, http.StatusOK, "Menu item updated successfully")
}

func (h *MenuHandler) DeleteMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.PathValue("id")
	err := h.MenuSvc.DeleteMenu(ctx, id)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusGatewayTimeout, "Request timed out")
			return
		}

		slog.Error("Failed to delete menu item: %s", err.Error())
		Err := FromError(err)
		Respond(w, Err.Status, "Failed to delete menu item")
		return
	}

	Respond(w, http.StatusOK, "Menu item deleted successfully")
}
