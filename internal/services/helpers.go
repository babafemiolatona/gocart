package services

import (
	"errors"
	"net/http"

	apperrors "gocart/internal/errors"

	"gorm.io/gorm"
)

func repoErr(
	err error,
	fetchCode, fetchMsg, notFoundCode, notFoundMsg string,
) *apperrors.AppError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.New(http.StatusNotFound, notFoundCode, notFoundMsg, err)
	}

	return apperrors.New(http.StatusInternalServerError, fetchCode, fetchMsg, err)
}
