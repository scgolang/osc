package osc

import (
	"net"
	"testing"
	"time"
)

type TestClientUDP struct {
	Conn *UDPConn
}

func newTestClientUDP(t *testing.T, addr net.Addr) *TestClientUDP {
	raddr, err := net.ResolveUDPAddr(addr.Network(), addr.String())
	if err != nil {
		t.Fatal(err)
	}
	client, err := DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatal(err)
	}
	return &TestClientUDP{Conn: client}
}

type TestServerUDP struct {
	MsgChan chan *Message
	ErrChan chan error
	Conn    *UDPConn
}

func (ts *TestServerUDP) Close() error {
	// TODO: handle Serve errors after closing
	// close(ts.MsgChan)
	// close(ts.ErrChan)
	return ts.Conn.Close()
}

func newTestServerUDP(t *testing.T) *TestServerUDP {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server, err := ListenUDP("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	ts := &TestServerUDP{
		MsgChan: make(chan *Message),
		ErrChan: make(chan error),
		Conn:    server,
	}
	return ts
}

func TestBundle(t *testing.T) {
	// TODO: fix
	t.SkipNow()

	ts := newTestServerUDP(t)
	dispatcher := map[string]Method{
		"/osc/address": func(msg *Message) error {
			ts.MsgChan <- msg
			return nil
		},
	}
	defer func() { _ = ts.Close() }() // Best effort.

	go func() {
		ts.ErrChan <- ts.Conn.Serve(dispatcher) // Best effort.
	}()

	client := newTestClientUDP(t, ts.Conn.LocalAddr())

	msg, err := NewMessage("/osc/address")
	if err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteInt32(0, 111); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(1, true); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteString(2, "hello"); err != nil {
		t.Fatal(err)
	}

	bundle := NewBundle(time.Now(), msg)

	// Send a message.
	if _, err := client.Conn.Send(bundle); err != nil {
		t.Fatal(err)
	}

	select {
	default:
	case err := <-ts.ErrChan:
		if err != nil {
			t.Fatal(err)
		}
	}

	recvMsg := <-ts.MsgChan

	if err := msg.Compare(recvMsg); err != nil {
		t.Fatal(err)
	}
}
