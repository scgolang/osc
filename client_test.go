package osc

import (
	"log"
	"testing"
)

func TestClientSetLocalAddr(t *testing.T) {
	client := NewClient("localhost", 8967)
	err := client.SetLocalAddr("localhost", 41789)
	if err != nil {
		t.Error(err.Error())
	}
	expectedAddr := "127.0.0.1:41789"
	if client.laddr.String() != expectedAddr {
		t.Errorf("Expected laddr to be %s but was %s", expectedAddr, client.laddr.String())
	}
}

func ExampleClient() {
	addr, port := "127.0.0.1", 8765
	server := NewServer(addr, port)

	done := make(chan error)

	server.AddMsgHandler("/osc/address", func(msg *Message) {
		PrintMessage(msg)
		done <-nil
	})

	go server.ListenAndDispatch()

	err := <-server.Listening

	if err != nil {
		log.Fatal(err)
	}

	client := NewClient(addr, port)
	msg := NewMessage("/osc/address")
	msg.Append(int32(111))
	msg.Append(true)
	msg.Append("hello")
	client.Send(msg)

	err = <-done

	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// /osc/address ,iTs 111 true hello
}
