package osc

import (
	"errors"
	"testing"
	"time"
)

var matchedError = errors.New("matched")

// Test a successful method invocation.
func TestRegexpMatchingDispatchOK(t *testing.T) {
	d := RegexpMatching{
		"/bar": Method(func(msg Message) error {
			return matchedError
		}),
		"/foo.*": Method(func(msg Message) error {
			return matchedError
		}),
	}
	later := time.Now().Add(20 * time.Millisecond)
	b := Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Message{Address: "/bar"},
		},
	}

	if err := d.Dispatch(b, false); !errors.Is(err, matchedError) {
		t.Fatalf("expected match, got: %v", err)
	}

	b = Bundle{
		Packets: []Packet{
			Message{Address: "/foo"},
		},
	}
	if err := d.Dispatch(b, false); !errors.Is(err, matchedError) {
		t.Fatalf("expected match, got: %v", err)
	}
	b = Bundle{
		Packets: []Packet{
			Message{Address: "/foobar"},
		},
	}
	if err := d.Dispatch(b, false); !errors.Is(err, matchedError) {
		t.Fatalf("expected match, got: %v", err)
	}

}

// Test a method that returns an error.
func TestRegexpMatchingDispatchError(t *testing.T) {
	d := RegexpMatching{
		"/foo": Method(func(msg Message) error {
			return matchedError
		}),
		"^/baz$": Method(func(msg Message) error {
			return matchedError
		}),
	}
	later := time.Now().Add(20 * time.Millisecond)
	b := Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Message{Address: "/foo"},
		},
	}
	if err := d.Dispatch(b, false); !errors.Is(err, matchedError) {
		t.Fatalf("expected match, got: %v", err)
	}
	b = Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Message{Address: "all matches /foo if it contains /foo"},
		},
	}
	if err := d.Dispatch(b, false); !errors.Is(err, matchedError) {
		t.Fatalf("expected match, got: %v", err)
	}

	b = Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Message{Address: "/baz"},
		},
	}
	if err := d.Dispatch(b, false); !errors.Is(err, matchedError) {
		t.Fatalf("expected match, got: %v", err)
	}

	later = time.Now().Add(20 * time.Millisecond)
	b = Bundle{
		Timetag: FromTime(later),
		Packets: []Packet{
			Message{Address: "/bazbaz"},
		},
	}
	if err := d.Dispatch(b, false); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRegexpMatchingDispatchNestedBundle(t *testing.T) {
	c := make(chan struct{})
	d := RegexpMatching{
		"/foo": Method(func(msg Message) error {
			close(c)
			return nil
		}),
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
	if err := d.Dispatch(b, false); err != nil {
		t.Fatal(err)
	}
	<-c
}

func TestRegexpMatchingMiss(t *testing.T) {
	d := RegexpMatching{
		"/foo": Method(func(msg Message) error {
			return nil
		}),
	}
	b := Bundle{
		Timetag: FromTime(time.Now()),
	}
	if err := d.Dispatch(b, false); err != nil {
		t.Fatal(err)
	}
}

func TestRegexpMatchingInvoke(t *testing.T) {
	d := RegexpMatching{
		"/foo": Method(func(msg Message) error {
			return matchedError
		}),
		"/bar": Method(func(msg Message) error {
			return nil
		}),
	}
	msg := Message{Address: "/foo"}
	if err := d.Invoke(msg, false); !errors.Is(err, matchedError) {
		t.Fatal("expected matched error, got: %w", err)
	}
	badMsg := Message{Address: "/["}
	if err := d.Invoke(badMsg, false); err != nil {
		t.Fatal("expected nil, got: %w", err)
	}
	if err := d.Invoke(Message{Address: "/bar"}, false); err != nil {
		t.Fatal("expected nil, got: %w", err)
	}
	if err := d.Invoke(Message{Address: "/baz"}, false); err != nil {
		t.Fatal("expected nil, got: %w", err)
	}

	if err := d.invoke(badPacket{}, false); err == nil {
		t.Fatal("expected no error, got: %w", err)
	}
}
