package errors

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidationError(err error) *AppError {
	return New(http.StatusBadRequest, CodeValidationError, validationMessage(err), err)
}

func validationMessage(err error) string {
	var verr validator.ValidationErrors

	if errors.As(err, &verr) && len(verr) > 0 {
		parts := make([]string, 0, len(verr))

		for _, e := range verr {
			parts = append(parts, e.Field()+" is "+validationTagMessage(e.Tag()))
		}

		return strings.Join(parts, "; ")
	}

	return "invalid request body or parameters"
}

func validationTagMessage(tag string) string {
	switch tag {
	case "required":
		return "required"
	case "email":
		return "invalid"
	case "min":
		return "below the minimum"
	case "max":
		return "above the maximum"
	case "gt":
		return "out of range"
	default:
		return "invalid"
	}
}
