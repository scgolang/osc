package osc

import (
	"log"
	"net"
	"sync"
	"testing"
)

func TestMulticastSend(t *testing.T) {
	const group = "224.10.10.1:9999"

	// BUG(briansorahan): How to reliably test multicast everywhere?
	ifi, err := net.InterfaceByIndex(2)
	if err != nil {
		t.Fatal(err)
	}
	gaddr, err := net.ResolveUDPAddr("udp", group)
	if err != nil {
		t.Fatal(err)
	}
	server1, err := ListenMulticastUDP("udp", ifi, gaddr)
	if err != nil {
		t.Fatal(err)
	}
	defer server1.Close()

	server2, err := ListenMulticastUDP("udp", ifi, gaddr)
	if err != nil {
		t.Fatal(err)
	}
	defer server2.Close()

	errChan, wg := make(chan error), &sync.WaitGroup{}

	wg.Add(2)

	go func() {
		errChan <- server1.Serve(map[string]Method{
			"/mcast/method": func(msg *Message) error {
				log.Println("server1")
				wg.Done()
				return nil
			},
		})
	}()

	go func() {
		errChan <- server2.Serve(map[string]Method{
			"/mcast/method": func(msg *Message) error {
				log.Println("server2")
				wg.Done()
				return nil
			},
		})
	}()

	client, err := DialUDP("udp", nil, gaddr)
	if err != nil {
		t.Fatal(err)
	}
	
	msg, err := NewMessage("/mcast/method")
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Send(msg); err != nil {
		t.Fatal(err)
	}

	select {
	default:
		wg.Wait()
	case err := <-errChan:
		if err != nil {
			t.Fatal(err)
		}
	}
}
