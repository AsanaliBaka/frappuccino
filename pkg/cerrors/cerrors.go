package cerrors

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotExist      = errors.New("not exist")
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
	ErrForeignKey    = errors.New("foreign key violation")
)

func IsNotFoundError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func IsAlreadyExistsError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func IsForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func MapSQLError(err error) error {
	if IsNotFoundError(err) {
		return ErrNotFound
	}
	if IsAlreadyExistsError(err) {
		return ErrAlreadyExists
	}
	if IsForeignKeyError(err) {
		return ErrForeignKey
	}
	return err
}
