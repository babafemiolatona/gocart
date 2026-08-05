package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
)

func TestProcessPaymentSuccess(t *testing.T) {
	var gotUser uint
	var gotRef string
	svc := &stubPaymentService{
		processFn: func(userID uint, reference string) (*dto.PaymentResponse, error) {
			gotUser, gotRef = userID, reference
			return &dto.PaymentResponse{ID: 1, Reference: reference, Status: "succeeded"}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/payments/:reference/process", NewPaymentHandler(svc).ProcessPayment)

	w := doRequest(t, r, http.MethodPost, "/payments/PAY_1/process", "")
	assertStatus(t, w, http.StatusOK)

	var resp dto.PaymentResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.Reference != "PAY_1" || resp.Status != "succeeded" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if gotUser != 7 || gotRef != "PAY_1" {
		t.Errorf("unexpected service args: user=%d ref=%q", gotUser, gotRef)
	}
}

func TestProcessPaymentMissingReference(t *testing.T) {
	svc := &stubPaymentService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/payments/:reference/process", NewPaymentHandler(svc).ProcessPayment)
	registerHandler(r, http.MethodPost, "/payments/process", NewPaymentHandler(svc).ProcessPayment)

	w := doRequest(t, r, http.MethodPost, "/payments/process", "")
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "invalid_payment_reference" {
		t.Errorf("expected invalid_payment_reference, got %q", code)
	}
}

func TestProcessPaymentServiceError(t *testing.T) {
	svc := &stubPaymentService{
		processFn: func(userID uint, reference string) (*dto.PaymentResponse, error) {
			return nil, appErr(http.StatusConflict, "payment_conflict", "conflict")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/payments/:reference/process", NewPaymentHandler(svc).ProcessPayment)

	w := doRequest(t, r, http.MethodPost, "/payments/PAY_1/process", "")
	assertStatus(t, w, http.StatusConflict)
}

func TestGetPaymentSuccess(t *testing.T) {
	svc := &stubPaymentService{
		getFn: func(userID uint, reference string) (*dto.PaymentResponse, error) {
			return &dto.PaymentResponse{Reference: reference, Status: "succeeded"}, nil
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/payments/:reference", NewPaymentHandler(svc).GetPayment)

	w := doRequest(t, r, http.MethodGet, "/payments/PAY_1", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetPaymentMissingReference(t *testing.T) {
	svc := &stubPaymentService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/payments/:reference", NewPaymentHandler(svc).GetPayment)
	registerHandler(r, http.MethodGet, "/payments/none", NewPaymentHandler(svc).GetPayment)

	w := doRequest(t, r, http.MethodGet, "/payments/none", "")
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "invalid_payment_reference" {
		t.Errorf("expected invalid_payment_reference, got %q", code)
	}
}

func TestProcessPaymentUnauthorized(t *testing.T) {
	svc := &stubPaymentService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/payments/:reference/process", NewPaymentHandler(svc).ProcessPayment)

	w := doRequest(t, r, http.MethodPost, "/payments/PAY_1/process", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetPaymentUnauthorized(t *testing.T) {
	svc := &stubPaymentService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/payments/:reference", NewPaymentHandler(svc).GetPayment)

	w := doRequest(t, r, http.MethodGet, "/payments/PAY_1", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetPaymentServiceError(t *testing.T) {
	svc := &stubPaymentService{
		getFn: func(userID uint, reference string) (*dto.PaymentResponse, error) {
			return nil, appErr(http.StatusNotFound, "payment_not_found", "not found")
		},
	}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/payments/:reference", NewPaymentHandler(svc).GetPayment)

	w := doRequest(t, r, http.MethodGet, "/payments/PAY_1", "")
	assertStatus(t, w, http.StatusNotFound)
}
