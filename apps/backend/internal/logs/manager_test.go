package logs

import "testing"

func TestMaskSecretsMasksUUIDAndProxyURL(t *testing.T) {
	input := "connect vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?security=reality uuid 550e8400-e29b-41d4-a716-446655440000"
	got := MaskSecrets(input)
	if got == input {
		t.Fatal("expected masked output")
	}
	if got == "" || got == "connect " {
		t.Fatalf("unexpected masked output: %q", got)
	}
	if contains(got, "446655440000") {
		t.Fatalf("uuid tail was not masked: %s", got)
	}
}

func contains(value string, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
