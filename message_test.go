package osc

import (
	"bytes"
	"testing"
)

func TestTypeTagsString(t *testing.T) {
	msg := NewMessage("/some/address")
	if err := msg.WriteInt32(100); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(true); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(false); err != nil {
		t.Fatal(err)
	}
	if expected, got := ",iTF", msg.TypeTags(); expected != got {
		t.Fatalf("Expected %s got %s", expected, got)
	}
}

func TestWriteRead(t *testing.T) {
	const blob = `Able was I ere I saw Elba.`

	// create a message and write some data
	msg := NewMessage("/some/address")
	if err := msg.WriteInt32(20); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteFloat32(3.14); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteString("foobar"); err != nil {
		t.Fatal(err)
	}
	// writing an empty blob should have no effect whatsoever
	if err := msg.WriteBlob([]byte{}); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBlob([]byte(blob)); err != nil {
		t.Fatal(err)
	}
	if err := msg.WriteBool(true); err != nil {
		t.Fatal(err)
	}

	// read the data from a clone
	clone, err := msg.clone()
	if err != nil {
		t.Fatal(err)
	}

	if intval, err := clone.ReadInt32(); err != nil {
		if err != nil {
			t.Fatal(err)
		}
		if expected, got := int32(20), intval; expected != got {
			t.Fatalf("Expected %d got %d", expected, got)
		}
	}
	if floatval, err := clone.ReadFloat32(); err != nil {
		if err != nil {
			t.Fatal(err)
		}
		if expected, got := float32(3.14), floatval; expected != got {
			t.Fatalf("Expected %f got %f", expected, got)
		}
	}
	if strval, err := clone.ReadString(); err != nil {
		if err != nil {
			t.Fatal(err)
		}
		if expected, got := "foobar", strval; expected != got {
			t.Fatalf("Expected %s got %s", expected, got)
		}
	}
	if blobval, err := clone.ReadBlob(); err != nil {
		if err != nil {
			t.Fatal(err)
		}
		if expected, got := []byte(blob), blobval; 0 != bytes.Compare(expected, got) {
			t.Fatalf("Expected %q got %q", expected, got)
		}
	}
	if boolval, err := clone.ReadBool(); err != nil {
		if err != nil {
			t.Fatal(err)
		}
		if expected, got := true, boolval; expected != got {
			t.Fatalf("Expected %t got %t", expected, got)
		}
	}
}
