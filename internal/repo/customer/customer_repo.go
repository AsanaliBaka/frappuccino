package customer_repo

import (
	"context"
	"database/sql"

	"frappuccino/internal/models"
)

type CustomerRepo struct {
	DB *sql.DB
}

func NewCustomerRepo(db *sql.DB) *CustomerRepo {
	return &CustomerRepo{
		DB: db,
	}
}

func (c *CustomerRepo) GetCustomerByID(ctx context.Context, id string) (*models.Customer, error) {
	sql := `SELECT * FROM Customers WHERE Customer_ID = $1`

	row := c.DB.QueryRowContext(ctx, sql, id)

	var customer models.Customer
	err := row.Scan(
		&customer.ID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.Preferences,
	)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}
