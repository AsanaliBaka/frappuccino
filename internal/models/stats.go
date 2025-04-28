package models

import "time"

type RevenueSummary struct {
	Sum float64 `json:"total_amount"`
}

type TopProduct struct {
	ItemID   string  `json:"product_id"`
	ItemName string  `json:"name"`
	Sold     float64 `json:"quantity"`
}

type ProductPreview struct {
	ItemID string  `json:"id"`
	Title  string  `json:"name"`
	Info   string  `json:"description"`
	Cost   float64 `json:"price"`
}

type OrderBrief struct {
	OrderID string   `json:"id"`
	Client  string   `json:"customer_name"`
	Items   []string `json:"items"`
	Amount  float64  `json:"total"`
}

type LookupResult struct {
	Products     []ProductPreview `json:"menu_items"`
	RecentOrders []OrderBrief     `json:"orders"`
	MatchesFound int              `json:"total_matches"`
}

type OrderStats struct {
	Date  time.Time
	Count int
}
