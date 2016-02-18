package osc

import "encoding/binary"

const (
	// The time tag value consisting of 63 zero bits followed by a one in the
	// least signifigant bit is a special case meaning "immediately."
	timeTagImmediate      = uint64(1)
	secondsFrom1900To1970 = 2208988800
	bundleTag             = "#bundle"
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
	Contents() ([]byte, error)
}
