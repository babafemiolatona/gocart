package services

import (
	"net/http"
	"testing"

	apperrors "gocart/internal/errors"
	"gocart/internal/models"
)

func TestGetMeSuccess(t *testing.T) {
	repo := &stubUserRepo{
		getByIDFn: func(id uint) (*models.User, error) {
			return &models.User{ID: 1, Username: "chris", Email: "chris@example.com", Role: models.RoleCustomer}, nil
		},
	}
	svc := NewUserService(repo)

	resp, err := svc.GetMe(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != 1 || resp.Username != "chris" || resp.Email != "chris@example.com" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetMeNotFound(t *testing.T) {
	svc := NewUserService(&stubUserRepo{})

	_, err := svc.GetMe(99)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeUserNotFound)
}
