package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// Packet is the interface for Message and Bundle.
type Packet interface {
	ToByteArray() (buffer []byte, err error)
}

// Represents a single OSC message. An OSC message consists of an OSC address
// pattern and zero or more arguments.
type Message struct {
	Address   string
	Arguments []interface{}
}

// An OSC Bundle consists of the OSC-string "#bundle" followed by an OSC Time Tag,
// followed by zero or more OSC bundle/message elements. The OSC-timetag is a 64-bit fixed
// point time tag. See http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag  Timetag
	Messages []*Message
	Bundles  []*Bundle
}

// NewMessage returns a new Message. The address parameter is the OSC address.
func NewMessage(address string) (msg *Message) {
	return &Message{Address: address}
}

// Append appends the given argument to the arguments list.
func (msg *Message) Append(argument interface{}) {
	msg.Arguments = append(msg.Arguments, argument)
}

// Equals determines if the given OSC Message b is equal to the current OSC Message.
// It checks if the OSC address and the arguments are equal. Returns, true if the
// current object and b are equal.
func (msg *Message) Equals(b *Message) bool {
	// Check OSC address
	if msg.Address != b.Address {
		return false
	}

	// Check if the number of arguments are equal
	if msg.CountArguments() != b.CountArguments() {
		return false
	}

	// Check arguments
	for i, arg := range msg.Arguments {
		switch arg.(type) {
		case bool, int32, int64, float32, float64, string:
			if arg != b.Arguments[i] {
				return false
			}

		case []byte:
			ba := arg.([]byte)
			bb := b.Arguments[i].([]byte)
			if !bytes.Equal(ba, bb) {
				return false
			}

		case Timetag:
			if arg.(*Timetag).TimeTag() != b.Arguments[i].(*Timetag).TimeTag() {
				return false
			}
		}
	}

	return true
}

// Clear clears the OSC address and all arguments.
func (msg *Message) Clear() {
	msg.Address = ""
	msg.ClearData()
}

// ClearData removes all arguments from the OSC Message.
func (msg *Message) ClearData() {
	msg.Arguments = msg.Arguments[len(msg.Arguments):]
}

// Returns true, if the address of the OSC Message matches the given address.
// Case sensitive!
func (msg *Message) Match(address string) bool {
	exp := getRegEx(msg.Address)

	if exp.MatchString(address) {
		return true
	}

	return false
}

// TypeTags returns the type tag string.
func (msg *Message) TypeTags() (tags string, err error) {
	tags = ","
	for _, m := range msg.Arguments {
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
	return len(msg.Arguments)
}

// ToByteBuffer serializes the OSC message to a byte buffer. The byte buffer
// is of the following format:
// 1. OSC Address Pattern
// 2. OSC Type Tag String
// 3. OSC Arguments
func (msg *Message) ToByteArray() (buffer []byte, err error) {
	// The byte buffer for the message
	var data = new(bytes.Buffer)

	// We can start with the OSC address and add it to the buffer
	_, err = writePaddedString(msg.Address, data)
	if err != nil {
		return nil, err
	}

	// Type tag string starts with ","
	typetags := []byte{','}

	// Process the type tags and collect all arguments
	var payload = new(bytes.Buffer)
	for _, arg := range msg.Arguments {
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

////
// Bundle
////

// NewBundle returns an OSC Bundle. Use this function to create a new OSC
// Bundle.
func NewBundle(time time.Time) (bundle *Bundle) {
	return &Bundle{Timetag: *NewTimetag(time)}
}

// Append appends an OSC bundle or OSC message to the bundle.
func (self *Bundle) Append(pkt Packet) (err error) {
	switch t := pkt.(type) {
	default:
		return errors.New(fmt.Sprintf("Unsupported OSC packet type: only Bundle and Message are supported.", t))

	case *Bundle:
		self.Bundles = append(self.Bundles, t)

	case *Message:
		self.Messages = append(self.Messages, t)
	}

	return nil
}

// ToByteArray serializes the OSC bundle to a byte array with the following format:
// 1. Bundle string: '#bundle'
// 2. OSC timetag
// 3. Length of first OSC bundle element
// 4. First bundle element
// 5. Length of n OSC bundle element
// 6. n bundle element
func (self *Bundle) ToByteArray() (buffer []byte, err error) {
	var data = new(bytes.Buffer)

	// Add the '#bundle' string
	_, err = writePaddedString("#bundle", data)
	if err != nil {
		return nil, err
	}

	// Add the timetag
	if _, err = data.Write(self.Timetag.ToByteArray()); err != nil {
		return nil, err
	}

	// Process all OSC Messages
	for _, m := range self.Messages {
		var msgLen int
		var msgBuf []byte

		msgBuf, err = m.ToByteArray()
		if err != nil {
			return nil, err
		}

		// Append the length of the OSC message
		msgLen = len(msgBuf)
		if err = binary.Write(data, binary.BigEndian, int32(msgLen)); err != nil {
			return nil, err
		}

		// Append the OSC message
		data.Write(msgBuf)
	}

	// Process all OSC Bundles
	for _, b := range self.Bundles {
		var bLen int
		var bBuf []byte

		bBuf, err = b.ToByteArray()
		if err != nil {
			return nil, err
		}

		// Write the size of the bundle
		bLen = len(bBuf)
		if err = binary.Write(data, binary.BigEndian, int32(bLen)); err != nil {
			return nil, err
		}

		// Append the bundle
		data.Write(bBuf)
	}

	return data.Bytes(), nil
}

// PrintMessages pretty prints an Message to the standard output.
func PrintMessage(msg *Message) {
	tags, err := msg.TypeTags()
	if err != nil {
		return
	}

	var formatString string
	var arguments []interface{}
	formatString += "%s %s"
	arguments = append(arguments, msg.Address)
	arguments = append(arguments, tags)

	for _, arg := range msg.Arguments {
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
			arguments = append(arguments, timeTag.TimeTag())
		}
	}
	fmt.Println(fmt.Sprintf(formatString, arguments...))
}
