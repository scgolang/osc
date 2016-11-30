package osc

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

// Test a successful method invocation.
func TestDispatcherDispatchOK(t *testing.T) {
	c := make(chan struct{})
	d := Dispatcher{
		"/bar": func(msg Message) error {
			close(c)
			return nil
		},
	}
	later := time.Now().Add(20 * time.Millisecond)
	b := Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Message{Address: "/bar"},
		},
	}
	if err := d.Dispatch(b); err != nil {
		t.Fatal(err)
	}
	<-c
}

// Test a method that returns an error.
func TestDispatcherDispatchError(t *testing.T) {
	d := Dispatcher{
		"/foo": func(msg Message) error {
			return errors.New("oops")
		},
	}
	later := time.Now().Add(20 * time.Millisecond)
	b := Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Message{Address: "/foo"},
		},
	}
	if err := d.Dispatch(b); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDispatcherDispatchNestedBundle(t *testing.T) {
	c := make(chan struct{})
	d := Dispatcher{
		"/foo": func(msg Message) error {
			close(c)
			return nil
		},
	}
	later := time.Now().Add(20 * time.Millisecond)
	b := Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Bundle{
				Timetag: FromTime(later.Add(20 * time.Millisecond)),
				Packets: []Packet{
					Message{Address: "/foo"},
				},
			},
		},
	}
	if err := d.Dispatch(b); err != nil {
		t.Fatal(err)
	}
	<-c
}

func TestDispatcherMiss(t *testing.T) {
	d := Dispatcher{
		"/foo": func(msg Message) error {
			return nil
		},
	}
	b := Bundle{
		Timetag: FromTime(time.Now()),
	}
	if err := d.Dispatch(b); err != nil {
		t.Fatal(err)
	}
}

func TestDispatcherInvoke(t *testing.T) {
	d := Dispatcher{
		"/foo": func(msg Message) error {
			return errors.New("foo error")
		},
		"/bar": func(msg Message) error {
			return nil
		},
	}
	msg := Message{Address: "/foo"}
	if err := d.Invoke(msg); err == nil {
		t.Fatal("expected error, got nil")
	}
	badMsg := Message{Address: "/["}
	if err := d.Invoke(badMsg); err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := d.Invoke(Message{Address: "/bar"}); err != nil {
		t.Fatal(err)
	}
	if err := d.Invoke(Message{Address: "/baz"}); err != nil {
		t.Fatal(err)
	}
	if err := d.invoke(badPacket{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}
