package osc

import (
	"bytes"
	"io/ioutil"
	"net"
	"testing"

	"github.com/pkg/errors"
)

func TestMessageEqual(t *testing.T) {
	for _, testcase := range []struct {
		M1       Message
		M2       Packet
		Expected bool
	}{
		{
			M1:       Message{Address: "/foo"},
			M2:       Message{Address: "/foo"},
			Expected: true,
		},
		{
			M1:       Message{Address: "/foo"},
			M2:       Message{Address: "/bar"},
			Expected: false,
		},
		{
			M1:       Message{Address: "/foo"},
			M2:       Message{Address: "/foo", Arguments: []Argument{Int(32)}},
			Expected: false,
		},
		{
			M1:       Message{Address: "/foo", Arguments: []Argument{Int(31)}},
			M2:       Message{Address: "/foo", Arguments: []Argument{Int(32)}},
			Expected: false,
		},
		{
			M1:       Message{Address: "/foo", Arguments: []Argument{Int(31)}},
			M2:       Bundle{},
			Expected: false,
		},
	} {
		m1, m2 := testcase.M1, testcase.M2
		if testcase.Expected {
			if !m1.Equal(m2) {
				t.Fatalf("expected %s to equal %s", m1, m2)
			}
		} else {
			if m1.Equal(m2) {
				t.Fatalf("expected %s to not equal %s", m1, m2)
			}
		}
	}
}

func TestVerifyParts(t *testing.T) {
	// Pairs that should match.
	for _, pair := range [][2]string{
		{"/osc/address", "/osc/address"},
		{"/path/to/method", "/path/to/meth?d"},
	} {
		if !VerifyParts(pair[0], pair[1]) {
			t.Fatalf("Expected %s to match %s", pair[0], pair[1])
		}
	}

	// Pairs that should not match.
	for _, pair := range [][2]string{
		{"/osc/address///", "//osc/a/b/"},
	} {
		if VerifyParts(pair[0], pair[1]) {
			t.Fatalf("Expected %s to not match %s", pair[0], pair[1])
		}
	}
}

func TestMatch(t *testing.T) {
	// matches where 1st string is pattern and second is address
	for _, pair := range [][2]string{
		{"/path/to/method", "/path/to/method"},
		{"/path/to/meth?d", "/path/to/method"},
		{"/path/to/*", "/path/to/method"},
		{"/path/to/method*", "/path/to/method"},
		{"/path/to/m[aei]thod", "/path/to/method"},
	} {
		msg := Message{Address: pair[0]}
		match, err := msg.Match(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Fatalf("Expected %s to match %s", pair[1], pair[0])
		}
	}

	// misses where 1st string is pattern and second is address
	for _, pair := range [][2]string{
		{"/path/to/destruction", "/path/to/method"},
		{"/path/to/me?thod", "/path/to/method"},
		{"/path/to?method", "/path/to/method"},
		{"/path/to*", "/path/to/method"},
		{"/path/to/[domet]", "/path/to/method"},
	} {
		msg := Message{Address: pair[0]}
		match, err := msg.Match(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if match {
			t.Fatalf("Expected %s to not match %s", pair[1], pair[0])
		}
	}

	msg := Message{Address: `/[`}
	if _, err := msg.Match(`/a`); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetRegex(t *testing.T) {
	if _, err := GetRegex(`[`); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMesssageBytes(t *testing.T) {
	for _, testcase := range []struct {
		Message  Message
		Expected []byte
	}{
		{
			Message: Message{
				Address:   "/foo",
				Arguments: []Argument{Int(1), Blob([]byte("bar"))},
			},
			Expected: bytes.Join(
				[][]byte{
					{'/', 'f', 'o', 'o', 0, 0, 0, 0},
					{TypetagPrefix, TypetagInt, TypetagBlob, 0},
					{0, 0, 0, 1},
					{0, 0, 0, 3, 'b', 'a', 'r', 0},
				},
				[]byte{},
			),
		},
	} {
		b := testcase.Message.Bytes()
		if expected, got := testcase.Expected, b; !bytes.Equal(expected, got) {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
}

type errWriter struct {
	erridx int
	curr   int
}

func (e *errWriter) Write(b []byte) (n int, err error) {
	e.curr++
	if e.curr == e.erridx {
		return 0, errors.New("oops")
	}
	return 0, nil
}

func TestMessageWriteTo(t *testing.T) {
	var (
		msg = Message{Address: "/foo", Arguments: []Argument{String("bar")}}
		e1  = &errWriter{erridx: 1}
		e2  = &errWriter{erridx: 2}
	)
	if _, err := msg.WriteTo(e1); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, err := msg.WriteTo(e2); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, err := msg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestParseMessage(t *testing.T) {
	type Input struct {
		data   []byte
		sender net.Addr
	}
	type Output struct {
		Message Message
		Err     error
	}
	for i, testcase := range []struct {
		Input    Input
		Expected Output
	}{
		{
			Input: Input{
				data: bytes.Join(
					[][]byte{
						{'/', 'f', 'o', 'o', 0, 0, 0, 0},
						{TypetagPrefix, TypetagInt, TypetagBlob, 0},
						{0, 0, 0, 1},
						{0, 0, 0, 3, 'b', 'a', 'r', 0},
					},
					[]byte{},
				),
			},
			Expected: Output{
				Message: Message{
					Address: "/foo",
					Arguments: []Argument{
						Int(1),
						Blob([]byte{'b', 'a', 'r', 0}),
					},
				},
			},
		},
		{
			Input: Input{
				data: bytes.Join(
					[][]byte{
						{'/', 'f', 'o', 'o', 0, 0, 0, 0},
						{TypetagPrefix, 'Q', 0, 0},
					},
					[]byte{},
				),
			},
			Expected: Output{Err: errors.New(`read argument 0: typetag "Q": invalid typetag`)},
		},
	} {
		msg, err := ParseMessage(testcase.Input.data, testcase.Input.sender)
		if testcase.Expected.Err == nil {
			if err != nil {
				t.Fatalf("(testcase %d) %s", i, err)
			}
			if expected, got := testcase.Expected.Message, msg; !expected.Equal(got) {
				t.Fatalf("(testcase %d) expected %s, got %s", i, expected, got)
			}
		} else {
		}
	}
}
