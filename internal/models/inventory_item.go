package models

import (
	"errors"
	"strings"
)

type InventoryItem struct {
	IngredientID string  `json:"inventory_id"`
	Title        string  `json:"title"`
	Stock        float64 `json:"stock"`
	Measure      string  `json:"measure"`
	UnitCost     float64 `json:"unit_cost"`
}

type InventoryTransaction struct {
	ID         string  `json:"transaction_id"`
	ItemRef    string  `json:"inventory_ref"`
	Delta      float64 `json:"delta"`
	Type       string  `json:"type"`
	OccurredAt string  `json:"occurred_at"`
}

func (inv *InventoryItem) Validate() error {
	if strings.TrimSpace(inv.Title) == "" {
		return errors.New("inventory item title is required")
	}
	if inv.Stock <= 0 {
		return errors.New("stock must be greater than zero")
	}
	if inv.UnitCost <= 0 {
		return errors.New("unit cost must be greater than zero")
	}

	validMeasures := map[string]bool{"kg": true, "l": true, "pcs": true}
	if !validMeasures[inv.Measure] {
		return errors.New("invalid measurement unit (expected: kg, l, pcs)")
	}

	return nil
}
