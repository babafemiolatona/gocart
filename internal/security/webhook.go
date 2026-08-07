package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// SignWebhookBody computes the HMAC-SHA256 signature of a raw webhook body
// using the shared secret. Used by the mock provider (and tests) to produce
// signatures the server will accept.
func SignWebhookBody(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature reports whether the provided hex signature matches the
// HMAC-SHA256 of the raw body. Comparison is constant-time via hmac.Equal.
func VerifyWebhookSignature(secret, body []byte, provided string) bool {
	expected := SignWebhookBody(secret, body)
	return hmac.Equal([]byte(expected), []byte(provided))
}
