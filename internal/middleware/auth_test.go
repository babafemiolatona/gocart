package middleware

import (
	"net/http"
	"testing"

	"gocart/internal/models"
	"gocart/internal/services"
)

func TestAuthMissingHeader(t *testing.T) {
	r, called := newTestRouter(AuthMiddleware(&stubTokenVerifier{}))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "missing_auth_header" {
		t.Errorf("expected missing_auth_header, got %q", code)
	}
	if *called {
		t.Error("handler should not run when unauthorized")
	}
}

func TestAuthEmptyBearerToken(t *testing.T) {
	r, _ := newTestRouter(AuthMiddleware(&stubTokenVerifier{}))

	w := doRequest(r, map[string]string{"Authorization": "Bearer  "})
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "invalid_auth_header" {
		t.Errorf("expected invalid_auth_header, got %q", code)
	}
}

func TestAuthBareTokenString(t *testing.T) {
	var verified string
	verifier := &stubTokenVerifier{
		verifyFn: func(tokenStr string) (*services.CustomClaims, error) {
			verified = tokenStr
			return claimsWith("7", models.RoleCustomer), nil
		},
	}
	r, called := newTestRouter(AuthMiddleware(verifier))

	w := doRequest(r, map[string]string{"Authorization": "some-raw-token"})
	assertStatus(t, w, http.StatusOK)

	if verified != "some-raw-token" {
		t.Errorf("expected raw token to be verified, got %q", verified)
	}
	if !*called {
		t.Error("expected handler to run")
	}
}

func TestAuthInvalidToken(t *testing.T) {
	verifier := &stubTokenVerifier{
		verifyFn: func(tokenStr string) (*services.CustomClaims, error) {
			return nil, errBoom
		},
	}
	r, _ := newTestRouter(AuthMiddleware(verifier))

	w := doRequest(r, map[string]string{"Authorization": "Bearer invalid"})
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "invalid_token" {
		t.Errorf("expected invalid_token, got %q", code)
	}
}

func TestAuthValidTokenSetsContext(t *testing.T) {
	verifier := &stubTokenVerifier{
		verifyFn: func(tokenStr string) (*services.CustomClaims, error) {
			return claimsWith("42", models.RoleAdmin), nil
		},
	}
	r := newTestRouterWithCapture(AuthMiddleware(verifier))

	w := doRequest(r, map[string]string{"Authorization": "Bearer valid-token"})
	assertStatus(t, w, http.StatusOK)

	gotUser, gotRole := capturedContext()
	if gotUser != 42 {
		t.Errorf("expected userID 42, got %v", gotUser)
	}
	if gotRole != models.RoleAdmin {
		t.Errorf("expected role admin, got %v", gotRole)
	}
}

func TestAuthInvalidUserIDSubject(t *testing.T) {
	verifier := &stubTokenVerifier{
		verifyFn: func(tokenStr string) (*services.CustomClaims, error) {
			return claimsWith("not-a-number", models.RoleCustomer), nil
		},
	}
	r, _ := newTestRouter(AuthMiddleware(verifier))

	w := doRequest(r, map[string]string{"Authorization": "Bearer valid-token"})
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "invalid_user_id" {
		t.Errorf("expected invalid_user_id, got %q", code)
	}
}
