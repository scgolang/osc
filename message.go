package osc

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"

	"github.com/pkg/errors"
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

// ParseMessage parses an OSC message from a slice of bytes.
func ParseMessage(data []byte, sender net.Addr) (Message, error) {
	address, idx := ReadString(data)
	msg := Message{
		Address: address,
		Sender:  sender,
	}
	data = data[idx:]
	typetags, idx := ReadString(data)
	data = data[idx:]

	// Read all arguments.
	args, err := ReadArguments([]byte(typetags), data)
	if err != nil {
		return Message{}, errors.Wrap(err, "parse message")
	}
	msg.Arguments = args

	return msg, nil
}

// Bytes returns the contents of the message as a slice of bytes.
func (msg Message) Bytes() []byte {
	b := [][]byte{
		ToBytes(msg.Address),
		msg.Typetags(),
	}
	for _, a := range msg.Arguments {
		b = append(b, a.Bytes())
	}
	return bytes.Join(b, []byte{})
}

// Equal returns true if the messages are equal, false otherwise.
func (msg Message) Equal(other Packet) bool {
	msg2, ok := other.(Message)
	if !ok {
		return false
	}
	if msg.Address != msg2.Address {
		return false
	}
	if len(msg.Arguments) != len(msg2.Arguments) {
		return false
	}
	for i, a := range msg.Arguments {
		if !a.Equal(msg2.Arguments[i]) {
			return false
		}
	}
	return true
}

// Match returns true if the address of the OSC Message matches the given address.
func (msg Message) Match(address string) (bool, error) {
	// Verify same number of parts.
	if !VerifyParts(address, msg.Address) {
		return false, nil
	}
	exp, err := GetRegex(msg.Address)
	if err != nil {
		return false, err
	}
	return exp.MatchString(address), nil
}

// Typetags returns a padded byte slice of the message's type tags.
func (msg Message) Typetags() []byte {
	tt := make([]byte, len(msg.Arguments)+1)
	tt[0] = TypetagPrefix
	for i, a := range msg.Arguments {
		tt[i+1] = a.Typetag()
	}
	return Pad(append(tt, 0))
}

// WriteTo writes the Message to an io.Writer.
func (msg Message) WriteTo(w io.Writer) (int64, error) {
	var bytesWritten int

	nw, err := fmt.Fprintf(w, "%s%s", msg.Address, msg.Typetags())
	if err != nil {
		return 0, err
	}
	bytesWritten += nw

	for _, a := range msg.Arguments {
		nw, err := a.WriteTo(w)
		if err != nil {
			return 0, err
		}
		bytesWritten += int(nw)
	}

	return int64(bytesWritten), nil
}

// GetRegex compiles and returns a regular expression object for the given address pattern.
func GetRegex(pattern string) (*regexp.Regexp, error) {
	pattern = strings.Replace(pattern, ".", "\\.", -1) // Escape all '.' in the pattern
	pattern = strings.Replace(pattern, "(", "\\(", -1) // Escape all '(' in the pattern
	pattern = strings.Replace(pattern, ")", "\\)", -1) // Escape all ')' in the pattern
	pattern = strings.Replace(pattern, "*", ".*", -1)  // Replace a '*' with '.*' that matches zero or more characters
	pattern = strings.Replace(pattern, "{", "(", -1)   // Change a '{' to '('
	pattern = strings.Replace(pattern, ",", "|", -1)   // Change a ',' to '|'
	pattern = strings.Replace(pattern, "}", ")", -1)   // Change a '}' to ')'
	pattern = strings.Replace(pattern, "?", ".", -1)   // Change a '?' to '.'
	pattern = "^" + pattern + "$"
	return regexp.Compile(pattern)
}

// VerifyParts verifies that m1 and m2 have the same number of parts,
// where a part is a nonempty string between pairs of '/' or a nonempty
// string at the end.
func VerifyParts(m1, m2 string) bool {
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
