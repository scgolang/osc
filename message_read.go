package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// ReadInt32 reads an int32 value from an OSC message.
func (msg *Message) ReadInt32(index int) (val int32, err error) {
	if tt := msg.Typetag[index]; tt != typetagInt {
		return 0, fmt.Errorf("Unexpected type %c", tt)
	}
	if index >= len(msg.Args) {
		return 0, ErrIndexOutOfBounds
	}
	r := bytes.NewReader(msg.Args[index])
	if err := binary.Read(r, byteOrder, &val); err != nil {
		return 0, err
	}
	return val, nil
}

// ReadFloat32 reads a float32 value from an OSC message.
func (msg *Message) ReadFloat32(index int) (float32, error) {
	if tt := msg.Typetag[index]; tt != typetagFloat {
		return 0, fmt.Errorf("Unexpected type %c", tt)
	}
	if index >= len(msg.Args) {
		return 0, ErrIndexOutOfBounds
	}
	var val float32
	r := bytes.NewReader(msg.Args[index])
	if err := binary.Read(r, byteOrder, &val); err != nil {
		return 0, err
	}
	return val, nil
}

// ReadBool reads a boolean value from an OSC message.
func (msg *Message) ReadBool(index int) (bool, error) {
	if index >= len(msg.Typetag) {
		return false, ErrIndexOutOfBounds
	}
	tt := msg.Typetag[index]
	if tt != typetagTrue && tt != typetagFalse {
		return false, fmt.Errorf("Unexpected type %c", tt)
	}
	return tt == typetagTrue, nil
}

// ReadString reads a string value from an OSC message.
func (msg *Message) ReadString(index int) (string, error) {
	if tt := msg.Typetag[index]; tt != typetagString {
		return "", fmt.Errorf("Unexpected type %c", tt)
	}
	if index >= len(msg.Args) {
		return "", ErrIndexOutOfBounds
	}
	return string(msg.Args[index]), nil
}

// ReadBlob reads a binary blob from an OSC message.
func (msg *Message) ReadBlob(index int) ([]byte, error) {
	if tt := msg.Typetag[index]; tt != typetagBlob {
		return nil, fmt.Errorf("Unexpected type %c", tt)
	}
	if index >= len(msg.Args) {
		return nil, ErrIndexOutOfBounds
	}
	return msg.Args[index], nil
}
