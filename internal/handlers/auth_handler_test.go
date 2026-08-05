package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
)

func TestRegisterSuccess(t *testing.T) {
	svc := &stubAuthService{
		registerFn: func(req *dto.RegisterRequest) (*dto.UserResponse, error) {
			return &dto.UserResponse{ID: 1, Email: "chris@example.com", Username: "chris"}, nil
		},
	}

	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/auth/register", NewAuthHandler(svc).Register)

	w := doRequest(t, r, http.MethodPost, "/auth/register", `{"email":"chris@example.com","username":"chris","password":"secret123","confirm_password":"secret123","first_name":"Chris","last_name":"Taylor"}`)
	assertStatus(t, w, http.StatusCreated)

	var resp dto.UserResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.ID != 1 || resp.Email != "chris@example.com" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestRegisterValidationError(t *testing.T) {
	svc := &stubAuthService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/auth/register", NewAuthHandler(svc).Register)

	w := doRequest(t, r, http.MethodPost, "/auth/register", `{"email":"not-an-email"}`)
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "validation_error" {
		t.Errorf("expected validation_error, got %q", code)
	}
}

func TestRegisterServiceError(t *testing.T) {
	svc := &stubAuthService{
		registerFn: func(req *dto.RegisterRequest) (*dto.UserResponse, error) {
			return nil, appErr(http.StatusConflict, "user_exists", "user already exists")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/auth/register", NewAuthHandler(svc).Register)

	w := doRequest(t, r, http.MethodPost, "/auth/register", `{"email":"a@b.com","username":"chris","password":"secret123","confirm_password":"secret123","first_name":"Chris","last_name":"Taylor"}`)
	assertStatus(t, w, http.StatusConflict)

	code, message := decodeError(t, w)
	if code != "user_exists" || message != "user already exists" {
		t.Errorf("unexpected error: %q %q", code, message)
	}
}

func TestLoginSuccess(t *testing.T) {
	svc := &stubAuthService{
		loginFn: func(req *dto.LoginRequest) (*dto.AuthResponse, error) {
			return &dto.AuthResponse{Token: "abc", ExpiresAt: 1234}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/auth/login", NewAuthHandler(svc).Login)

	w := doRequest(t, r, http.MethodPost, "/auth/login", `{"username_or_email":"chris@example.com","password":"secret123"}`)
	assertStatus(t, w, http.StatusOK)

	var resp dto.AuthResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.Token != "abc" || resp.ExpiresAt != 1234 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestLoginValidationError(t *testing.T) {
	svc := &stubAuthService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/auth/login", NewAuthHandler(svc).Login)

	w := doRequest(t, r, http.MethodPost, "/auth/login", `{}`)
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "validation_error" {
		t.Errorf("expected validation_error, got %q", code)
	}
}

func TestLoginUnauthorized(t *testing.T) {
	svc := &stubAuthService{
		loginFn: func(req *dto.LoginRequest) (*dto.AuthResponse, error) {
			return nil, appErr(http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/auth/login", NewAuthHandler(svc).Login)

	w := doRequest(t, r, http.MethodPost, "/auth/login", `{"username_or_email":"chris","password":"wrong"}`)
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "invalid_credentials" {
		t.Errorf("expected invalid_credentials, got %q", code)
	}
}
