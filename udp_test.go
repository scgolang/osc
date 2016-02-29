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

func TestSend(t *testing.T) {
	ts := newTestServerUDP(t)

	dispatcher := map[string]Method{
		"/osc/address": func(msg *Message) error {
			ts.MsgChan <- msg
			return nil
		},
	}

	defer func() { _ = ts.Close() }() // Best effort.

	go func() {
		ts.ErrChan <- ts.Conn.Serve(dispatcher) // Best effort.
	}()

	client := newTestClientUDP(t, ts.Conn.LocalAddr())

	msg, err := NewMessage("/osc/address")
	if err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteInt32(0, 111); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(1, true); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteString(2, "hello"); err != nil {
		t.Fatal(err)
	}

	// Send a message.
	if _, err := client.Conn.Send(msg); err != nil {
		t.Fatal(err)
	}

	select {
	default:
	case err := <-ts.ErrChan:
		if err != nil {
			t.Fatal(err)
		}
	}

	recvMsg := <-ts.MsgChan

	if err := msg.Compare(recvMsg); err != nil {
		t.Fatal(err)
	}
}
