package osc

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"testing"
)

func TestInt(t *testing.T) {
	arg := Int(0)
	i, err := arg.ReadInt32()
	if err != nil {
		t.Fatal(err)
	}
	if other := Int(0); !arg.Equal(other) {
		t.Fatal("expected %s to equal %s", arg, other)
	}
	if other := Int(2); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
	if other := String("foo"); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
	if expected, got := int32(0), i; expected != got {
		t.Fatalf("expected %d, got %d", expected, got)
	}
	if _, err := arg.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if expected, got := "Int(0)", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
	if expected, got := TypetagInt, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestFloat(t *testing.T) {
	arg := Float(0)
	f, err := arg.ReadFloat32()
	if err != nil {
		t.Fatal(err)
	}
	if other := Float(0); !arg.Equal(other) {
		t.Fatal("expected %s to equal %s", arg, other)
	}
	if other := Float(3.14); arg.Equal(other) {
		t.Fatal("expected %s to not equal %s", arg, other)
	}
	if other := String("foo"); arg.Equal(other) {
		t.Fatal("expected %s to not equal %s", arg, other)
	}
	if expected, got := float32(0), f; expected != got {
		t.Fatalf("expected %f, got %f", expected, got)
	}
	if _, err := arg.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if expected, got := "Float(0.000000)", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
	if expected, got := TypetagFloat, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestBool(t *testing.T) {
	arg := Bool(false)
	b, err := arg.ReadBool()
	if err != nil {
		t.Fatal(err)
	}
	if other := Bool(false); !arg.Equal(other) {
		t.Fatal("expected %s to equal %s", arg, other)
	}
	if other := Int(3); arg.Equal(other) {
		t.Fatal("expected %s to not equal %s", arg, other)
	}
	if expected, got := false, b; expected != got {
		t.Fatalf("expected %f, got %f", expected, got)
	}
	if _, err := arg.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if expected, got := "Bool(false)", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
	if expected, got := TypetagFalse, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}

	argTrue := Bool(true)
	if expected, got := TypetagTrue, argTrue.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}

	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestString(t *testing.T) {
	arg := String("foo")
	s, err := arg.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if other := String("foo"); !arg.Equal(other) {
		t.Fatalf("expected %s to equal %s", arg, other)
	}
	if other := Int(4); arg.Equal(other) {
		t.Fatal("expected %s to not equal %s", arg, other)
	}
	if expected, got := "foo", s; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
	if _, err := arg.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if expected, got := "foo", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
	if expected, got := TypetagString, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestBlob(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	b, err := arg.ReadBlob()
	if err != nil {
		t.Fatal(err)
	}
	if other := Blob([]byte{'f', 'o', 'o'}); !arg.Equal(other) {
		t.Fatalf("expected %s to equal %s", arg, other)
	}
	if other := Blob([]byte{'f'}); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
	if other := String("bar"); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
	if expected, got := []byte{'f', 'o', 'o'}, b; !bytes.Equal(expected, got) {
		t.Fatalf("expected %f, got %f", expected, got)
	}
	if _, err := arg.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if expected, got := base64.StdEncoding.EncodeToString(b), arg.String(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
	if expected, got := TypetagBlob, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestParseArgument(t *testing.T) {
	type Input struct {
		tt   byte
		data []byte
	}
	type Output struct {
		Argument Argument
		Consumed int64 // number of bytes consumed
		Err      error
	}
	for _, testcase := range []struct {
		Input    Input
		Expected Output
	}{
		{
			Input: Input{
				tt:   'i',
				data: []byte{0, 0, 0, 1},
			},
			Expected: Output{
				Argument: Int(1),
				Consumed: 4,
			},
		},
		{
			Input: Input{
				tt:   'f',
				data: []byte{0x40, 0x48, 0xf5, 0xc3},
			},
			Expected: Output{
				Argument: Float(3.14),
				Consumed: 4,
			},
		},
		{
			Input: Input{tt: 'T'},
			Expected: Output{
				Argument: Bool(true),
			},
		},
		{
			Input: Input{tt: 'F'},
			Expected: Output{
				Argument: Bool(false),
			},
		},
		{
			Input: Input{
				tt:   's',
				data: []byte{'a', 'b', 'c', 'd', 'e'},
			},
			Expected: Output{
				Argument: String("abcde"),
				Consumed: 8,
			},
		},
		{
			Input: Input{
				tt: 'b',
				// Length followed by blob
				data: []byte{0, 0, 0, 5, 'a', 'b', 'c', 'd', 'e'},
			},
			Expected: Output{
				Argument: Blob([]byte{'a', 'b', 'c', 'd', 'e', 0, 0, 0}),
				Consumed: 12,
			},
		},
		{
			Input:    Input{tt: 'b', data: []byte{}},
			Expected: Output{Err: io.EOF},
		},
		{
			Input:    Input{tt: 'Q'},
			Expected: Output{Err: ErrInvalidTypeTag},
		},
	} {
		a, consumed, err := ParseArgument(testcase.Input.tt, testcase.Input.data)
		if expected, got := testcase.Expected.Err, err; expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
		if testcase.Expected.Err == nil {
			if expected, got := testcase.Expected.Consumed, consumed; expected != got {
				t.Fatalf("expected %d, got %d", expected, got)
			}
			if expected, got := testcase.Expected.Argument, a; !expected.Equal(got) {
				t.Fatalf("expected %+v, got %+v", expected, got)
			}
		}
	}
}
