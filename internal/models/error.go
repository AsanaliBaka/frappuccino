package models

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrElemExist             = errors.New("element already exists")
	ErrInvalidInput          = errors.New("invalid input")
	ErrInternal              = errors.New("internal error")
	ErrInventoryNotAvailable = errors.New("inventory not available")
)

type Error struct {
	appErr   error
	svcError error
}

func NewError(svcError, appErr error) error {
	return Error{
		svcError: svcError,
		appErr:   appErr,
	}
}

func (e Error) AppError() error {
	return e.appErr
}

func (e Error) SvcError() error {
	return e.svcError
}

func (e Error) Error() string {
	return errors.Join(e.svcError, e.appErr).Error()
}
