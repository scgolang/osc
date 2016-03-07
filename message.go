package osc

import (
	"bytes"
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
