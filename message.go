package osc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

// Common errors.
var (
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrInvalidTypeTag   = errors.New("invalid type tag")
	ErrNilWriter        = errors.New("writer must not be nil")
	ErrParse            = errors.New("error parsing message")
)

// Message is an OSC message.
// An OSC message consists of an OSC address pattern and zero or more arguments.
type Message struct {
	Address   string `json:"address"`
	Arguments []Argument
	Sender    net.Addr
}

// NewMessage creates a new OSC message.
func NewMessage(addr string) (*Message, error) {
	return &Message{
		Address: addr,
	}, nil
}

// Match returns true, if the address of the OSC Message matches the given address.
// Case sensitive!
func (msg *Message) Match(address string) (bool, error) {
	// verify same number of parts
	if !verifyParts(address, msg.Address) {
		return false, nil
	}

	exp, err := getRegex(msg.Address)
	if err != nil {
		return false, err
	}
	return exp.MatchString(address), nil
}

// Contents returns the contents of the message as a slice of bytes.
func (msg *Message) Contents() ([]byte, error) {
	w := &bytes.Buffer{}

	// Write address
	if _, err := w.Write(OscString(msg.Address)); err != nil {
		return nil, err
	}

	// Write the typetags.
	if _, err := w.Write(msg.typetags()); err != nil {
		return nil, err
	}

	// Write arguments
	// for _, a := range msg.Arguments {
	// }

	return w.Bytes(), nil
}

// typetags returns a padded byte slice of the message's type tags.
func (msg *Message) typetags() []byte {
	tt := make([]byte, len(msg.Arguments))
	for i, a := range msg.Arguments {
		tt[i] = a.Typetag()
	}
	return Pad(tt)
}

// WriteTo writes the Message to an io.Writer.
func (msg *Message) Print(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s%s", msg.Address, msg.typetags()); err != nil {
		return err
	}

	for _, a := range msg.Arguments {
		if _, err := a.WriteTo(w); err != nil {
			return err
		}
	}

	return nil
}

// parseMessage parses an OSC message from a slice of bytes.
func parseMessage(data []byte, sender net.Addr) (*Message, error) {
	address, idx := ReadString(data)
	msg := &Message{
		Address: address,
		Sender:  sender,
	}

	// Read all arguments
	if err := msg.parseArguments(data[idx:]); err != nil {
		return nil, err
	}

	return msg, nil
}

// parseArguments reads all arguments from the reader and adds it to the OSC message.
func (msg *Message) parseArguments(data []byte) error {
	if len(data) == 0 || data[0] != TypetagPrefix {
		return ErrInvalidTypeTag
	}
	// tt, idx := readString(data[1:]) // strip the prefix

	return nil
}

// verifyParts verifies that m1 and m2 have the same number of parts,
// where a part is a nonempty string between pairs of '/' or a nonempty
// string at the end.
func verifyParts(m1, m2 string) bool {
	if m1 == m2 {
		return true
	}

	mc := string(MessageChar)

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
