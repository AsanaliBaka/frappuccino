package order_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"frappuccino/internal/models"
)

func (o *orderRepo) createOrderRecord(ctx context.Context, order *models.Purchase) error {
	query := `
		INSERT INTO Orders (Customer_ID, Status, Total_Amount)
		VALUES ($1, $2, $3)
		RETURNING Order_ID, Created_At, Updated_At
	`

	var id int
	err := o.DB.QueryRowContext(
		ctx,
		query,
		order.CustomerID,
		order.Status,
		order.Amount,
	).Scan(&id, &order.Created, &order.Updated)
	if err != nil {
		return err
	}

	order.PurchaseID = strconv.Itoa(id)
	return nil
}

func (r *orderRepo) insertOrderItems(ctx context.Context, orderID string, items []*models.LineItem) error {
	query := `
		INSERT INTO Order_Items (
			Order_ID,
			Menu_Item_ID,
			Quantity,
			Price,
			Customization
		) VALUES ($1, $2, $3, $4, $5)
	`

	orderIDInt, err := strconv.Atoi(orderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	for _, item := range items {
		menuID, err := strconv.Atoi(item.ItemID)
		if err != nil {
			return fmt.Errorf("invalid item ID: %w", err)
		}

		customization, err := json.Marshal(item.Adjustments)
		if err != nil {
			return fmt.Errorf("marshal customization: %w", err)
		}

		_, err = r.DB.ExecContext(ctx, query,
			orderIDInt,
			menuID,
			item.Count,
			item.UnitPrice,
			customization,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return nil
}

func (r *orderRepo) reserveInventory(ctx context.Context, orderID string, items []*models.LineItem) error {
	orderIDInt, err := strconv.Atoi(orderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	for _, item := range items {
		menuItemID, err := strconv.Atoi(item.ItemID)
		if err != nil {
			return fmt.Errorf("invalid menu item ID: %w", err)
		}

		query := `
			SELECT Inventory_ID, Quantity
			FROM Menu_Item_Ingredients
			WHERE Menu_Item_ID = $1
		`

		rows, err := r.DB.QueryContext(ctx, query, menuItemID)
		if err != nil {
			return fmt.Errorf("query ingredients: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var inventoryID int
			var qtyPerUnit float64

			if err := rows.Scan(&inventoryID, &qtyPerUnit); err != nil {
				return fmt.Errorf("scan ingredient: %w", err)
			}

			reserveQty := qtyPerUnit * float64(item.Count)

			insert := `
				INSERT INTO Inventory_Reservations (Order_ID, Inventory_ID, Reserved_Quantity)
				VALUES ($1, $2, $3)
			`
			_, err = r.DB.ExecContext(ctx, insert, orderIDInt, inventoryID, reserveQty)
			if err != nil {
				return fmt.Errorf("insert reservation: %w", err)
			}
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("read rows: %w", err)
		}
	}

	return nil
}

func (r *orderRepo) removeReserve(ctx context.Context, id string) error {
	orderIDInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, `DELETE FROM Inventory_Reservations WHERE Order_ID = $1`, orderIDInt)
	return err
}

func (r *orderRepo) recordOrderStatus(ctx context.Context, order *models.Purchase) error {
	query := `
		INSERT INTO Order_Status_History (
			Order_ID,
			Status
		) VALUES ($1, $2)
	`

	orderIDInt, err := strconv.Atoi(order.PurchaseID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, query, orderIDInt, order.Status)
	if err != nil {
		return fmt.Errorf("insert status history: %w", err)
	}

	return nil
}

func (r *orderRepo) removeItemsByOrderID(ctx context.Context, id string) error {
	orderIDInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, `DELETE FROM Order_Items WHERE Order_ID = $1`, orderIDInt)
	return err
}
