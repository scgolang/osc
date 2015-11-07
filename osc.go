// osc provides a package for sending and receiving OpenSoundControl messages.
// The package is implemented in pure Go.
package osc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	// The time tag value consisting of 63 zero bits followed by a one in the
	// least signifigant bit is a special case meaning "immediately."
	timeTagImmediate      = uint64(1)
	secondsFrom1900To1970 = 2208988800
	BundleTag             = "#bundle"
)

// OSC message handler interface. Every handler function for an OSC message must
// implement this interface.
type Handler interface {
	HandleMessage(msg *Message)
}

// Type defintion for an OSC handler function
type HandlerFunc func(msg *Message)

// HandleMessage calls themeself with the given OSC Message. Implements the
// Handler interface.
func (f HandlerFunc) HandleMessage(msg *Message) {
	f(msg)
}

// writeBlob writes the data byte array as an OSC blob into buff. If the length of
// data isn't 32-bit aligned, padding bytes will be added.
func writeBlob(data []byte, buff *bytes.Buffer) (numberOfBytes int, err error) {
	// Add the size of the blob
	dlen := int32(len(data))
	err = binary.Write(buff, binary.BigEndian, dlen)
	if err != nil {
		return 0, err
	}

	// Write the data
	if _, err = buff.Write(data); err != nil {
		return 0, nil
	}

	// Add padding bytes if necessary
	numPadBytes := padBytesNeeded(len(data))
	if numPadBytes > 0 {
		padBytes := make([]byte, numPadBytes)
		if numPadBytes, err = buff.Write(padBytes); err != nil {
			return 0, err
		}
	}

	return 4 + len(data) + numPadBytes, nil
}

// readPaddedString reads a padded string from the given reader. The padding bytes
// are removed from the reader.
func readPaddedString(reader *bufio.Reader) (str string, n int, err error) {
	// Read the string from the reader
	str, err = reader.ReadString(0)
	if err != nil {
		return "", 0, err
	}
	n = len(str)

	// Remove the string delimiter, in order to calculate the right amount
	// of padding bytes
	str = str[:len(str)-1]

	// Remove the padding bytes
	padLen := padBytesNeeded(len(str)) - 1
	if padLen > 0 {
		n += padLen
		padBytes := make([]byte, padLen)
		if _, err = reader.Read(padBytes); err != nil {
			return "", 0, err
		}
	}

	return str, n, nil
}

// writePaddedString writes a string with padding bytes to the a buffer.
// Returns, the number of written bytes and an error if any.
func writePaddedString(str string, buff *bytes.Buffer) (numberOfBytes int, err error) {
	// Write the string to the buffer
	n, err := buff.WriteString(str)
	if err != nil {
		return 0, err
	}

	// Calculate the padding bytes needed and create a buffer for the padding bytes
	numPadBytes := padBytesNeeded(len(str))
	if numPadBytes > 0 {
		padBytes := make([]byte, numPadBytes)
		// Add the padding bytes to the buffer
		if numPadBytes, err = buff.Write(padBytes); err != nil {
			return 0, err
		}
	}

	return n + numPadBytes, nil
}

// padBytesNeeded determines how many bytes are needed to fill up to the next 4
// byte length.
func padBytesNeeded(elementLen int) int {
	return 4*(elementLen/4+1) - elementLen
}

////
// Timetag utility functions
////

// timeToTimetag converts the given time to an OSC timetag.
//
// An OSC timetage is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
//
// The time tag value consisting of 63 zero bits followed by a one in the least
// signifigant bit is a special case meaning "immediately."
func timeToTimetag(time time.Time) (timetag uint64) {
	timetag = uint64((secondsFrom1900To1970 + time.Unix()) << 32)
	return (timetag + uint64(uint32(time.Nanosecond())))
}

// timetagToTime converts the given timetag to a time object.
func timetagToTime(timetag uint64) (t time.Time) {
	return time.Unix(int64((timetag>>32)-secondsFrom1900To1970), int64(timetag&0xffffffff))
}

// existsAddress returns true if the address s is found in handlers. Otherwise, false.
func existsAddress(s string, handlers map[string]Handler) bool {
	for address, _ := range handlers {
		if address == s {
			return true
		}
	}

	return false
}

// getRegEx compiles and returns a regular expression object for the given address
// pattern.
func getRegEx(pattern string) *regexp.Regexp {
	pattern = strings.Replace(pattern, ".", "\\.", -1) // Escape all '.' in the pattern
	pattern = strings.Replace(pattern, "(", "\\(", -1) // Escape all '(' in the pattern
	pattern = strings.Replace(pattern, ")", "\\)", -1) // Escape all ')' in the pattern
	pattern = strings.Replace(pattern, "*", ".*", -1)  // Replace a '*' with '.*' that matches zero or more characters
	pattern = strings.Replace(pattern, "{", "(", -1)   // Change a '{' to '('
	pattern = strings.Replace(pattern, ",", "|", -1)   // Change a ',' to '|'
	pattern = strings.Replace(pattern, "}", ")", -1)   // Change a '}' to ')'
	pattern = strings.Replace(pattern, "?", ".", -1)   // Change a '?' to '.'

	return regexp.MustCompile(pattern)
}

// getTypeTag returns the OSC type tag for the given argument.
func getTypeTag(arg interface{}) (s string, err error) {
	switch t := arg.(type) {
	default:
		return "", errors.New(fmt.Sprintf("Unsupported type: %T", t))

	case bool:
		if arg.(bool) {
			s = "T"
		} else {
			s = "F"
		}

	case nil:
		s = "N"

	case int32:
		s = "i"

	case float32:
		s = "f"

	case string:
		s = "s"

	case []byte:
		s = "b"

	case int64:
		s = "h"

	case float64:
		s = "d"

	case Timetag:
		s = "t"
	}
	return s, nil
}
