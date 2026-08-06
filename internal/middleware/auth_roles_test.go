package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"

	"gocart/internal/models"
)

func TestRequireRoleAllowed(t *testing.T) {
	r, called := newTestRouter(func(c *gin.Context) {
		c.Set("userRole", models.RoleAdmin)
		c.Next()
	}, RequireRole(models.RoleAdmin))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusOK)
	if !*called {
		t.Error("expected handler to run for allowed role")
	}
}

func TestRequireRoleMissingRole(t *testing.T) {
	r, _ := newTestRouter(RequireRole(models.RoleAdmin))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "unauthorized" {
		t.Errorf("expected unauthorized, got %q", code)
	}
}

func TestRequireRoleWrongType(t *testing.T) {
	r, _ := newTestRouter(func(c *gin.Context) {
		c.Set("userRole", "not-a-role")
		c.Next()
	}, RequireRole(models.RoleAdmin))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestRequireRoleForbidden(t *testing.T) {
	r, called := newTestRouter(func(c *gin.Context) {
		c.Set("userRole", models.RoleCustomer)
		c.Next()
	}, RequireRole(models.RoleAdmin))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusForbidden)

	code, _ := decodeError(t, w)
	if code != "forbidden" {
		t.Errorf("expected forbidden, got %q", code)
	}
	if *called {
		t.Error("handler should not run when role is insufficient")
	}
}
