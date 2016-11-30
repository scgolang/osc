package osc

import (
	"net"
	"testing"
)

func TestUDPConn(t *testing.T) {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	lc, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	var c Conn = lc
	_ = c
}

func TestValidateAddress(t *testing.T) {
	if err := ValidateAddress("/foo"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAddress("/foo@^#&*$^*%)()#($*@"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
