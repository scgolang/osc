package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// Represents a single OSC message. An OSC message consists of an OSC address
// pattern and zero or more arguments.
type Message struct {
	address   string
	arguments []interface{}
	sender    net.Addr
}

// NewMessage creates a new OSC message.
func NewMessage(addr string) *Message {
	return &Message{address: addr}
}

// Sender returns the address from which a message was sent.
func (msg *Message) Sender() net.Addr {
	return msg.sender
}

// Equals determines if the given OSC Message b is equal to the current OSC Message.
// It checks if the OSC address and the arguments are equal.
func (msg *Message) Equals(b *Message) bool {
	// Check OSC address
	if msg.address != b.address {
		return false
	}

	// Check if the number of arguments are equal
	if msg.CountArguments() != b.CountArguments() {
		return false
	}

	// Check arguments
	for i, arg := range msg.arguments {
		switch arg.(type) {
		case bool, int32, int64, float32, float64, string:
			if arg != b.arguments[i] {
				return false
			}

		case []byte:
			ba := arg.([]byte)
			bb := b.arguments[i].([]byte)
			if !bytes.Equal(ba, bb) {
				return false
			}

		case Timetag:
			if arg.(Timetag) != b.arguments[i].(Timetag) {
				return false
			}
		}
	}

	return true
}

// Append appends the given argument to the arguments list.
func (msg *Message) Append(arguments ...interface{}) {
	msg.arguments = append(msg.arguments, arguments...)
}

// Clear clears the OSC address and all arguments.
func (msg *Message) Clear() {
	msg.ClearData()
}

// ClearData removes all arguments from the OSC Message.
func (msg *Message) ClearData() {
	msg.arguments = msg.arguments[len(msg.arguments):]
}

// Returns true, if the address of the OSC Message matches the given address.
// Case sensitive!
func (msg *Message) Match(address string) bool {
	exp := getRegEx(msg.address)

	if exp.MatchString(address) {
		return true
	}

	return false
}

// TypeTags returns the type tag string.
func (msg *Message) TypeTags() (tags string, err error) {
	tags = ","
	for _, m := range msg.arguments {
		s, err := getTypeTag(m)
		if err != nil {
			return "", err
		}
		tags += s
	}

	return tags, nil
}

// CountArguments returns the number of arguments.
func (msg *Message) CountArguments() int {
	return len(msg.arguments)
}

// ToByteArray serializes the OSC message to a byte buffer.
// The byte buffer is of the following format:
// 1. OSC Address Pattern
// 2. OSC Type Tag String
// 3. OSC Arguments
func (msg *Message) ToByteArray() (buffer []byte, err error) {
	// The byte buffer for the message
	var data = &bytes.Buffer{}

	// We can start with the OSC address and add it to the buffer
	_, err = writePaddedString(msg.address, data)
	if err != nil {
		return nil, err
	}

	// Type tag string starts with ","
	typetags := []byte{','}

	// Process the type tags and collect all arguments
	var payload = new(bytes.Buffer)
	for _, arg := range msg.arguments {
		// FIXME: Use t instead of arg
		switch t := arg.(type) {
		default:
			return nil, errors.New(fmt.Sprintf("OSC - unsupported type: %T", t))

		case bool:
			if arg.(bool) == true {
				typetags = append(typetags, 'T')
			} else {
				typetags = append(typetags, 'F')
			}

		case nil:
			typetags = append(typetags, 'N')

		case int32:
			typetags = append(typetags, 'i')

			if err = binary.Write(payload, binary.BigEndian, int32(t)); err != nil {
				return nil, err
			}

		case float32:
			typetags = append(typetags, 'f')

			if err = binary.Write(payload, binary.BigEndian, float32(t)); err != nil {
				return nil, err
			}

		case string:
			typetags = append(typetags, 's')

			if _, err = writePaddedString(t, payload); err != nil {
				return nil, err
			}

		case []byte:
			typetags = append(typetags, 'b')

			if _, err = writeBlob(t, payload); err != nil {
				return nil, err
			}

		case int64:
			typetags = append(typetags, 'h')

			if err = binary.Write(payload, binary.BigEndian, int64(t)); err != nil {
				return nil, err
			}

		case float64:
			typetags = append(typetags, 'd')

			if err = binary.Write(payload, binary.BigEndian, float64(t)); err != nil {
				return nil, err
			}

		case Timetag:
			typetags = append(typetags, 't')

			timeTag := arg.(Timetag)
			payload.Write(timeTag.ToByteArray())
		}
	}

	// Write the type tag string to the data buffer
	_, err = writePaddedString(string(typetags), data)
	if err != nil {
		return nil, err
	}

	// Write the payload (OSC arguments) to the data buffer
	data.Write(payload.Bytes())

	return data.Bytes(), nil
}

// Write pretty prints a Message to the standard output.
func (msg *Message) Write(w io.Writer) error {
	tags, err := msg.TypeTags()
	if err != nil {
		return err
	}

	var formatString string
	var arguments []interface{}
	formatString += "%s %s"
	arguments = append(arguments, msg.address)
	arguments = append(arguments, tags)

	for _, arg := range msg.arguments {
		switch arg.(type) {
		case bool, int32, int64, float32, float64, string:
			formatString += " %v"
			arguments = append(arguments, arg)

		case nil:
			formatString += " %s"
			arguments = append(arguments, "Nil")

		case []byte:
			formatString += " %s"
			arguments = append(arguments, "blob")

		case Timetag:
			formatString += " %d"
			timeTag := arg.(Timetag)
			arguments = append(arguments, uint64(timeTag))
		}
	}
	fmt.Fprintln(w, fmt.Sprintf(formatString, arguments...))
	return nil
}
