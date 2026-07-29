package webhook

import "testing"

func TestNormalizeEvent(t *testing.T) {
	e, err := NormalizeEvent("approval.requested")
	if err != nil || e != "approval.requested" {
		t.Fatalf("%v %v", e, err)
	}
	if _, err := NormalizeEvent("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeURL(t *testing.T) {
	if _, err := NormalizeURL("https://example.com/hook"); err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeURL("ftp://x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSignature(t *testing.T) {
	sig := Signature(`{"a":1}`, "secret")
	if len(sig) != 64 {
		t.Fatalf("sig=%s", sig)
	}
}
