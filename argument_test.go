package osc

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestInt(t *testing.T) {
	arg := Int(0)
	i, err := arg.ReadInt32()
	if err != nil {
		t.Fatal(err)
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
	if expected, got := "foo", s; expected != got {
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
	if _, err := arg.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
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
	if expected, got := TypetagBlob, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestParseArgument(t *testing.T) {
}
