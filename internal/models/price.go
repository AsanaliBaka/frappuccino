package models

type PriceHistory struct {
	ID         string  `json:"price_id"`
	MenuItemID string  `json:"product_id"`
	OldPrice   float64 `json:"old_price"`
	NewPrice   float64 `json:"new_price"`
	ChangeDate string  `json:"change_date"`
}
