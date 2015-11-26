// osc provides a package for sending and receiving OpenSoundControl messages.
package osc

import "encoding/binary"

const (
	// The time tag value consisting of 63 zero bits followed by a one in the
	// least signifigant bit is a special case meaning "immediately."
	timeTagImmediate      = uint64(1)
	secondsFrom1900To1970 = 2208988800
	BundleTag             = "#bundle"
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
