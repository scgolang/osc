package osc

import (
	"bytes"
	"testing"
)

func TestToBytes(t *testing.T) {
	for _, testcase := range []struct {
		Input    string
		Expected []byte
	}{
		{
			Input:    "",
			Expected: []byte{},
		},
		{
			Input:    "a",
			Expected: []byte{'a', 0, 0, 0},
		},
		{
			Input:    "abc",
			Expected: []byte{'a', 'b', 'c', 0},
		},
	} {
		got := ToBytes(testcase.Input)
		if expected := testcase.Expected; !bytes.Equal(expected, got) {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
}

func TestPad(t *testing.T) {
	for _, testcase := range []struct {
		Input    []byte
		Expected []byte
	}{
		{
			Input:    []byte{},
			Expected: []byte{},
		},
		{
			Input:    []byte("a"),
			Expected: []byte{'a', 0, 0, 0},
		},
		{
			Input:    []byte("abc"),
			Expected: []byte{'a', 'b', 'c', 0},
		},
	} {
		got := Pad(testcase.Input)
		if expected := testcase.Expected; !bytes.Equal(expected, got) {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
}

func TestReadString(t *testing.T) {
	type Output struct {
		String string
		Length int64
	}
	for _, testcase := range []struct {
		Input    []byte
		Expected Output
	}{
		{
			Input: []byte{},
			Expected: Output{
				String: "",
				Length: 0,
			},
		},
		{
			Input: []byte{'a'},
			Expected: Output{
				String: "a",
				Length: 4,
			},
		},
		{
			Input: []byte("abc"),
			Expected: Output{
				String: "abc",
				Length: 4,
			},
		},
		{
			Input: []byte("/foo\x00\x00\x00\x00,ib\x00\x00\x00\x00\x01\x00\x00\x00\x03bar\x00"),
			Expected: Output{
				String: "/foo",
				Length: 8,
			},
		},
		{
			Input: []byte{',', 'Q'},
			Expected: Output{
				String: ",Q",
				Length: 4,
			},
		},
	} {
		s, bw := ReadString(testcase.Input)
		if expected, got := testcase.Expected.Length, bw; expected != got {
			t.Fatalf("expected %d, got %d", expected, got)
		}
		if expected, got := testcase.Expected.String, s; expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestReadBlob(t *testing.T) {
	type Input struct {
		Length int32
		Data   []byte
	}
	type Output struct {
		Data   []byte
		Length int64
	}
	for _, testcase := range []struct {
		Input    Input
		Expected Output
	}{
		{
			Input: Input{
				Length: 0,
				Data:   []byte{},
			},
			Expected: Output{
				Data:   []byte{},
				Length: 0,
			},
		},
		{
			Input: Input{
				Length: 20,
				Data:   []byte{'a'},
			},
			Expected: Output{
				Data:   []byte{'a', 0, 0, 0},
				Length: 4,
			},
		},
	} {
		b, bl := ReadBlob(testcase.Input.Length, testcase.Input.Data)
		if expected, got := testcase.Expected.Length, bl; expected != got {
			t.Fatalf("expected %d, got %d", expected, got)
		}
		if expected, got := testcase.Expected.Data, b; !bytes.Equal(expected, got) {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
}
