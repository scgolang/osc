package osc

import (
	"testing"
)

func TestWriteRead(t *testing.T) {
}

func TestVerifyParts(t *testing.T) {
	// Pairs that should match
	for _, pair := range [][2]string{
		{"/osc/address", "/osc/address"},
		{"/path/to/method", "/path/to/meth?d"},
	} {
		if !verifyParts(pair[0], pair[1]) {
			t.Fatalf("Expected %s to have the same parts as %s", pair[0], pair[1])
		}
	}
}
