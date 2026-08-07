package security

import (
	"strings"
	"testing"
)

func TestVerifyWebhookSignature(t *testing.T) {
	secret := []byte("shared-secret")
	body := []byte(`{"reference":"PAY_1","status":"succeeded"}`)
	sig := SignWebhookBody(secret, body)

	if !VerifyWebhookSignature(secret, body, sig) {
		t.Fatal("expected signature to verify")
	}
}

func TestVerifyWebhookSignatureWrongSecret(t *testing.T) {
	sig := SignWebhookBody([]byte("secret-a"), []byte("hello"))
	if VerifyWebhookSignature([]byte("secret-b"), []byte("hello"), sig) {
		t.Fatal("expected signature from another secret to be rejected")
	}
}

func TestVerifyWebhookSignatureTamperedBody(t *testing.T) {
	secret := []byte("shared-secret")
	body := []byte(`{"reference":"PAY_1","status":"succeeded"}`)
	sig := SignWebhookBody(secret, body)

	tampered := strings.Replace(string(body), "succeeded", "failed", 1)
	if VerifyWebhookSignature(secret, []byte(tampered), sig) {
		t.Fatal("expected tampered body to be rejected")
	}
}

func TestVerifyWebhookSignatureEmptyOrGarbage(t *testing.T) {
	secret := []byte("shared-secret")
	body := []byte("hello")

	if VerifyWebhookSignature(secret, body, "") {
		t.Fatal("expected empty signature to be rejected")
	}
	if VerifyWebhookSignature(secret, body, "not-hex") {
		t.Fatal("expected garbage signature to be rejected")
	}
}
