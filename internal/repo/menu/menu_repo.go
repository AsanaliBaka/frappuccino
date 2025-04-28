package menu_repo

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"frappuccino/internal/models"

	"github.com/lib/pq"
)

type MenuRepo interface {
	GetAllProducts(ctx context.Context) (products []*models.Product, err error)
	FetchProductsByIDs(ctx context.Context, ids []string) (products []*models.Product, err error)
	CreateProduct(ctx context.Context, product *models.Product) (err error)
	GetProductByID(ctx context.Context, id string) (product *models.Product, err error)
	UpdateProduct(ctx context.Context, id string, product *models.Product) (err error)
	UpdatePrice(ctx context.Context, ph *models.PriceHistory) (err error)
	DeleteProduct(ctx context.Context, id string) (err error)
}

type menuRepo struct {
	DB *sql.DB
}

func NewMenuRepo(db *sql.DB) MenuRepo {
	return &menuRepo{
		DB: db,
	}
}

func (m *menuRepo) GetAllProducts(ctx context.Context) ([]*models.Product, error) {
	query := `
		SELECT 
			Menu_Item_ID, 
			Name, 
			Description, 
			Price, 
			Size, 
			Category, 
			Tags, 
			Metadata 
		FROM Menu_Items
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product

	for rows.Next() {
		var (
			id   int
			item models.Product
		)

		err := rows.Scan(
			&id,
			&item.Title,
			&item.Details,
			&item.UnitPrice,
			&item.SizeLabel,
			&item.Group,
			&item.Labels,
			&item.Extras,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}

		item.ProductID = strconv.Itoa(id)
		item.Components = []*models.ProductComponent{}

		if err := m.loadProductComponents(ctx, &item); err != nil {
			return nil, fmt.Errorf("load components: %w", err)
		}

		products = append(products, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (m *menuRepo) FetchProductsByIDs(ctx context.Context, ids []string) ([]*models.Product, error) {
	query := `
		SELECT 
			Menu_Item_ID, 
			Name, 
			Description, 
			Price, 
			Size, 
			Category, 
			Tags, 
			Metadata 
		FROM Menu_Items
		WHERE Menu_Item_ID = ANY($1)
	`

	intIDs := make([]int, len(ids))
	for i, s := range ids {
		id, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid menu item ID: %w", err)
		}
		intIDs[i] = id
	}

	rows, err := m.DB.QueryContext(ctx, query, pq.Array(intIDs))
	if err != nil {
		return nil, fmt.Errorf("query products by ids: %w", err)
	}
	defer rows.Close()

	var products []*models.Product

	for rows.Next() {
		var (
			id   int
			item models.Product
		)

		err := rows.Scan(
			&id,
			&item.Title,
			&item.Details,
			&item.UnitPrice,
			&item.SizeLabel,
			&item.Group,
			&item.Labels,
			&item.Extras,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}

		item.ProductID = strconv.Itoa(id)
		item.Components = []*models.ProductComponent{}

		if err := m.loadProductComponents(ctx, &item); err != nil {
			return nil, fmt.Errorf("load components: %w", err)
		}

		products = append(products, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (m *menuRepo) CreateProduct(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO Menu_Items (
			Name, Description, Price, Size, Category, Tags, Metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING Menu_Item_ID
	`

	var newID int
	err := m.DB.QueryRowContext(ctx, query,
		product.Title,
		product.Details,
		product.UnitPrice,
		product.SizeLabel,
		product.Group,
		product.Labels,
		product.Extras,
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("insert menu item: %w", err)
	}
	product.ProductID = strconv.Itoa(newID)

	insertComponent := `
		INSERT INTO Menu_Item_Ingredients (
			Menu_Item_ID, Inventory_ID, Quantity
		) VALUES ($1, $2, $3)
	`

	for _, comp := range product.Components {
		invID, err := strconv.Atoi(comp.ComponentID)
		if err != nil {
			return fmt.Errorf("convert component ID: %w", err)
		}

		_, err = m.DB.ExecContext(ctx, insertComponent, newID, invID, comp.RequiredQty)
		if err != nil {
			return fmt.Errorf("insert ingredient: %w", err)
		}
	}

	return nil
}

func (m *menuRepo) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT Menu_Item_ID, Name, Description, Price, Size, Category, Tags, Metadata
		FROM Menu_Items
		WHERE Menu_Item_ID = $1
	`

	var prod models.Product
	var prodID int

	err = m.DB.QueryRowContext(ctx, query, intID).Scan(
		&prodID,
		&prod.Title,
		&prod.Details,
		&prod.UnitPrice,
		&prod.SizeLabel,
		&prod.Group,
		&prod.Labels,
		&prod.Extras,
	)
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}
	prod.ProductID = strconv.Itoa(prodID)
	prod.Components = []*models.ProductComponent{}

	err = m.loadProductComponents(ctx, &prod)
	if err != nil {
		return nil, fmt.Errorf("load components: %w", err)
	}

	return &prod, nil
}

func (m *menuRepo) UpdateProduct(ctx context.Context, id string, product *models.Product) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		UPDATE Menu_Items
		SET Name = $1, Description = $2, Price = $3, Size = $4, Category = $5, Tags = $6, Metadata = $7
		WHERE Menu_Item_ID = $8
	`
	_, err = m.DB.ExecContext(ctx, query,
		product.Title,
		product.Details,
		product.UnitPrice,
		product.SizeLabel,
		product.Group,
		product.Labels,
		product.Extras,
		intID,
	)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	_, err = m.DB.ExecContext(ctx, `DELETE FROM Menu_Item_Ingredients WHERE Menu_Item_ID = $1`, intID)
	if err != nil {
		return fmt.Errorf("delete old components: %w", err)
	}

	insertQuery := `
		INSERT INTO Menu_Item_Ingredients (Menu_Item_ID, Inventory_ID, Quantity)
		VALUES ($1, $2, $3)
	`
	for _, comp := range product.Components {
		compID, err := strconv.Atoi(comp.ComponentID)
		if err != nil {
			return fmt.Errorf("convert component ID: %w", err)
		}
		_, err = m.DB.ExecContext(ctx, insertQuery, intID, compID, comp.RequiredQty)
		if err != nil {
			return fmt.Errorf("insert component: %w", err)
		}
	}

	return nil
}

func (m *menuRepo) UpdatePrice(ctx context.Context, ph *models.PriceHistory) error {
	intID, err := strconv.Atoi(ph.MenuItemID)
	if err != nil {
		return fmt.Errorf("invalid menu item ID: %w", err)
	}

	query := `
		INSERT INTO Price_History (
			Menu_Item_ID,
			Old_Price,
			New_Price
		)
		VALUES ($1, $2, $3)
	`
	_, err = m.DB.ExecContext(ctx, query, intID, ph.OldPrice, ph.NewPrice)
	if err != nil {
		return fmt.Errorf("insert price history: %w", err)
	}

	return nil
}

func (m *menuRepo) DeleteProduct(ctx context.Context, id string) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	query := `DELETE FROM Menu_Items WHERE Menu_Item_ID = $1`
	_, err = m.DB.ExecContext(ctx, query, intID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	return nil
}

func (m *menuRepo) loadProductComponents(ctx context.Context, product *models.Product) error {
	query := `
		SELECT 
			i.Inventory_ID, 
			i.Name,
			m.Quantity 
		FROM Menu_Item_Ingredients m
		JOIN Inventory i ON m.Inventory_ID = i.Inventory_ID
		WHERE m.Menu_Item_ID = $1
	`

	menuID, err := strconv.Atoi(product.ProductID)
	if err != nil {
		return fmt.Errorf("convert ProductID: %w", err)
	}

	rows, err := m.DB.QueryContext(ctx, query, menuID)
	if err != nil {
		return fmt.Errorf("query product components: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var comp models.ProductComponent
		var inventoryID int

		err := rows.Scan(&inventoryID, &comp.ComponentName, &comp.RequiredQty)
		if err != nil {
			return fmt.Errorf("scan component: %w", err)
		}

		comp.ComponentID = strconv.Itoa(inventoryID)
		product.Components = append(product.Components, &comp)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	return nil
}
