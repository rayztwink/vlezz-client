package subscription

import (
	"encoding/base64"
	"testing"
)

func TestDecodeMaybeBase64SupportsRawStdEncoding(t *testing.T) {
	payload := "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?type=tcp#Example"
	encoded := base64.RawStdEncoding.EncodeToString([]byte(payload))
	decoded, err := decodeMaybeBase64(encoded)
	if err != nil {
		t.Fatalf("decodeMaybeBase64 returned error: %v", err)
	}
	if string(decoded) != payload {
		t.Fatalf("unexpected decoded payload: %s", decoded)
	}
}
