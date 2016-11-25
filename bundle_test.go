package osc

import (
	"bytes"
	"testing"
)

func TestBundle(t *testing.T) {
	b := Bundle{}
	if expected, got := []byte{}, b.Bytes(); !bytes.Equal(expected, got) {
		t.Fatalf("expected %x, got %x", expected, got)
	}
}
