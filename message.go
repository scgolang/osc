package osc

import (
	"bytes"
	"encoding/binary"
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
	address       []byte
	typetag       []byte
	argbuf        *bytes.Buffer
	ttReadIndex   int
	senderAddress net.Addr
}

// NewMessage creates a new OSC message.
func NewMessage(addr string) (*Message, error) {
	return &Message{
		address:     []byte(addr),
		argbuf:      &bytes.Buffer{},
		typetag:     []byte{typetagPrefix},
		ttReadIndex: 1, // skip the leading ','
	}, nil
}

// Address returns the address of the message.
func (msg *Message) Address() string {
	return string(msg.address)
}

// CountArguments returns the number of arguments in the OSC message.
func (msg *Message) CountArguments() int {
	return len(msg.typetag) - 1
}

// ReadInt32 reads an int32 value from an OSC message.
func (msg *Message) ReadInt32() (int32, error) {
	tt := msg.typetag[msg.ttReadIndex]
	if tt != typetagInt {
		return 0, fmt.Errorf("Unexpected type %c", tt)
	}
	var val int32
	if err := binary.Read(msg.argbuf, byteOrder, &val); err != nil {
		return 0, err
	}
	msg.ttReadIndex++
	return val, nil
}

// ReadFloat32 reads a float32 value from an OSC message.
func (msg *Message) ReadFloat32() (float32, error) {
	tt := msg.typetag[msg.ttReadIndex]
	if tt != typetagFloat {
		return 0, fmt.Errorf("Unexpected type %c", tt)
	}
	var val float32
	if err := binary.Read(msg.argbuf, byteOrder, &val); err != nil {
		return 0, err
	}
	msg.ttReadIndex++
	return val, nil
}

// ReadBool reads a boolean value from an OSC message.
func (msg *Message) ReadBool() (bool, error) {
	tt := msg.typetag[msg.ttReadIndex]
	if tt != typetagTrue && tt != typetagFalse {
		return false, fmt.Errorf("Unexpected type %c", tt)
	}
	msg.ttReadIndex++
	return tt == typetagTrue, nil
}

// ReadString reads a string value from an OSC message.
func (msg *Message) ReadString() (string, error) {
	tt := msg.typetag[msg.ttReadIndex]
	if tt != typetagString {
		return "", fmt.Errorf("Unexpected type %c", tt)
	}

	val := []byte{}
	for i := 0; i < msg.argbuf.Len(); i++ {
		c, err := msg.argbuf.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if c == 0 {
			// string values are padded to 32 bits by null bytes
			for j := i; j%4 != 0; j++ {
				if _, err := msg.argbuf.ReadByte(); err != nil && err != io.EOF {
					return "", err
				}
			}
			break
		}
		val = append(val, c)
	}

	msg.ttReadIndex++
	return string(val), nil
}

// ReadBlob reads a binary blob from an OSC message.
func (msg *Message) ReadBlob() ([]byte, error) {
	tt := msg.typetag[msg.ttReadIndex]
	if tt != typetagBlob {
		return nil, fmt.Errorf("Unexpected type %c", tt)
	}
	// read length
	var bl int32
	if err := binary.Read(msg.argbuf, byteOrder, &bl); err != nil {
		return nil, err
	}
	// read blob
	blob := make([]byte, bl)
	if _, err := msg.argbuf.Read(blob); err != nil {
		return nil, err
	}
	msg.ttReadIndex++
	return blob, nil
}

// WriteInt32 writes an int32 value to an OSC message.
func (msg *Message) WriteInt32(val int32) error {
	msg.typetag = append(msg.typetag, typetagInt)
	return binary.Write(msg.argbuf, byteOrder, val)
}

// WriteFloat32 writes a float32 value to an OSC message.
func (msg *Message) WriteFloat32(val float32) error {
	msg.typetag = append(msg.typetag, typetagFloat)
	return binary.Write(msg.argbuf, byteOrder, val)
}

// WriteBool writes a boolean value to an OSC message.
func (msg *Message) WriteBool(val bool) error {
	if val {
		msg.typetag = append(msg.typetag, typetagTrue)
	} else {
		msg.typetag = append(msg.typetag, typetagFalse)
	}
	return nil
}

// WriteString writes a string value to an OSC message.
func (msg *Message) WriteString(val string) error {
	msg.typetag = append(msg.typetag, typetagString)
	i := 0
	for _, c := range append([]byte(val), 0) {
		if err := msg.argbuf.WriteByte(c); err != nil {
			return err
		}
		i++
	}
	for j := i; j%4 != 0; j++ {
		if err := msg.argbuf.WriteByte(0); err != nil {
			return err
		}
	}
	return nil
}

// WriteBlob writes a binary blob to an OSC message.
func (msg *Message) WriteBlob(blob []byte) error {
	bl := len(blob)
	if bl == 0 {
		return nil
	}

	// add typetag
	msg.typetag = append(msg.typetag, typetagBlob)

	// write length
	if err := binary.Write(msg.argbuf, byteOrder, int32(bl)); err != nil {
		return err
	}
	// write blob
	if _, err := msg.argbuf.Write(blob); err != nil {
		return err
	}
	// make it 32-bit aligned
	for i := bl; i%4 != 0; i++ {
		if err := msg.argbuf.WriteByte(0); err != nil {
			return err
		}
	}
	return nil
}

// TypeTags returns the message's typetags as a string.
func (msg *Message) TypeTags() string {
	return string(msg.typetag)
}

// Sender returns the address from which a message was sent.
func (msg *Message) Sender() net.Addr {
	return msg.senderAddress
}

// Match returns true, if the address of the OSC Message matches the given address.
// Case sensitive!
func (msg *Message) Match(address string) (bool, error) {
	addr := string(msg.address)

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

// Contents returns the contents of the message
// as a slice of bytes.
func (msg *Message) Contents() ([]byte, error) {
	var (
		w            = &bytes.Buffer{}
		bytesWritten int64
	)

	// Write address
	bw, err := w.Write(append(msg.address, 0))
	if err != nil {
		return nil, err
	}
	bytesWritten += int64(bw)

	// Write padding
	for i := bytesWritten; i%4 != 0; i++ {
		if err := w.WriteByte(0); err != nil {
			return nil, err
		}
		bytesWritten++
	}

	// Write typetag
	bw, err = w.Write(append(msg.typetag, 0))
	if err != nil {
		return nil, err
	}
	bytesWritten += int64(bw)

	// Write padding
	for i := bytesWritten; i%4 != 0; i++ {
		if err := w.WriteByte(0); err != nil {
			return nil, err
		}
		bytesWritten++
	}

	// Write arguments
	args := msg.argbuf.Bytes()
	if _, err := w.Write(args); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// Print pretty prints a Message to an io.Writer.
func (msg *Message) Print(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s%s", msg.address, msg.typetag); err != nil {
		return err
	}

	for _, tt := range msg.typetag[1:] {
		switch tt {
		case typetagInt:
			val, err := msg.ReadInt32()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, " %d", val); err != nil {
				return err
			}
		case typetagFloat:
			val, err := msg.ReadFloat32()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, " %f", val); err != nil {
				return err
			}
		case typetagString:
			val, err := msg.ReadString()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, " %s", val); err != nil {
				return err
			}
			// TODO: handle blobs
		case typetagTrue, typetagFalse:
			val, err := msg.ReadBool()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, " %t", val); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseMessage parses an OSC message from a slice of bytes.
func parseMessage(data []byte, senderAddress net.Addr) (*Message, error) {
	var (
		address []byte
		i       = 0
		n       = len(data)
	)
	for i < n {
		if data[i] == 0 {
			address = data[0:i]
			i++
			break
		}
		i++
	}

	msg := &Message{
		address:       address,
		senderAddress: senderAddress,
		ttReadIndex:   1,
	}

	// advance i to the next multiple of 4
	for i%4 != 0 {
		i++
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
		return ErrInvalidTypeTag
	}

	var (
		i = 0
		n = len(data)
	)

	// read the typetag
	for i < n {
		if data[i] == 0 {
			msg.typetag = data[0:i]
			i++
			break
		}
		i++
	}

	// advance i to the next multiple of 4
	for i%4 != 0 {
		i++
	}

	msg.argbuf = bytes.NewBuffer(data[i:])
	return nil
}

// clone creates a clone of an OSC message.
func (msg *Message) clone() (*Message, error) {
	return &Message{
		address:       msg.address,
		typetag:       msg.typetag,
		argbuf:        bytes.NewBuffer(msg.argbuf.Bytes()),
		ttReadIndex:   1,
		senderAddress: msg.senderAddress,
	}, nil
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
