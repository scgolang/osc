package osc

import (
	"bytes"
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
	var (
		doneChan = make(chan *Message)
		errChan  = make(chan error, 1)
	)

	dispatcher := map[string]Method{
		"/osc/address": func(msg *Message) error {
			doneChan <- msg
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

	go func() {
		errChan <- server.Serve(dispatcher) // Best effort.
	}()

	serverAddr := server.LocalAddr()
	raddr, err := net.ResolveUDPAddr(serverAddr.Network(), serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}

	client, err := DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := NewMessage("/osc/address")
	if err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteInt32(111); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(true); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteString("hello"); err != nil {
		t.Fatal(err)
	}

	// Send a message.
	if err := client.Send(msg); err != nil {
		t.Fatal(err)
	}

	select {
	default:
	case err := <-errChan:
		if err != nil {
			t.Fatal(err)
		}
	}

	recvMsg := <-doneChan

	recvData, err := recvMsg.Contents()
	if err != nil {
		t.Fatal(err)
	}

	data, err := msg.Contents()
	if err != nil {
		t.Fatal(err)
	}

	if 0 != bytes.Compare(data, recvData[0:len(data)]) {
		t.Fatalf("Expected %s got %s", data, recvData)
	}
}
