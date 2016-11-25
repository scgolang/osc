package osc

import (
	"errors"
	"testing"
)

func TestDispatcher(t *testing.T) {
	d := Dispatcher{
		"/foo": func(msg Message) error {
			return errors.New("foo error")
		},
		"/bar": func(msg Message) error {
			return nil
		},
	}
	msg := Message{Address: "/foo"}
	if err := d.Dispatch(msg); err == nil {
		t.Fatal("expected error, got nil")
	}
	badMsg := Message{Address: "/["}
	if err := d.Dispatch(badMsg); err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := d.Dispatch(Message{Address: "/bar"}); err != nil {
		t.Fatal(err)
	}
	if err := d.Dispatch(Message{Address: "/baz"}); err != nil {
		t.Fatal(err)
	}
}
