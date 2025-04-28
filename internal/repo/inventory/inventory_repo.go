package inventory_repo

import (
	"context"
	"database/sql"
	"fmt"

	"frappuccino/internal/models"
)

type InventoryRepo interface {
	GetAllInventory(ctx context.Context) ([]*models.InventoryItem, error)
	GetInventoryByID(ctx context.Context, id string) (*models.InventoryItem, error)
	CreateInventory(ctx context.Context, item *models.InventoryItem) error
	UpdateInventoryByID(ctx context.Context, id string, item *models.InventoryItem) error
	DeleteInventoryByID(ctx context.Context, id string) error
	GetInventoryList(ctx context.Context, sortBy string, page, pageSize int) ([]models.InventoryItem, int, bool, int, error)
}

type inventoryRepo struct {
	DB *sql.DB
}

func NewInventoryRepo(db *sql.DB) InventoryRepo {
	return &inventoryRepo{DB: db}
}

func (r *inventoryRepo) GetAllInventory(ctx context.Context) ([]*models.InventoryItem, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT Inventory_ID, Name, Quantity, Unit, Price FROM Inventory
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		if err := rows.Scan(
			&item.IngredientID,
			&item.Title,
			&item.Stock,
			&item.Measure,
			&item.UnitCost,
		); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *inventoryRepo) GetInventoryByID(ctx context.Context, id string) (*models.InventoryItem, error) {
	var item models.InventoryItem
	err := r.DB.QueryRowContext(ctx,
		`SELECT Inventory_ID, Name, Quantity, Unit, Price FROM Inventory WHERE Inventory_ID = $1`, id).
		Scan(&item.IngredientID, &item.Title, &item.Stock, &item.Measure, &item.UnitCost)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *inventoryRepo) CreateInventory(ctx context.Context, item *models.InventoryItem) error {
	query := `
		INSERT INTO Inventory (Name, Quantity, Unit, Price)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.DB.ExecContext(ctx, query,
		item.Title,
		item.Stock,
		item.Measure,
		item.UnitCost,
	)
	return err
}

func (r *inventoryRepo) UpdateInventoryByID(ctx context.Context, id string, item *models.InventoryItem) error {
	query := `
		UPDATE Inventory 
		SET Name = $1, Quantity = $2, Unit = $3, Price = $4 
		WHERE Inventory_ID = $5
	`
	_, err := r.DB.ExecContext(ctx, query,
		item.Title,
		item.Stock,
		item.Measure,
		item.UnitCost,
		id,
	)
	return err
}

func (r *inventoryRepo) DeleteInventoryByID(ctx context.Context, id string) error {
	query := `DELETE FROM Inventory WHERE Inventory_ID = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *inventoryRepo) GetInventoryList(ctx context.Context, sortBy string, page, pageSize int) ([]models.InventoryItem, int, bool, int, error) {
	var results []models.InventoryItem
	var totalCount int

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	sortColumn := "Quantity"
	if sortBy == "price" {
		sortColumn = "Price"
	}

	query := fmt.Sprintf(`
		SELECT Inventory_ID, Name, Quantity, Unit, Price
		FROM Inventory 
		ORDER BY %s ASC
		LIMIT $1 OFFSET $2
	`, sortColumn)

	rows, err := r.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, false, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.InventoryItem
		if err := rows.Scan(&item.IngredientID, &item.Title, &item.Stock, &item.Measure, &item.UnitCost); err != nil {
			return nil, 0, false, 0, err
		}
		results = append(results, item)
	}

	err = r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM Inventory").Scan(&totalCount)
	if err != nil {
		return nil, 0, false, 0, err
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNextPage := page < totalPages

	return results, page, hasNextPage, totalPages, nil
}
