package order_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"frappuccino/internal/models"

	"github.com/lib/pq"
)

type OrderRepo interface {
	GetAllOrders(ctx context.Context) ([]*models.Purchase, error)
	GetOrderByID(ctx context.Context, id string) (*models.Purchase, error)
	CreateOrder(ctx context.Context, order *models.Purchase) error
	UpdateOrder(ctx context.Context, id string, order *models.Purchase) error
	DeleteOrder(ctx context.Context, id string) error
	CheckInventoryForOrder(ctx context.Context, order *models.Purchase) (bool, error)
	CloseOrder(ctx context.Context, order *models.Purchase) error
	GetNumberOfOrderedItems(ctx context.Context, startDate, endDate *time.Time) (map[string]int, error)
}

type orderRepo struct {
	DB *sql.DB
}

func NewOrderRepo(db *sql.DB) OrderRepo {
	return &orderRepo{
		DB: db,
	}
}

func (r *orderRepo) GetAllOrders(ctx context.Context) ([]*models.Purchase, error) {
	query := `
		SELECT o.Order_ID, o.Customer_ID, o.Status, o.Total_Amount, o.Created_At, o.Updated_At,
		       oi.Order_Item_ID, oi.Menu_Item_ID, oi.Quantity, oi.Price, oi.Customization
		FROM Orders o
		LEFT JOIN Order_Items oi ON o.Order_ID = oi.Order_ID
		ORDER BY o.Created_At DESC, oi.Order_Item_ID
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	ordersMap := make(map[string]*models.Purchase)
	var orders []*models.Purchase

	for rows.Next() {
		var orderID int
		var temp models.Purchase
		var itemID, productID sql.NullInt64
		var quantity, price sql.NullFloat64
		var customization sql.NullString

		err := rows.Scan(
			&orderID,
			&temp.CustomerID,
			&temp.Status,
			&temp.Amount,
			&temp.Created,
			&temp.Updated,
			&itemID,
			&productID,
			&quantity,
			&price,
			&customization,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order row: %w", err)
		}

		orderKey := strconv.Itoa(orderID)
		temp.PurchaseID = orderKey

		existingOrder, exists := ordersMap[orderKey]
		if !exists {
			temp.Positions = []*models.LineItem{}
			ordersMap[orderKey] = &temp
			orders = append(orders, &temp)
			existingOrder = &temp
		}

		if itemID.Valid {
			var customMap models.ConfigMap
			if customization.Valid {
				if err := json.Unmarshal([]byte(customization.String), &customMap); err != nil {
					return nil, fmt.Errorf("unmarshal customization: %w", err)
				}
			}

			existingOrder.Positions = append(existingOrder.Positions, &models.LineItem{
				ItemID:      strconv.FormatInt(productID.Int64, 10),
				Count:       int(quantity.Float64),
				UnitPrice:   price.Float64,
				Adjustments: customMap,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

func (o *orderRepo) CreateOrder(ctx context.Context, order *models.Purchase) error {
	if err := o.createOrderRecord(ctx, order); err != nil {
		return fmt.Errorf("create order record: %w", err)
	}

	if err := o.insertOrderItems(ctx, order.PurchaseID, order.Positions); err != nil {
		return fmt.Errorf("insert order items: %w", err)
	}

	if err := o.reserveInventory(ctx, order.PurchaseID, order.Positions); err != nil {
		return fmt.Errorf("reserve inventory: %w", err)
	}

	if err := o.recordOrderStatus(ctx, order); err != nil {
		return fmt.Errorf("record order status: %w", err)
	}

	return nil
}

func (r *orderRepo) GetOrderByID(ctx context.Context, id string) (*models.Purchase, error) {
	orderIDInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	query := `
		SELECT o.Order_ID, o.Customer_ID, o.Status, o.Total_Amount, o.Created_At, o.Updated_At,
		       oi.Order_Item_ID, oi.Menu_Item_ID, oi.Quantity, oi.Price, oi.Customization
		FROM Orders o
		LEFT JOIN Order_Items oi ON o.Order_ID = oi.Order_ID
		WHERE o.Order_ID = $1
	`

	rows, err := r.DB.QueryContext(ctx, query, orderIDInt)
	if err != nil {
		return nil, fmt.Errorf("query order by ID: %w", err)
	}
	defer rows.Close()

	var order *models.Purchase
	var items []*models.LineItem

	for rows.Next() {
		var temp models.Purchase
		var itemID, productID sql.NullInt64
		var quantity, price sql.NullFloat64
		var customization sql.NullString

		err := rows.Scan(
			&temp.PurchaseID,
			&temp.CustomerID,
			&temp.Status,
			&temp.Amount,
			&temp.Created,
			&temp.Updated,
			&itemID,
			&productID,
			&quantity,
			&price,
			&customization,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order row: %w", err)
		}

		if order == nil {
			temp.PurchaseID = strconv.Itoa(orderIDInt)
			order = &temp
			order.Positions = []*models.LineItem{}
		}

		if itemID.Valid {
			var customMap models.ConfigMap
			if customization.Valid {
				if err := json.Unmarshal([]byte(customization.String), &customMap); err != nil {
					return nil, fmt.Errorf("unmarshal customization: %w", err)
				}
			}

			items = append(items, &models.LineItem{
				ItemID:      strconv.FormatInt(productID.Int64, 10),
				Count:       int(quantity.Float64),
				UnitPrice:   price.Float64,
				Adjustments: customMap,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order rows: %w", err)
	}

	if order == nil {
		return nil, sql.ErrNoRows
	}

	order.Positions = items
	return order, nil
}

func (r *orderRepo) UpdateOrder(ctx context.Context, id string, order *models.Purchase) error {
	orderIDInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	if err := r.removeItemsByOrderID(ctx, id); err != nil {
		return fmt.Errorf("remove order items: %w", err)
	}

	if err := r.insertOrderItems(ctx, id, order.Positions); err != nil {
		return fmt.Errorf("insert new items: %w", err)
	}

	if err := r.removeReserve(ctx, id); err != nil {
		return fmt.Errorf("remove old reserve: %w", err)
	}

	if err := r.reserveInventory(ctx, id, order.Positions); err != nil {
		return fmt.Errorf("reserve new inventory: %w", err)
	}

	query := `
		UPDATE Orders
		SET Customer_ID = $1,
			Total_Amount = $2,
			Updated_At = NOW()
		WHERE Order_ID = $3
	`

	_, err = r.DB.ExecContext(ctx, query,
		order.CustomerID,
		order.Amount,
		orderIDInt,
	)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	return nil
}

func (r *orderRepo) DeleteOrder(ctx context.Context, id string) error {
	orderIDInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, `DELETE FROM Order_Items WHERE Order_ID = $1`, orderIDInt)
	if err != nil {
		return fmt.Errorf("delete order items: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, `DELETE FROM Orders WHERE Order_ID = $1`, orderIDInt)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}

	if err := r.removeReserve(ctx, id); err != nil {
		return fmt.Errorf("remove reserve: %w", err)
	}

	return nil
}

func (r *orderRepo) CheckInventoryForOrder(ctx context.Context, order *models.Purchase) (bool, error) {
	type pair struct {
		MenuItemID int
		Quantity   int
	}

	var menuPairs []pair

	for _, item := range order.Positions {
		id, err := strconv.Atoi(item.ItemID)
		if err != nil {
			return false, fmt.Errorf("invalid menu item ID: %w", err)
		}
		menuPairs = append(menuPairs, pair{MenuItemID: id, Quantity: item.Count})
	}

	// Подготовим массивы для передачи в UNNEST
	menuIDs := make([]int, len(menuPairs))
	quantities := make([]int, len(menuPairs))
	for i, p := range menuPairs {
		menuIDs[i] = p.MenuItemID
		quantities[i] = p.Quantity
	}

	query := `
		WITH Needed_Ingredients AS (
			SELECT 
				mii.Inventory_ID, 
				SUM(mii.Quantity * q.qty) AS Required_Quantity
			FROM UNNEST($1::int[], $2::int[]) WITH ORDINALITY AS q(menu_id, qty, ord)
			JOIN Menu_Item_Ingredients mii ON mii.Menu_Item_ID = q.menu_id
			GROUP BY mii.Inventory_ID
		),
		Available_Stock AS (
			SELECT 
				i.Inventory_ID,
				i.Quantity - COALESCE(SUM(ir.Reserved_Quantity), 0) AS Available_Quantity
			FROM Inventory i
			LEFT JOIN Inventory_Reservations ir ON ir.Inventory_ID = i.Inventory_ID
			GROUP BY i.Inventory_ID, i.Quantity
		)
		SELECT ni.Inventory_ID, ni.Required_Quantity, a.Available_Quantity
		FROM Needed_Ingredients ni
		JOIN Available_Stock a ON a.Inventory_ID = ni.Inventory_ID
		WHERE ni.Required_Quantity > a.Available_Quantity
	`

	rows, err := r.DB.QueryContext(ctx, query, pq.Array(menuIDs), pq.Array(quantities))
	if err != nil {
		return false, fmt.Errorf("check inventory: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return false, nil // Недостаточно хотя бы одного ингредиента
	}

	return true, nil
}

func (r *orderRepo) CloseOrder(ctx context.Context, order *models.Purchase) error {
	orderIDInt, err := strconv.Atoi(order.PurchaseID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	query := `
		UPDATE Inventory
		SET Quantity = Quantity - ir.Reserved_Quantity
		FROM Inventory_Reservations ir
		WHERE ir.Order_ID = $1 AND Inventory.Inventory_ID = ir.Inventory_ID
	`
	_, err = r.DB.ExecContext(ctx, query, orderIDInt)
	if err != nil {
		return fmt.Errorf("deduct inventory: %w", err)
	}

	query = `
		UPDATE Orders
		SET Status = 'completed', Updated_At = NOW()
		WHERE Order_ID = $1
	`
	_, err = r.DB.ExecContext(ctx, query, orderIDInt)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if err := r.removeReserve(ctx, order.PurchaseID); err != nil {
		return fmt.Errorf("remove reserve: %w", err)
	}

	if err := r.recordOrderStatus(ctx, order); err != nil {
		return fmt.Errorf("record status history: %w", err)
	}

	return nil
}

func (r *orderRepo) GetNumberOfOrderedItems(ctx context.Context, startDate, endDate *time.Time) (map[string]int, error) {
	results := make(map[string]int)

	query := `
		SELECT 
			mi.Name AS menu_item_name, 
			COUNT(*) AS order_count
		FROM Order_Items oi
		JOIN Menu_Items mi ON oi.Menu_Item_ID = mi.Menu_Item_ID
		WHERE oi.Created_At BETWEEN $1 AND $2
		GROUP BY mi.Name
		ORDER BY order_count DESC
	`

	rows, err := r.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query ordered items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, fmt.Errorf("scan ordered item: %w", err)
		}
		results[name] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}
