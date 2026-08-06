package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"

	apperrors "gocart/internal/errors"
)

func TestErrorHandlerPassThroughWhenNoErrors(t *testing.T) {
	// The ErrorHandler is already applied in newTestRouter, so a clean request
	// passes straight through to the terminal handler.
	r, called := newTestRouter()
	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusOK)
	if !*called {
		t.Error("expected terminal handler to run")
	}
}

func TestErrorHandlerRendersAppError(t *testing.T) {
	r, _ := newTestRouter(func(c *gin.Context) {
		c.Error(apperrors.New(
			http.StatusNotFound,
			"order_not_found",
			"order not found",
			nil,
		))
		c.Abort()
	})

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusNotFound)

	code, message := decodeError(t, w)
	if code != "order_not_found" || message != "order not found" {
		t.Errorf("unexpected error body: %q %q", code, message)
	}
}

func TestErrorHandlerRendersGenericForPlainError(t *testing.T) {
	r, _ := newTestRouter(func(c *gin.Context) {
		c.Error(errBoom)
		c.Abort()
	})

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusInternalServerError)

	code, message := decodeError(t, w)
	if code != "internal_server_error" || message != "internal server error" {
		t.Errorf("unexpected error body: %q %q", code, message)
	}
}
