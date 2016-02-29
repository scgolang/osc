package osc

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	// The time tag value consisting of 63 zero bits followed by a one in the
	// least significant bit is a special case meaning "immediately."
	secondsFrom1900To1970 = 2208988800
	messageChar           = '/'
	bundleChar            = '#'
	typetagPrefix         = ','
	typetagInt            = 'i'
	typetagFloat          = 'f'
	typetagString         = 's'
	typetagBlob           = 'b'
	typetagFalse          = 'F'
	typetagTrue           = 'T'
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
	io.WriterTo
}

// padString returns the padded bytes for an OSC string.
// It adds a null byte, then adds 0 to 3 more null bytes so that
// the length of the resulting slice is a multiple of 4.
func padString(s string) []byte {
	b := append([]byte(s), 0)
	for i := len(b); i%4 != 0; i++ {
		b = append(b, 0)
	}
	return b
}

// padBytes pads a byte slice to make it a valid OSC string.
// It adds a null byte, then adds 0 to 3 more null bytes so that
// the length of the resulting slice is a multiple of 4.
func padBytes(b []byte) []byte {
	bb := append(b, 0)
	for i := len(bb); i%4 != 0; i++ {
		bb = append(bb, 0)
	}
	return bb
}

// toBytes converts a float32 to a byte slice
func toBytes(x interface{}) []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, byteOrder, x) // Can't fail
	return buf.Bytes()
}

// paddedLength returns the padded length for a string or blob.
func paddedLength(l int) int {
	for i := l + 1; i%4 != 0; i++ {
		l++
	}
	return l
}
