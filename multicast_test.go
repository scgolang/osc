package osc

import (
	"net"
	"sync"
	"testing"
)

const group = "224.10.10.1:9999"

type TestMulticastServer struct {
	Conn *UDPConn
}

func newTestMulticastServer(t *testing.T, ifIndex int) (*TestMulticastServer, net.Addr) {
	// BUG(briansorahan): How to reliably test multicast everywhere?
	ifi, err := net.InterfaceByIndex(ifIndex)
	if err != nil {
		t.Fatal(err)
	}
	gaddr, err := net.ResolveUDPAddr("udp", group)
	if err != nil {
		t.Fatal(err)
	}
	server, err := ListenMulticastUDP("udp", ifi, gaddr)
	if err != nil {
		t.Fatal(err)
	}
	return &TestMulticastServer{Conn: server}, gaddr
}

func TestMulticastSend(t *testing.T) {
	var (
		errChan = make(chan error)
		wg      = &sync.WaitGroup{}
	)
	ts1, gaddr := newTestMulticastServer(t, 4)
	ts2, _ := newTestMulticastServer(t, 4)
	defer func() { _ = ts1.Conn.Close() }() // Best effort.
	defer func() { _ = ts2.Conn.Close() }() // Best effort.

	wg.Add(2)

	go func() {
		errChan <- ts1.Conn.Serve(map[string]Method{
			"/mcast/method": func(msg *Message) error {
				wg.Done()
				return nil
			},
		})
	}()

	go func() {
		errChan <- ts2.Conn.Serve(map[string]Method{
			"/mcast/method": func(msg *Message) error {
				wg.Done()
				return nil
			},
		})
	}()

	client := newTestClientUDP(t, gaddr)

	msg, err := NewMessage("/mcast/method")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Conn.Send(msg); err != nil {
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
