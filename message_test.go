package osc

import (
	"bytes"
	"testing"
)

func TestAppendArguments(t *testing.T) {
	var (
		oscAddress = "/address"
		message    = NewMessage(oscAddress)
	)
	if message.address != oscAddress {
		t.Errorf("OSC address should be \"%s\" and is \"%s\"", oscAddress, message.address)
	}

	if err := message.WriteString("string argument"); err != nil {
		t.Fatal(err)
	}
	if err := message.WriteInt32(123456789); err != nil {
		t.Fatal(err)
	}
	if err := message.WriteBool(true); err != nil {
		t.Fatal(err)
	}
}

func TestEqualMessage(t *testing.T) {
	var (
		msg1 = NewMessage("/address")
		msg2 = NewMessage("/address")
	)

	if err := msg1.WriteInt32(1234); err != nil {
		t.Fatal(err)
	}
	if err := msg2.WriteInt32(1234); err != nil {
		t.Fatal(err)
	}
	if err := msg1.WriteString("test string"); err != nil {
		t.Fatal(err)
	}
	if err := msg2.WriteString("test string"); err != nil {
		t.Fatal(err)
	}
}

func TestWritePaddedString(t *testing.T) {
	buf := []byte{}
	bytesBuffer := bytes.NewBuffer(buf)
	testString := "testString"
	expectedNumberOfWrittenBytes := len(testString) + padBytesNeeded(len(testString))

	n, err := writePaddedString(testString, bytesBuffer)
	if err != nil {
		t.Errorf(err.Error())
	}

	if n != expectedNumberOfWrittenBytes {
		t.Errorf("Expected number of written bytes should be \"%d\" and is \"%d\"", expectedNumberOfWrittenBytes, n)
	}
}

func TestPadSize(t *testing.T) {
	var n int
	n = padBytesNeeded(4)
	if n != 4 {
		t.Errorf("Number of pad bytes should be 4 and is: %d", n)
	}

	n = padBytesNeeded(3)
	if n != 1 {
		t.Errorf("Number of pad bytes should be 1 and is: %d", n)
	}

	n = padBytesNeeded(1)
	if n != 3 {
		t.Errorf("Number of pad bytes should be 3 and is: %d", n)
	}

	n = padBytesNeeded(0)
	if n != 4 {
		t.Errorf("Number of pad bytes should be 4 and is: %d", n)
	}

	n = padBytesNeeded(32)
	if n != 4 {
		t.Errorf("Number of pad bytes should be 4 and is: %d", n)
	}

	n = padBytesNeeded(63)
	if n != 1 {
		t.Errorf("Number of pad bytes should be 1 and is: %d", n)
	}

	n = padBytesNeeded(10)
	if n != 2 {
		t.Errorf("Number of pad bytes should be 2 and is: %d", n)
	}
}

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
