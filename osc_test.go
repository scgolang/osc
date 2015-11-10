package osc

import "testing"

func TestReadPaddedString(t *testing.T) {
	str, n := readPaddedString([]byte{'a', 'b', 'c', 'd', 0})
	if expected := "abcd"; str != expected {
		t.Fatalf("Expected %s got %s", expected, str)
	}
	if expected := 8; n != expected {
		t.Fatalf("Expected %d got %d", expected, n)
	}
}

func TestPaddedSize(t *testing.T) {
	if expected, got := 0, paddedSize(-1); expected != got {
		t.Fatalf("Expected %d got %d.", expected, got)
	}
	if expected, got := 0, paddedSize(0); expected != got {
		t.Fatalf("Expected %d got %d.", expected, got)
	}
	if expected, got := 4, paddedSize(1); expected != got {
		t.Fatalf("Expected %d got %d.", expected, got)
	}
	if expected, got := 8, paddedSize(6); expected != got {
		t.Fatalf("Expected %d got %d.", expected, got)
	}
	if expected, got := 12, paddedSize(11); expected != got {
		t.Fatalf("Expected %d got %d.", expected, got)
	}
}
