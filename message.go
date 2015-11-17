package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

var (
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrInvalidTypeTag   = errors.New("invalid type tag")
	ErrNilWriter        = errors.New("writer must not be nil")
	ErrParse            = errors.New("error parsing message")
)

// Represents a single OSC message. An OSC message consists of an OSC address
// pattern and zero or more arguments.
type Message struct {
	address       []byte
	typetag       []byte
	argbuf        *bytes.Buffer
	ttReadIndex   int
	senderAddress net.Addr
}

// NewMessage creates a new OSC message.
func NewMessage(addr string) *Message {
	return &Message{
		address:     []byte(addr),
		argbuf:      &bytes.Buffer{},
		typetag:     []byte{typetagPrefix},
		ttReadIndex: 1, // skip the leading ','
	}
}

// parseMessage parses an OSC message from a slice of bytes.
func parseMessage(data []byte, senderAddress net.Addr) (*Message, error) {
	var (
		address []byte
		i       = 0
		n       = len(data)
	)
	for i < n {
		if data[i] == ',' {
			address = data[0:i]
			break
		}
		i++
	}

	msg := &Message{
		address:       address,
		senderAddress: senderAddress,
		ttReadIndex:   1,
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

// ReadFloat reads a float32 value from an OSC message.
func (msg *Message) ReadFloat() (float32, error) {
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
	for _, c := range []byte(val) {
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

// TypeTags returns the message's typetags as a string.
func (msg *Message) TypeTags() string {
	return string(msg.typetag)
}

// Sender returns the address from which a message was sent.
func (msg *Message) Sender() net.Addr {
	return msg.senderAddress
}

// Returns true, if the address of the OSC Message matches the given address.
// Case sensitive!
func (msg *Message) Match(address string) (bool, error) {
	exp, err := getRegEx(string(msg.address))
	if err != nil {
		return false, err
	}

	if exp.MatchString(address) {
		return true, nil
	}

	return false, nil
}

// Bytes returns the message as a slice of bytes.
func (msg *Message) Bytes() ([]byte, error) {
	var (
		w            = &bytes.Buffer{}
		bytesWritten int64
	)

	// Write address
	if bw, err := w.Write(msg.address); err != nil {
		return nil, err
	} else {
		bytesWritten += int64(bw)
	}

	// Write typetag
	if bw, err := w.Write(msg.typetag); err != nil {
		return nil, err
	} else {
		bytesWritten += int64(bw)
	}

	// Write padding
	for i := bytesWritten; i%4 != 0; i++ {
		if err := w.WriteByte(0); err != nil {
			return nil, err
		}
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
			val, err := msg.ReadFloat()
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

// readBlob reads an OSC Blob from the blob byte array.
// Padding bytes are removed from the reader and not returned.
func readBlob(r io.Reader) (blob []byte, n int, err error) {
	// First, get the length
	var blobLen int
	if err = binary.Read(r, binary.BigEndian, &blobLen); err != nil {
		return nil, 0, err
	}
	n = 4 + blobLen

	// Read the data
	blob = make([]byte, blobLen)
	if _, err = r.Read(blob); err != nil {
		return nil, 0, err
	}

	// Remove the padding bytes
	numPadBytes := padBytesNeeded(blobLen)
	if numPadBytes > 0 {
		n += numPadBytes
		dummy := make([]byte, numPadBytes)
		if _, err = r.Read(dummy); err != nil {
			return nil, 0, err
		}
	}

	return blob, n, nil
}
