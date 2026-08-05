package handlers

import (
	"errors"

	apperrors "gocart/internal/errors"
)

var errBoom = errors.New("boom")

func appErr(status int, code, message string) *apperrors.AppError {
	return apperrors.New(status, code, message, nil)
}
