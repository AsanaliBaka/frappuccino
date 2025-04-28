package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type Product struct {
	ProductID  string              `json:"product_id"`
	Title      string              `json:"title"`
	Details    string              `json:"details"`
	UnitPrice  float64             `json:"unit_price"`
	SizeLabel  string              `json:"size_label"`
	Group      string              `json:"group"`
	Labels     pq.StringArray      `json:"labels"`
	Extras     ExtrasMap           `json:"extras"`
	Components []*ProductComponent `json:"components"`
}

type ProductComponent struct {
	ComponentID   string   `json:"component_id"`
	ComponentName string   `json:"component_name"`
	RequiredQty   float64  `json:"required_qty"`
	InStock       *float64 `json:"in_stock,omitempty"`
}

func (p *Product) CheckRequiredFields() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("product title is required")
	}
	if p.UnitPrice <= 0 {
		return errors.New("unit price must be greater than 0")
	}
	if strings.TrimSpace(p.Group) == "" {
		return errors.New("product group is required")
	}
	if len(p.Components) == 0 {
		return errors.New("components list cannot be empty")
	}
	for _, comp := range p.Components {
		if strings.TrimSpace(comp.ComponentID) == "" {
			return errors.New("component id is required")
		}
		if comp.RequiredQty <= 0 {
			return errors.New("component quantity must be greater than 0")
		}
	}
	return nil
}

type ExtrasMap map[string]interface{}

func (e ExtrasMap) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *ExtrasMap) Scan(src interface{}) error {
	data, ok := src.([]byte)
	if !ok {
		return errors.New("failed to scan ExtrasMap: expected []byte")
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result, ok := raw.(map[string]interface{})
	if !ok {
		return errors.New("failed to scan ExtrasMap: invalid format")
	}

	*e = result
	return nil
}
