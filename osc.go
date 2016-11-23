package osc

import (
	"bytes"
	"encoding/binary"
)

const (
	// TimeTagImmediate represents the time tag value consisting of
	// 63 zero bits followed by a one in the least significant bit.
	TimeTagImmediate = uint64(1)

	// SecondsFrom1900To1970 is exactly what it sounds like.
	SecondsFrom1900To1970 = 2208988800

	// BundleTag is the tag on an OSC bundle message.
	BundleTag = "#bundle"

	// MessageChar is the first character of any valid OSC message.
	MessageChar = '/'
)

// Typetag constants.
const (
	TypetagPrefix byte = ','
	TypetagInt    byte = 'i'
	TypetagFloat  byte = 'f'
	TypetagString byte = 's'
	TypetagBlob   byte = 'b'
	TypetagFalse  byte = 'F'
	TypetagTrue   byte = 'T'
)

var (
	byteOrder = binary.BigEndian
)

// Packet is an OSC packet.
// An OSC packet consists of its contents, a contiguous block
// of binary data, and its size, the number of 8-bit bytes
// that comprise the contents. The size of an OSC packet
// is always a multiple of 4.
type Packet interface {
	Contents() ([]byte, error)
}

// OscString returns an OSC representation of the given string.
// This means that the returned byte slice is padded with null bytes
// so that it's length is a multiple of 4.
func OscString(s string) []byte {
	if len(s) == 0 {
		return []byte{}
	}
	return Pad(append([]byte(s), 0))
}

// Pad pads a slice of bytes with null bytes so that it's length is a multiple of 4.
func Pad(b []byte) []byte {
	for i := len(b); (i % 4) != 0; i++ {
		b = append(b, 0)
	}
	return b
}

// ReadString reads a string from a byte slice.
// If the byte slice does not have any null bytes,
// then one is appended to the end.
// If the length of the byte slice is not a multiple of 4
// we append as many null bytes as we need to make this true
// before converting to a string.
// What this means is that the second return value, which is
// the number of bytes that are consumed to create the string is
// always a multiple of 4.
// We also strip off any trailing null bytes in the returned string.
func ReadString(data []byte) (string, int64) {
	if len(data) == 0 {
		return "", 0
	}
	nullidx := bytes.IndexByte(data, 0)
	if nullidx == -1 {
		// This should never happen!
		data = append(data, 0)
		nullidx = len(data)
	}
	b, bl := ReadBlob(int32(nullidx), data)
	nullidx = bytes.IndexByte(b, 0)
	return string(b[:nullidx]), bl
}

// ReadBlob reads a blob of the given length from the given slice of bytes.
func ReadBlob(length int32, data []byte) ([]byte, int64) {
	l := length
	if length > int32(len(data)) {
		l = int32(len(data))
	}
	var idx int32
	for idx = l; (idx % 4) != 0; idx++ {
		if idx >= int32(len(data)) {
			data = append(data, 0)
		}
	}
	return data[:idx], int64(idx)
}
