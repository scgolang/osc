package osc

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestClientSend(t *testing.T) {
	var (
		doneChan = make(chan *Message)
		errChan  = make(chan error, 1)
	)

	handlers := map[string]Method{
		"/osc/address": func(msg *Message) {
			doneChan <- msg
		},
	}

	server, err := NewServer("127.0.0.1:0", handlers)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	go func() {
		errChan <- server.Listen()
	}()

	_ = <-server.Listening

	client, err := NewClient(server.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}

	msg := NewMessage("/osc/address")
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
	data, err := msg.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Send(data); err != nil {
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

	recvData, err := recvMsg.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if 0 != bytes.Compare(data, recvData[0:len(data)]) {
		t.Fatalf("Expected %s got %s", data, recvData)
	}
}

func ExampleClient() {
	var (
		doneChan = make(chan struct{})
		errChan  = make(chan error, 1)
	)

	handlers := map[string]Method{
		"/osc/address": func(msg *Message) {
			errChan <- msg.Print(os.Stdout)
			doneChan <- struct{}{}
		},
	}

	server, err := NewServer("127.0.0.1:0", handlers)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = server.Close() }() // Best effort.

	go func() {
		errChan <- server.Listen()
	}()

	_ = <-server.Listening

	client, err := NewClient(server.LocalAddr())
	if err != nil {
		log.Fatal(err)
	}

	msg := NewMessage("/osc/address")
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
	data, err := msg.Bytes()
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Send(data); err != nil {
		log.Fatal(err)
	}

	select {
	default:
	case err := <-errChan:
		if err != nil {
			log.Fatal(err)
		}
	}

	_ = <-doneChan
	// Output:
	// /osc/address,iTs 111 true hello
}
