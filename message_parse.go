package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// parseMessage parses an OSC message from a slice of bytes.
func parseMessage(data []byte) (*Message, error) {
	var (
		i = 0
		n = len(data)
	)
	for i < n && data[i] != 0 {
		i++
	}
	// return an error if we've reached the end of the data
	if i == n {
		logger.Println("could not read address")
		return nil, ErrIncomplete
	}
	msg, err := NewMessage(string(data[0:i]))
	if err != nil {
		logger.Println("could not create message")
		return nil, err
	}
	i++ // advance past the null byte

	// advance i to the next multiple of 4
	for i%4 != 0 {
		i++
	}
	// return an error if we've reached the end of the data
	if i >= n {
		return nil, ErrIncomplete
	}

	// Read all arguments
	if err := msg.parseArguments(data[i:]); err != nil {
		logger.Printf("error parsing data %q\n", data)
		return nil, err
	}
	return msg, nil
}

// parseArguments reads all arguments from the reader and adds it to the OSC message.
func (msg *Message) parseArguments(data []byte) error {
	if len(data) == 0 || data[0] != typetagPrefix {
		return fmt.Errorf("invalid type tag: %q", data[0])
	}
	var (
		n       = len(data)
		i       = 1 // skip the typetag prefix
		ttstart = 1
	)
	// read the typetag
	for i < n && data[i] != 0 {
		i++
	}
	// return an error if we've reached the end of the data
	if i == n {
		logger.Println("incomplete message reading typetag")
		return ErrIncomplete
	}
	msg.Typetag = data[ttstart:i]
	i++ // advance past the null byte
	for i%4 != 0 {
		i++
	}
	// return an error if we've reached the end of the data
	if i >= n {
		logger.Println("incomplete message after reading typetag")
		return ErrIncomplete
	}
	// allocate storage for the arguments, then read the arguments
	msg.Args = make([][]byte, 0)
	for _, tt := range msg.Typetag {
		if i >= n {
			logger.Println("incomplete message while reading arguments")
			return ErrIncomplete
		}
		n, err := msg.parseArgForTypetag(tt, data[i:])
		if err != nil {
			return err
		}
		i += n
	}
	return nil
}

// parseArgForTypetag
func (msg *Message) parseArgForTypetag(tt byte, data []byte) (int, error) {
	n := len(data)
	// return an error if we've reached the end of the data
	switch tt {
	default:
		return 0, ErrParse
	case typetagInt, typetagFloat:
		// return an error if we've reached the end of the data
		if n < 4 {
			logger.Println("incomplete message while reading numeric argument")
			return 0, ErrIncomplete
		}
		msg.Args = append(msg.Args, data[:4])
		return 4, nil
	case typetagString:
		n, str := msg.parseString(data)
		msg.Args = append(msg.Args, str)
		return n, nil
	case typetagBlob:
		n, blob, err := msg.parseBlob(data)
		if err != nil {
			return 0, err
		}
		msg.Args = append(msg.Args, blob)
		return n, nil
	case typetagTrue, typetagFalse:
		msg.Args = append(msg.Args, nil)
		return 0, nil
	}
}

// parseString
func (msg *Message) parseString(data []byte) (int, []byte) {
	var i int
	for _, b := range data {
		if b == 0 {
			break
		}
		i++
	}
	result := data[:i]
	for i%4 != 0 {
		i++
	}
	return i, result
}

// parseBlob parses a binary blob.
func (msg *Message) parseBlob(data []byte) (int, []byte, error) {
	var (
		i = 0
		n = len(data)
	)
	// return an error if we'll reach the end of the data
	// by reading the blob length
	if n < 4 {
		logger.Println("could not parse blob (not enough bytes for length)")
		return 0, nil, ErrIncomplete
	}
	r := bytes.NewReader(data[i : i+4])
	var bl int32
	if err := binary.Read(r, byteOrder, &bl); err != nil {
		return 0, nil, err
	}
	i += 4
	// return an error if we've reached the end of the data
	if i+int(bl) > n {
		logger.Printf("expected at least %d bytes in blob, found %d)", i+int(bl), n)
		return 0, nil, ErrIncomplete
	}
	return i + int(bl), data[i : i+int(bl)], nil
}
