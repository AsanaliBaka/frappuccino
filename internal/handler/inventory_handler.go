package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"frappuccino/internal/models"
	"frappuccino/internal/service"
	"frappuccino/internal/slog"
	"frappuccino/pkg/cerrors"
)

type InventoryHandler struct {
	InvService service.InventoryService
}

func NewInventoryHandler(svc service.InventoryService) *InventoryHandler {
	return &InventoryHandler{InvService: svc}
}

func (h *InventoryHandler) GetInventoryItems(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling GetInventoryItems request: path=%v", r.URL.Path)
	items, err := h.InvService.GetAllInventory(r.Context())
	if err != nil {
		slog.Error("Error getting inventory items: %v", err)
		Respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	slog.Info("Retrieved %d inventory items", len(items))
	if len(items) == 0 {
		Respond(w, http.StatusOK, map[string]string{"message": "No inventory items found"})
		return
	}
	Respond(w, http.StatusOK, items)
}

func (h *InventoryHandler) AddInventoryItem(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		Respond(w, http.StatusUnsupportedMediaType, map[string]string{"error": "content type is not application/json"})
		return
	}
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		Respond(w, http.StatusInternalServerError, map[string]string{"error": "invalid request body"})
		return
	}
	var newItem models.InventoryItem
	if err := json.Unmarshal(data, &newItem); err != nil {
		Respond(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON format"})
		return
	}
	if err := h.InvService.CreateInventory(r.Context(), &newItem); err != nil {
		if errors.Is(err, cerrors.ErrAlreadyExists) {
			slog.Error("Inventory item already exists: %v", err)
			Respond(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		slog.Error("Error creating inventory item: %v", err)
		Respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	slog.Info("Inventory item created successfully: %v", newItem)
	Respond(w, http.StatusCreated, map[string]string{"message": "Inventory created"})
}

func (h *InventoryHandler) GetInventoryItemId(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling GetInventoryItemId request: path=%v", r.URL.Path)
	id := r.PathValue("id")
	item, err := h.InvService.GetInventoryByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, cerrors.ErrNotExist) {
			slog.Error("Inventory item not found: id=%v", id)
			Respond(w, http.StatusNotFound, map[string]string{"error": cerrors.ErrNotExist.Error()})
			return
		}
		slog.Error("Error retrieving inventory item: id=%v, error=%v", id, err)
		Respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	slog.Info("Retrieved inventory item: id=%v", id)
	Respond(w, http.StatusOK, item)
}

func (h *InventoryHandler) UpdateInventoryItem(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		Respond(w, http.StatusUnsupportedMediaType, map[string]string{"error": "content type is not application/json"})
		return
	}
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		Respond(w, http.StatusInternalServerError, map[string]string{"error": "invalid request body"})
		return
	}
	var newItem models.InventoryItem
	if err := json.Unmarshal(data, &newItem); err != nil {
		Respond(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON format"})
		return
	}
	id := r.PathValue("id")
	if err := h.InvService.UpdateInventoryByID(r.Context(), id, &newItem); err != nil {
		if errors.Is(err, cerrors.ErrNotExist) {
			slog.Error("Inventory item not found for update, id=%v", id)
			Respond(w, http.StatusNotFound, map[string]string{"error": cerrors.ErrNotExist.Error()})
			return
		}
		slog.Error("Error updating inventory item: id=%v, error=%v", id, err)
		Respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	slog.Info("Inventory item updated successfully, id=%v", id)
	Respond(w, http.StatusOK, map[string]string{"message": "Inventory updated"})
}

func (h *InventoryHandler) DeleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.InvService.DeleteInventoryByID(r.Context(), id); err != nil {
		if errors.Is(err, cerrors.ErrNotExist) {
			slog.Error("Inventory item not found for deletion, id=%v", id)
			Respond(w, http.StatusNotFound, map[string]string{"error": cerrors.ErrNotExist.Error()})
			return
		}
		slog.Error("Error deleting inventory item: id=%v, error=%v", id, err)
		Respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	slog.Info("Inventory item deleted successfully: id=%v", id)
	Respond(w, http.StatusOK, map[string]string{"message": "Inventory item deleted successfully"})
}

func (h *InventoryHandler) GetInventoryList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sortBy := query.Get("sortBy")
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	items, currentPage, hasNextPage, totalPages, err := h.InvService.GetInventoryList(r.Context(), sortBy, page, pageSize)
	if err != nil {
		Respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	response := map[string]interface{}{
		"leftovers":   items,
		"currentPage": currentPage,
		"hasNextPage": hasNextPage,
		"totalPages":  totalPages,
	}
	Respond(w, http.StatusOK, response)
}
