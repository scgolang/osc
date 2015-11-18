package osc

import "testing"

func TestTypeTagsString(t *testing.T) {
	msg := NewMessage("/some/address")
	if err := msg.WriteInt32(100); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(true); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(false); err != nil {
		t.Fatal(err)
	}
	if expected, got := ",iTF", msg.TypeTags(); expected != got {
		t.Fatalf("Expected %s got %s", expected, got)
	}
}
