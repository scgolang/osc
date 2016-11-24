package osc

import (
	"errors"
	"testing"
)

func TestDispatchMessage(t *testing.T) {
	d := Dispatcher{
		"/foo": func(msg *Message) error {
			return errors.New("foo error")
		},
	}
	msg, err := NewMessage("/foo")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.DispatchMessage(msg); err == nil {
		t.Fatal("expected error, got nil")
	}
}
