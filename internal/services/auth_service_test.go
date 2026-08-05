package services

import (
	"net/http"
	"testing"
	"time"

	"gocart/internal/config"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

func newTestAuthService(repo *stubAuthRepo) *AuthService {
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExpiry: time.Hour,
	}
	return NewAuthService(repo, cfg)
}

func TestRegisterSuccess(t *testing.T) {
	repo := &stubAuthRepo{}
	svc := newTestAuthService(repo)

	resp, err := svc.Register(&dto.RegisterRequest{
		Email:           "  CHRIS@Example.COM ",
		Username:        "chris",
		Password:        "secret123",
		ConfirmPassword: "secret123",
		FirstName:       "Chris",
		LastName:        "Taylor",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Email != "chris@example.com" {
		t.Errorf("email was not lowercased/trimmed: %q", resp.Email)
	}
	if resp.Username != "chris" || resp.FirstName != "Chris" || resp.LastName != "Taylor" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.Role != models.RoleCustomer {
		t.Errorf("expected customer role, got %q", resp.Role)
	}
}

func TestRegisterValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *dto.RegisterRequest
	}{
		{"missing email", &dto.RegisterRequest{Password: "secret123", ConfirmPassword: "secret123", FirstName: "A", LastName: "B"}},
		{"bad email format", &dto.RegisterRequest{Email: "not-an-email", Password: "secret123", ConfirmPassword: "secret123", FirstName: "A", LastName: "B"}},
		{"short password", &dto.RegisterRequest{Email: "a@b.com", Password: "abc", ConfirmPassword: "abc", FirstName: "A", LastName: "B"}},
		{"mismatched passwords", &dto.RegisterRequest{Email: "a@b.com", Password: "secret123", ConfirmPassword: "different", FirstName: "A", LastName: "B"}},
		{"missing names", &dto.RegisterRequest{Email: "a@b.com", Password: "secret123", ConfirmPassword: "secret123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(&stubAuthRepo{})
			_, err := svc.Register(tt.req)
			if err == nil {
				t.Fatal("expected validation error")
			}
			appErr, ok := err.(*apperrors.AppError)
			if !ok {
				t.Fatalf("expected AppError, got %T", err)
			}
			if appErr.Status != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", appErr.Status)
			}
		})
	}
}

func TestRegisterUserAlreadyExists(t *testing.T) {
	repo := &stubAuthRepo{
		existsFn: func(email string) (bool, error) { return true, nil },
	}
	svc := newTestAuthService(repo)

	_, err := svc.Register(&dto.RegisterRequest{
		Email: "chris@example.com", Username: "chris",
		Password: "secret123", ConfirmPassword: "secret123",
		FirstName: "Chris", LastName: "Taylor",
	})

	assertAppError(t, err, http.StatusConflict, apperrors.CodeUserExists)
}

func TestRegisterDuplicateOnCreate(t *testing.T) {
	repo := &stubAuthRepo{
		createFn: func(user *models.User) error { return repositories.ErrDuplicate },
	}
	svc := newTestAuthService(repo)

	_, err := svc.Register(&dto.RegisterRequest{
		Email: "chris@example.com", Username: "chris",
		Password: "secret123", ConfirmPassword: "secret123",
		FirstName: "Chris", LastName: "Taylor",
	})

	assertAppError(t, err, http.StatusConflict, apperrors.CodeUserExists)
}

func TestRegisterExistsCheckFails(t *testing.T) {
	repo := &stubAuthRepo{
		existsFn: func(email string) (bool, error) { return false, errBoom },
	}
	svc := newTestAuthService(repo)

	_, err := svc.Register(&dto.RegisterRequest{
		Email: "chris@example.com", Username: "chris",
		Password: "secret123", ConfirmPassword: "secret123",
		FirstName: "Chris", LastName: "Taylor",
	})

	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeCheckUser)
}

func TestRegisterHashesPassword(t *testing.T) {
	var created *models.User
	repo := &stubAuthRepo{
		createFn: func(u *models.User) error { created = u; return nil },
	}
	svc := newTestAuthService(repo)

	_, err := svc.Register(&dto.RegisterRequest{
		Email: "chris@example.com", Username: "chris",
		Password: "secret123", ConfirmPassword: "secret123",
		FirstName: "Chris", LastName: "Taylor",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.Password == "" || created.Password == "secret123" {
		t.Errorf("password was not hashed: %q", created.Password)
	}
	if !created.VerifyPassword("secret123") {
		t.Errorf("hashed password does not verify")
	}
}

func TestLoginSuccess(t *testing.T) {
	user := &models.User{
		ID:       1,
		Email:    "chris@example.com",
		Username: "chris",
		Role:     models.RoleCustomer,
	}
	_ = user.HashPassword("secret123")

	repo := &stubAuthRepo{
		getByIdentifierFn: func(identifier string) (*models.User, error) {
			if identifier != "chris@example.com" {
				t.Errorf("expected lowercased identifier, got %q", identifier)
			}
			return user, nil
		},
	}
	svc := newTestAuthService(repo)

	resp, err := svc.Login(&dto.LoginRequest{
		UsernameOrEmail: "  CHRIS@Example.COM ",
		Password:        "secret123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected a token")
	}
	if resp.ExpiresAt <= time.Now().Unix() {
		t.Errorf("expected future expiry, got %d", resp.ExpiresAt)
	}
}

func TestLoginUnknownUser(t *testing.T) {
	repo := &stubAuthRepo{}
	svc := newTestAuthService(repo)

	_, err := svc.Login(&dto.LoginRequest{UsernameOrEmail: "nobody@example.com", Password: "secret123"})
	assertAppError(t, err, http.StatusUnauthorized, apperrors.CodeInvalidCredentials)
}

func TestLoginWrongPassword(t *testing.T) {
	user := &models.User{ID: 1, Email: "chris@example.com"}
	_ = user.HashPassword("right-password")

	repo := &stubAuthRepo{
		getByIdentifierFn: func(identifier string) (*models.User, error) { return user, nil },
	}
	svc := newTestAuthService(repo)

	_, err := svc.Login(&dto.LoginRequest{UsernameOrEmail: "chris@example.com", Password: "wrong-password"})
	assertAppError(t, err, http.StatusUnauthorized, apperrors.CodeInvalidCredentials)
}

func TestGenerateAndVerifyTokenRoundTrip(t *testing.T) {
	svc := newTestAuthService(&stubAuthRepo{})
	user := &models.User{ID: 42, Role: models.RoleAdmin}

	token, _, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := svc.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}

	if claims.Subject != "42" {
		t.Errorf("expected subject 42, got %q", claims.Subject)
	}
	if claims.Role != models.RoleAdmin {
		t.Errorf("expected admin role, got %q", claims.Role)
	}
}

func TestVerifyTokenRejectsWrongSecret(t *testing.T) {
	svcA := newTestAuthService(&stubAuthRepo{})
	svcB := NewAuthService(&stubAuthRepo{}, &config.Config{JWTSecret: "other-secret", JWTExpiry: time.Hour})

	token, _, err := svcA.GenerateToken(&models.User{ID: 1, Role: models.RoleCustomer})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if _, err := svcB.VerifyToken(token); err == nil {
		t.Error("expected verification to fail with a different secret")
	}
}

func TestVerifyTokenRejectsGarbage(t *testing.T) {
	svc := newTestAuthService(&stubAuthRepo{})

	if _, err := svc.VerifyToken("not-a-token"); err == nil {
		t.Error("expected error for invalid token")
	}
}
