package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Common errors.
var (
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrIncomplete       = errors.New("incomplete read")
	ErrInvalidTypeTag   = errors.New("invalid type tag")
	ErrNilWriter        = errors.New("writer must not be nil")
	ErrParse            = errors.New("error parsing message")
)

// Message is an OSC message.
// An OSC message consists of an OSC address pattern and zero or more arguments.
type Message struct {
	Address string
	Args    [][]byte
	Typetag []byte
}

// NewMessage creates a message.
func NewMessage(address string) (*Message, error) {
	return &Message{Address: address}, nil
}

// String converts a message to a string.
func (msg *Message) String() string {
	str := msg.Address + " ," + string(msg.Typetag)
	for _, arg := range msg.Args {
		str += " " + fmt.Sprintf("%q", arg)
	}
	return str
}

// Length returns the number of bytes in the message.
func (msg *Message) Length() int32 {
	var (
		addrlen = paddedLength(len(msg.Address))
		ttlen   = paddedLength(len(msg.Typetag))
	)
	var argslen, argsidx int
	for _, tt := range msg.Typetag {
		switch tt {
		case typetagString, typetagBlob:
			argslen += paddedLength(len(msg.Args[argsidx]))
			argsidx++
		case typetagInt, typetagFloat:
			argslen += 4
		}
	}
	return int32(addrlen + ttlen + argslen)
}

// Match returns true, if the address of the OSC Message matches the given address.
// Case sensitive!
func (msg *Message) Match(address string) (bool, error) {
	addr := string(msg.Address)

	// verify same number of parts
	if !verifyParts(address, addr) {
		return false, nil
	}

	exp, err := getRegex(addr)
	if err != nil {
		return false, err
	}
	return exp.MatchString(address), nil
}

// Compare compares one message to another.
// If they are the same it returns nil, otherwise it returns
// an error describing what is different about them.
func (msg *Message) Compare(other *Message) error {
	if mine, other := msg.Address, other.Address; mine != other {
		return fmt.Errorf("addresses different mine=%q other%q", mine, other)
	}
	if mine, other := msg.Typetag, other.Typetag; bytes.Compare(mine, other) != 0 {
		return fmt.Errorf("typetags different mine=%q other=%q", mine, other)
	}
	if mine, other := len(msg.Args), len(other.Args); mine != other {
		return fmt.Errorf("different number of args mine=%d other=%d", mine, other)
	}
	for i, mine := range msg.Args {
		theirs := other.Args[i]
		if mine == nil {
			if theirs != nil {
				return fmt.Errorf("arg %d is different mine=%q other=%q", i, mine, theirs)
			}
		}
		if bytes.Compare(mine, theirs) != 0 {
			return fmt.Errorf("arg %d is different mine=%q other=%q", i, mine, theirs)
		}
	}
	return nil
}

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
		return nil, ErrIncomplete
	}
	msg, err := NewMessage(string(data[0:i]))
	if err != nil {
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
		return ErrIncomplete
	}
	msg.Typetag = data[ttstart:i]
	i++ // advance past the null byte
	for i%4 != 0 {
		i++
	}
	// return an error if we've reached the end of the data
	if i >= n {
		return ErrIncomplete
	}
	// allocate storage for the arguments, then read the arguments
	msg.Args = make([][]byte, 0)
	for _, tt := range msg.Typetag {
		if i >= n {
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

func (msg *Message) parseArgForTypetag(tt byte, data []byte) (int, error) {
	n := len(data)
	// return an error if we've reached the end of the data
	switch tt {
	default:
		return 0, ErrParse
	case typetagInt, typetagFloat:
		// return an error if we've reached the end of the data
		if n < 4 {
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
		return 0, nil, ErrIncomplete
	}
	return int(bl) + 4, data[i:int(bl)], nil
}

// verifyParts verifies that m1 and m2 have the same number of parts,
// where a part is a nonempty string between pairs of '/' or a nonempty
// string at the end.
func verifyParts(m1, m2 string) bool {
	if m1 == m2 {
		return true
	}

	mc := string(messageChar)

	p1, p2 := strings.Split(m1, mc), strings.Split(m2, mc)
	if len(p1) != len(p2) || len(p1) == 0 {
		return false
	}
	for i, p := range p1[1:] {
		if len(p) == 0 || len(p2[i+1]) == 0 {
			return false
		}
	}

	return true
}
