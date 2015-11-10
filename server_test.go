package osc

import "testing"

func TestInvalidAddress(t *testing.T) {
	handlers := map[string]HandlerFunc{
		"/address*/test": func(msg *Message) {},
	}
	server, err := NewServer("", handlers)
	if err != ErrInvalidAddress {
		t.Fatal("expected invalid address error")
	}
	if server != nil {
		_ = server.Close()
	}
}

func TestMessageDispatching(t *testing.T) {
	handlers := map[string]HandlerFunc{
		"/address/test": func(msg *Message) {
			val, err := msg.ReadInt32()
			if err != nil {
				t.Fatal(err)
			}
			if expected, got := int32(1122), val; expected != got {
				t.Fatalf("Expected %d got %d", expected, got)
			}
		},
	}

	server, err := NewServer("127.0.0.1:0", handlers)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	errChan := make(chan error)

	// Start the OSC server in a new go-routine
	go func() {
		errChan <- server.Listen()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatal(err)
		}
	case <-server.Listening:
		client := NewClient("localhost:6677")
		msg := NewMessage("/address/test")
		if err := msg.WriteInt32(1122); err != nil {
			t.Fatal(err)
		}
		client.Send(msg)
	}
}

func TestServerCloseBeforeListen(t *testing.T) {
	server, err := NewServer("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
}
