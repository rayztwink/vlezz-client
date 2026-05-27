package subscription

import "testing"

func TestParseVLESSReality(t *testing.T) {
	link := "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?type=tcp&security=reality&sni=example.com&fp=chrome&pbk=public-key&sid=abcd&flow=xtls-rprx-vision#Example"
	node, err := ParseVLESS(link)
	if err != nil {
		t.Fatalf("ParseVLESS returned error: %v", err)
	}
	if node.Name != "Example" {
		t.Fatalf("expected name Example, got %s", node.Name)
	}
	if node.Security != "reality" {
		t.Fatalf("expected reality security, got %s", node.Security)
	}
	if node.Transport != "tcp" {
		t.Fatalf("expected tcp transport, got %s", node.Transport)
	}
	if node.Params["flow"] != "xtls-rprx-vision" {
		t.Fatalf("expected flow query param")
	}
}

func TestParseVLESSRejectsInvalidUUID(t *testing.T) {
	_, err := ParseVLESS("vless://not-a-uuid@example.com:443?type=tcp#Bad")
	if err == nil {
		t.Fatal("expected invalid uuid error")
	}
}
