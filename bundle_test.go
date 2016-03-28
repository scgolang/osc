package osc

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestBundle(t *testing.T) {
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

	bundle := NewBundle(time.Now(), msg)

	// Send a message.
	if err := client.Send(bundle); err != nil {
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
