package osc

import "testing"

func TestAddMsgHandler(t *testing.T) {
	server, err := NewServer("localhost:6677")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	if err := server.AddMsgHandler("/address/test", func(msg *Message) {}); err != nil {
		t.Error("Expected that OSC address '/address/test' is valid")
	}
}

func TestAddMsgHandlerWithInvalidAddress(t *testing.T) {
	server, err := NewServer("localhost:6677")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	if err := server.AddMsgHandler("/address*/test", func(msg *Message) {}); err == nil {
		t.Error("Expected error with '/address*/test'")
	}
}

func TestServerMessageDispatching(t *testing.T) {
	server, err := NewServer("localhost:6677")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	if err := server.AddMsgHandler("/address/test", func(msg *Message) {
		if len(msg.arguments) != 1 {
			t.Error("Argument length should be 1 and is: " + string(len(msg.arguments)))
		}

		if msg.arguments[0].(int32) != 1122 {
			t.Error("Argument should be 1122 and is: " + string(msg.arguments[0].(int32)))
		}
	}); err != nil {
		t.Error("Error adding message handler")
	}

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
		msg.Append(int32(1122))
		client.Send(msg)
	}
}
