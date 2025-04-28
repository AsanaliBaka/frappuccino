package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Purchase struct {
	PurchaseID string      `json:"purchase_id"`
	CustomerID string      `json:"customer_id"`
	Positions  []*LineItem `json:"positions"`
	Status     string      `json:"status"`
	Amount     *float64    `json:"amount,omitempty"`
	Note       *string     `json:"note,omitempty"`
	Created    *time.Time  `json:"created,omitempty"`
	Updated    *time.Time  `json:"updated,omitempty"`
}

type LineItem struct {
	ItemID      string    `json:"item_id"`
	Count       int       `json:"count"`
	UnitPrice   float64   `json:"unit_price"`
	Adjustments ConfigMap `json:"adjustments"`
}

type PurchaseHistory struct {
	HistoryID  string    `json:"history_id"`
	PurchaseID string    `json:"purchase_id"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
}

type PurchaseBatch struct {
	Purchases []*Purchase `json:"purchases"`
}

type PurchaseResult struct {
	Handled []*Purchase  `json:"handled_purchases"`
	Report  ResultReport `json:"report"`
}

type ResultReport struct {
	Total       int                 `json:"total"`
	Confirmed   int                 `json:"confirmed"`
	Declined    int                 `json:"declined"`
	Revenue     float64             `json:"revenue"`
	StockEvents []*ProductComponent `json:"stock_events"`
}

type ConfigMap map[string]interface{}

func (c ConfigMap) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfigMap) Scan(src interface{}) error {
	raw, ok := src.([]byte)
	if !ok {
		return errors.New("failed to scan ConfigMap: expected []byte")
	}

	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}

	result, ok := parsed.(map[string]interface{})
	if !ok {
		return errors.New("failed to scan ConfigMap: invalid format")
	}

	*c = result
	return nil
}

func (p *Purchase) Validate() error {
	if strings.TrimSpace(p.CustomerID) == "" {
		return errors.New("client id is required")
	}
	if len(p.Positions) == 0 {
		return errors.New("at least one position is required")
	}
	if p.Status != "open" {
		return errors.New("purchase state must be 'open'")
	}
	for _, pos := range p.Positions {
		if strings.TrimSpace(pos.ItemID) == "" {
			return errors.New("missing item id")
		}
		if pos.Count <= 0 {
			return errors.New("invalid item count")
		}
	}
	return nil
}

func (p *Purchase) GetItemIDs() []string {
	var ids []string
	for _, item := range p.Positions {
		ids = append(ids, item.ItemID)
	}
	return ids
}
