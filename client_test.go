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
	errChan := make(chan error)

	server, err := NewServer("127.0.0.1:0", map[string]HandlerFunc{
		"/osc/address": func(msg *Message) {
			errChan <- msg.Write(os.Stdout)
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	go func() {
		errChan <- server.Listen()
	}()

	_ = <-server.Listening

	var (
		client = NewClient(server.LocalAddr().String())
		msg    = NewMessage("/osc/address")
	)
	if err := msg.WriteInt32(111); err != nil {
		log.Fatal(err)
	}
	if err := msg.WriteBool(true); err != nil {
		log.Fatal(err)
	}
	if err := msg.WriteString("hello"); err != nil {
		log.Fatal(err)
	}

	// Send a message.
	if err := client.Send(msg); err != nil {
		log.Fatal(err)
	}

	select {
	default:
	case err := <-errChan:
		if err != nil {
			log.Fatal(err)
		}
	}
	// Output:
	// /osc/address ,iTs 111 true hello
}
