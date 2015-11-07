package osc

import (
	"bufio"
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

	message.Append("string argument")
	message.Append(123456789)
	message.Append(true)

	if message.CountArguments() != 3 {
		t.Errorf("Number of arguments should be %d and is %d", 3, message.CountArguments())
	}
}

func TestEqualMessage(t *testing.T) {
	var (
		msg1 = NewMessage("/address")
		msg2 = NewMessage("/address")
	)

	msg1.Append(1234)
	msg2.Append(1234)
	msg1.Append("test string")
	msg2.Append("test string")

	if !msg1.Equals(msg2) {
		t.Error("Messages should be equal")
	}
}

func TestReadPaddedString(t *testing.T) {
	buf1 := []byte{'t', 'e', 's', 't', 's', 't', 'r', 'i', 'n', 'g', 0, 0}
	buf2 := []byte{'t', 'e', 's', 't', 0, 0, 0, 0}

	bytesBuffer := bytes.NewBuffer(buf1)
	st, n, err := readPaddedString(bufio.NewReader(bytesBuffer))
	if err != nil {
		t.Error("Error reading padded string: " + err.Error())
	}

	if n != 12 {
		t.Errorf("Number of bytes needs to be 12 and is: %d\n", n)
	}

	if st != "teststring" {
		t.Errorf("String should be \"teststring\" and is \"%s\"", st)
	}

	bytesBuffer = bytes.NewBuffer(buf2)
	st, n, err = readPaddedString(bufio.NewReader(bytesBuffer))
	if err != nil {
		t.Error("Error reading padded string: " + err.Error())
	}

	if n != 8 {
		t.Errorf("Number of bytes needs to be 8 and is: %d\n", n)
	}

	if st != "test" {
		t.Errorf("String should be \"test\" and is \"%s\"", st)
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

func TestPadBytesNeeded(t *testing.T) {
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
	msg.Append(int32(100))
	msg.Append(true)
	msg.Append(false)

	typeTags, err := msg.TypeTags()
	if err != nil {
		t.Error(err.Error())
	}

	if typeTags != ",iTF" {
		t.Errorf("Type tag string should be ',iTF' and is: %s", typeTags)
	}
}
