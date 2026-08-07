package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocart/internal/dto"
	"gocart/internal/security"

	"github.com/gin-gonic/gin"
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

func webhookBody(t *testing.T, reference, status string) string {
	t.Helper()
	body, err := json.Marshal(dto.PaymentWebhookEvent{Reference: reference, Status: status})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	return string(body)
}

func doRequestWithSignature(t *testing.T, r *gin.Engine, body, sig string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/webhooks/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", sig)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPaymentWebhookSuccess(t *testing.T) {
	var gotBody []byte
	var gotSig string
	svc := &stubPaymentService{
		processWebhookFn: func(body []byte, signature string) (*dto.PaymentResponse, error) {
			gotBody, gotSig = body, signature
			return &dto.PaymentResponse{Reference: "PAY_1", Status: "succeeded"}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/webhooks/payments", NewPaymentHandler(svc).PaymentWebhook)

	body := webhookBody(t, "PAY_1", "succeeded")
	sig := security.SignWebhookBody([]byte("test-secret"), []byte(body))

	w := doRequestWithSignature(t, r, body, sig)
	assertStatus(t, w, http.StatusOK)

	if string(gotBody) != body || gotSig != sig {
		t.Errorf("unexpected service args: body=%q sig=%q", gotBody, gotSig)
	}
}

func TestPaymentWebhookMissingSignature(t *testing.T) {
	svc := &stubPaymentService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/webhooks/payments", NewPaymentHandler(svc).PaymentWebhook)

	w := doRequest(t, r, http.MethodPost, "/webhooks/payments", webhookBody(t, "PAY_1", "succeeded"))
	assertStatus(t, w, http.StatusUnauthorized)

	code, _ := decodeError(t, w)
	if code != "invalid_webhook_signature" {
		t.Errorf("expected invalid_webhook_signature, got %q", code)
	}
}

func TestPaymentWebhookServiceError(t *testing.T) {
	svc := &stubPaymentService{
		processWebhookFn: func(body []byte, signature string) (*dto.PaymentResponse, error) {
			return nil, appErr(http.StatusUnauthorized, "invalid_webhook_signature", "bad signature")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/webhooks/payments", NewPaymentHandler(svc).PaymentWebhook)

	body := webhookBody(t, "PAY_1", "succeeded")
	sig := security.SignWebhookBody([]byte("test-secret"), []byte(body))

	w := doRequestWithSignature(t, r, body, sig)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestSimulatePaymentSuccess(t *testing.T) {
	var gotRef, gotStatus string
	svc := &stubPaymentService{
		simulateWebhookFn: func(reference string, status string) (*dto.PaymentResponse, error) {
			gotRef, gotStatus = reference, status
			return &dto.PaymentResponse{Reference: reference, Status: status}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/dev/simulate-payment", NewPaymentHandler(svc).SimulatePayment)

	w := doRequest(t, r, http.MethodPost, "/dev/simulate-payment", `{"reference":"PAY_1","status":"succeeded"}`)
	assertStatus(t, w, http.StatusOK)

	if gotRef != "PAY_1" || gotStatus != "succeeded" {
		t.Errorf("unexpected service args: ref=%q status=%q", gotRef, gotStatus)
	}
}

func TestSimulatePaymentInvalidBody(t *testing.T) {
	svc := &stubPaymentService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/dev/simulate-payment", NewPaymentHandler(svc).SimulatePayment)

	w := doRequest(t, r, http.MethodPost, "/dev/simulate-payment", `{not json`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSimulatePaymentServiceError(t *testing.T) {
	svc := &stubPaymentService{
		simulateWebhookFn: func(reference string, status string) (*dto.PaymentResponse, error) {
			return nil, appErr(http.StatusNotFound, "payment_not_found", "not found")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/dev/simulate-payment", NewPaymentHandler(svc).SimulatePayment)

	w := doRequest(t, r, http.MethodPost, "/dev/simulate-payment", `{"reference":"PAY_1","status":"succeeded"}`)
	assertStatus(t, w, http.StatusNotFound)
}
