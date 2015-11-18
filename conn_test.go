package osc

import (
	"net"
	"testing"
)

func TestInvalidAddress(t *testing.T) {
	dispatcher := map[string]Method{
		"/address*/test": func(msg *Message) error {
			return nil
		},
	}
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	if err := server.Serve(dispatcher); err != ErrInvalidAddress {
		t.Fatal("expected invalid address error")
	}
	if server != nil {
		_ = server.Close()
	}
}

func TestMessageDispatching(t *testing.T) {
	// dispatcher := map[string]Method{
	// 	"/address/test": func(msg *Message) error {
	// 		val, err := msg.ReadInt32()
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if expected, got := int32(1122), val; expected != got {
	// 			return fmt.Errorf("Expected %d got %d", expected, got)
	// 		}
	// 		return nil
	// 	},
	// }

	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.
}
