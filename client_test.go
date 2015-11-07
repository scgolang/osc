package osc

import (
	"log"
	"os"
	"testing"
)

func TestClientSetLocalAddr(t *testing.T) {
	client := NewClient("localhost:8967")
	if err := client.SetLocalAddr("localhost:41789"); err != nil {
		t.Error(err.Error())
	}

	if expected, got := "127.0.0.1:41789", client.laddr.String(); expected != got {
		t.Errorf("Expected laddr to be %s but got %s", expected, got)
	}
}

func ExampleClient() {
	const addr = "127.0.0.1:8765"

	server, err := NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	errChan := make(chan error)

	if err := server.AddMsgHandler("/osc/address", func(msg *Message) {
		errChan <- msg.Write(os.Stdout)
	}); err != nil {
		log.Fatal(err)
	}

	go func() {
		errChan <- server.Listen()
	}()

	_ = <-server.Listening

	var (
		client = NewClient(addr)
		msg    = NewMessage("/osc/address")
	)
	msg.Append(int32(111))
	msg.Append(true)
	msg.Append("hello")
	client.Send(msg)

	if err := <-errChan; err != nil {
		log.Fatal(err)
	}
	// Output:
	// /osc/address ,iTs 111 true hello
}
