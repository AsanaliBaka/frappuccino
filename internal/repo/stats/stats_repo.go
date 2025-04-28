package stats_repo

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"frappuccino/internal/models"

	"github.com/lib/pq"
)

type StatsRepo interface {
	GetPopularItem(ctx context.Context) ([]models.TopProduct, error)
	GetTotalSum(ctx context.Context) (*models.RevenueSummary, error)
	SearchMenu(ctx context.Context, query string, minPrice, maxPrice *float64) ([]models.ProductPreview, error)
	SearchOrders(ctx context.Context, query string, minPrice, maxPrice *float64) ([]models.OrderBrief, error)
	GetItemByPeriod(ctx context.Context, period string, month string, year int) ([]models.OrderStats, error)
}

type statsRepo struct {
	DB *sql.DB
}

func NewStatsRepo(db *sql.DB) StatsRepo {
	return &statsRepo{
		DB: db,
	}
}

func (m *statsRepo) GetPopularItem(ctx context.Context) ([]models.TopProduct, error) {
	query := `
		SELECT 
			oi.Menu_Item_ID, 
			mi.Name, 
			SUM(oi.Quantity)::FLOAT AS total_quantity
		FROM Order_Items oi
		JOIN Menu_Items mi ON oi.Menu_Item_ID = mi.Menu_Item_ID
		WHERE oi.Menu_Item_ID IS NOT NULL
		GROUP BY oi.Menu_Item_ID, mi.Name
		ORDER BY total_quantity DESC
		LIMIT 10;
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.TopProduct
	for rows.Next() {
		var item models.TopProduct
		if err := rows.Scan(&item.ItemID, &item.ItemName, &item.Sold); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (m *statsRepo) GetTotalSum(ctx context.Context) (*models.RevenueSummary, error) {
	totalSales := &models.RevenueSummary{}
	query := `SELECT COALESCE(SUM(total_amount), 0) FROM orders`
	err := m.DB.QueryRowContext(ctx, query).Scan(&totalSales.Sum)
	if err != nil {
		if err == sql.ErrNoRows {
			totalSales.Sum = 0
			return totalSales, nil
		}
		return nil, err
	}

	return totalSales, nil
}

func (m *statsRepo) SearchMenu(ctx context.Context, query string, minPrice, maxPrice *float64) ([]models.ProductPreview, error) {
	var results []models.ProductPreview
	baseQuery := `
        SELECT Menu_Item_ID AS id, Name, Description, Price
        FROM Menu_Items
        WHERE (Name ILIKE $1 OR Description ILIKE $1) 
    `
	params := []interface{}{"%" + query + "%"}
	conditions := ""

	if minPrice != nil {
		conditions += " AND Price >= $" + strconv.Itoa(len(params)+1)
		params = append(params, *minPrice)
	}
	if maxPrice != nil {
		conditions += " AND Price <= $" + strconv.Itoa(len(params)+1)
		params = append(params, *maxPrice)
	}

	baseQuery += conditions + " ORDER BY Name"

	rows, err := m.DB.QueryContext(ctx, baseQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.ProductPreview
		if err := rows.Scan(&item.ItemID, &item.Title, &item.Info, &item.Cost); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

func (m *statsRepo) SearchOrders(ctx context.Context, query string, minPrice, maxPrice *float64) ([]models.OrderBrief, error) {
	var results []models.OrderBrief
	baseQuery := `
        SELECT 
            o.Order_ID AS id, 
            c.Name AS customer_name, 
            ARRAY_AGG(mi.Name) AS items, 
            o.Total_Amount AS total
        FROM Orders o
        JOIN Customers c ON o.Customer_ID = c.Customer_ID
        JOIN Order_Items oi ON oi.Order_ID = o.Order_ID
        JOIN Menu_Items mi ON mi.Menu_Item_ID = oi.Menu_Item_ID
        WHERE 
            (c.Name ILIKE $1 OR mi.Name ILIKE $1) 
    `
	params := []interface{}{"%" + query + "%"}
	conditions := ""

	if minPrice != nil {
		conditions += " AND o.Total_Amount >= $" + strconv.Itoa(len(params)+1)
		params = append(params, *minPrice)
	}
	if maxPrice != nil {
		conditions += " AND o.Total_Amount <= $" + strconv.Itoa(len(params)+1)
		params = append(params, *maxPrice)
	}

	baseQuery += conditions + " GROUP BY o.Order_ID, c.Name ORDER BY o.Created_At DESC"

	rows, err := m.DB.QueryContext(ctx, baseQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.OrderBrief
		var items pq.StringArray
		if err := rows.Scan(&order.OrderID, &order.Client, &items, &order.Amount); err != nil {
			return nil, err
		}
		order.Items = items
		results = append(results, order)
	}

	return results, nil
}

func (m *statsRepo) GetItemByPeriod(ctx context.Context, period string, month string, year int) ([]models.OrderStats, error) {
	var query string
	var rows *sql.Rows
	var err error

	if period == "day" {
		query = `
			SELECT DATE_TRUNC('day', Created_At) AS period, COUNT(*) AS total_orders
			FROM Orders
			WHERE EXTRACT(MONTH FROM Created_At) = EXTRACT(MONTH FROM TO_DATE($1, 'Month'))
			AND EXTRACT(YEAR FROM Created_At) = $2
			GROUP BY period
			ORDER BY period;
		`
		rows, err = m.DB.QueryContext(ctx, query, month, year)
	} else if period == "month" {
		query = `
			SELECT DATE_TRUNC('month', Created_At) AS period, COUNT(*) AS total_orders
			FROM Orders
			WHERE EXTRACT(YEAR FROM Created_At) = $1
			GROUP BY period
			ORDER BY period;
		`
		rows, err = m.DB.QueryContext(ctx, query, year)
	} else {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.OrderStats
	for rows.Next() {
		var report models.OrderStats
		if err := rows.Scan(&report.Date, &report.Count); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}
