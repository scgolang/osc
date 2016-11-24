package osc

import (
	"testing"
)

func TestValidateAddress(t *testing.T) {
	if err := ValidateAddress("/foo"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAddress("/foo@^#&*$^*%)()#($*@"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
