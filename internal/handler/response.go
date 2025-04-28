package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"frappuccino/internal/models"
)

func Respond(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

type Error struct {
	Status  int
	Message string
}

func FromError(err error) Error {
	apiError := Error{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	}

	var svcError models.Error
	if errors.As(err, &svcError) {
		if appErr := svcError.AppError(); appErr != nil {
			apiError.Message = appErr.Error()
		}

		switch svcError.SvcError() {
		case models.ErrElemExist:
			apiError.Status = http.StatusConflict
		case models.ErrInvalidInput:
			apiError.Status = http.StatusBadRequest
		case models.ErrNotFound:
			apiError.Status = http.StatusNotFound
		}
	}

	return apiError
}
