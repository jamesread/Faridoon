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

func TestNormalizeEvents(t *testing.T) {
	got, err := NormalizeEvents([]string{"approval.requested", " approval.requested "})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "approval.requested" {
		t.Fatalf("got=%v", got)
	}
}

func TestNormalizeEventsRejectsUnknown(t *testing.T) {
	if _, err := NormalizeEvents([]string{"nope"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeEventsEmpty(t *testing.T) {
	empty, err := NormalizeEvents(nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty=%v err=%v", empty, err)
	}
}

func TestSignature(t *testing.T) {
	sig := Signature(`{"a":1}`, "secret")
	if len(sig) != 64 {
		t.Fatalf("sig=%s", sig)
	}
	assertHexDigest(t, sig)
}

func assertHexDigest(t *testing.T, sig string) {
	t.Helper()
	for _, c := range sig {
		if !isHexDigit(c) {
			t.Fatalf("non-hex sig=%s", sig)
		}
	}
}

func isHexDigit(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
}
