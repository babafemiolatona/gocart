package services

import (
	"errors"
	"testing"

	apperrors "gocart/internal/errors"
)

var errBoom = errors.New("boom")

func assertAppError(t *testing.T, err error, status int, code string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected *apperrors.AppError, got %T (%v)", err, err)
	}

	if appErr.Status != status {
		t.Errorf("expected status %d, got %d", status, appErr.Status)
	}

	if appErr.Code != code {
		t.Errorf("expected code %q, got %q", code, appErr.Code)
	}
}
