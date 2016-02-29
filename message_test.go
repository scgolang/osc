package osc

import (
	"bytes"
	"testing"
)

const testBlob = `Able was I ere I saw Elba.`

func TestTypeTagsString(t *testing.T) {
	msg, err := NewMessage("/some/address")
	if err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteInt32(0, 100); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(1, true); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(2, false); err != nil {
		t.Fatal(err)
	}
	if expected, got := "iTF", string(msg.Typetag); expected != got {
		t.Fatalf("Expected %s got %s", expected, got)
	}
}

// writeTestValues
func writeTestValues(msg *Message, t *testing.T) {
	if err := msg.WriteInt32(0, 20); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteFloat32(1, 3.14); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteString(2, "foobar"); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBlob(3, []byte{}); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBlob(4, []byte(testBlob)); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(5, true); err != nil {
		t.Fatal(err)
	}
}

// readTestValues
func readTestValues(msg *Message, t *testing.T) {
	intval, err := msg.ReadInt32(0)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := int32(20), intval; expected != got {
		t.Fatalf("Expected %d got %d", expected, got)
	}
	floatval, err := msg.ReadFloat32(1)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := float32(3.14), floatval; expected != got {
		t.Fatalf("Expected %f got %f", expected, got)
	}
	strval, err := msg.ReadString(2)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := "foobar", strval; expected != got {
		t.Fatalf("Expected %s got %s", expected, got)
	}
	blobval, err := msg.ReadBlob(4)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := []byte(testBlob), blobval; 0 != bytes.Compare(expected, got) {
		t.Fatalf("Expected %q got %q", expected, got)
	}
	boolval, err := msg.ReadBool(5)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := true, boolval; expected != got {
		t.Fatalf("Expected %t got %t", expected, got)
	}
}

func TestWriteRead(t *testing.T) {
	// create a message and write some data
	msg, err := NewMessage("/some/address")
	if err != nil {
		t.Fatal(err)
	}
	writeTestValues(msg, t)
	readTestValues(msg, t)
}

func TestVerifyParts(t *testing.T) {
	// Pairs that should match
	for _, pair := range [][2]string{
		{"/osc/address", "/osc/address"},
		{"/path/to/method", "/path/to/meth?d"},
	} {
		if !verifyParts(pair[0], pair[1]) {
			t.Fatalf("Expected %s to have the same parts as %s", pair[0], pair[1])
		}
	}
}
