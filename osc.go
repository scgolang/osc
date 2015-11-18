// osc provides a package for sending and receiving OpenSoundControl messages.
// The package is implemented in pure Go.
package osc

import (
	"encoding/binary"
	"regexp"
	"strings"
)

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

// getRegEx compiles and returns a regular expression object for the given address
// pattern.
func getRegEx(pattern string) (*regexp.Regexp, error) {
	pattern = strings.Replace(pattern, ".", "\\.", -1) // Escape all '.' in the pattern
	pattern = strings.Replace(pattern, "(", "\\(", -1) // Escape all '(' in the pattern
	pattern = strings.Replace(pattern, ")", "\\)", -1) // Escape all ')' in the pattern
	pattern = strings.Replace(pattern, "*", ".*", -1)  // Replace a '*' with '.*' that matches zero or more characters
	pattern = strings.Replace(pattern, "{", "(", -1)   // Change a '{' to '('
	pattern = strings.Replace(pattern, ",", "|", -1)   // Change a ',' to '|'
	pattern = strings.Replace(pattern, "}", ")", -1)   // Change a '}' to ')'
	pattern = strings.Replace(pattern, "?", ".", -1)   // Change a '?' to '.'
	return regexp.Compile(pattern)
}
