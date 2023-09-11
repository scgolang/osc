package osc

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/pkg/errors"
)

const (
	// SecondsFrom1900To1970 is exactly what it sounds like.
	SecondsFrom1900To1970 = 2208988800 // Source: RFC 868

	nanosecondsPerFraction = float64(0.23283064365386962891) // 1e9/(2^32)

	// TimetagSize is the number of 8-bit bytes in an OSC timetag.
	TimetagSize = 8

	// Immediately is a special timetag value that means "immediately".
	Immediately = Timetag(1)
)

// Timetag represents an OSC Time Tag.
// An OSC Time Tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
// The time tag value consisting of 63 zero bits followed by a one in the least
// significant bit is a special case meaning "immediately."
type Timetag uint64

// Bytes converts the timetag to a slice of bytes.
func (tt Timetag) Bytes() []byte {
	bs := make([]byte, 8)
	byteOrder.PutUint64(bs, uint64(tt))
	return bs
}

func (tt Timetag) String() string {
	return tt.Time().Format(time.RFC3339)
}

// Time converts an OSC timetag to a time.Time.
func (tt Timetag) Time() time.Time {
	t := uint64(tt)

	if t == 1 {
		// Means "immediately". It cannot occur otherwise as timetag == 0 gets
		// converted to January 1, 1900 while time.Time{} means year 1 in Go.
		// Use the time.Time.IsZero() method to detect it.
		return time.Time{}
	}

	return time.Unix(
		int64(t>>32)-SecondsFrom1900To1970,
		int64(nanosecondsPerFraction*float64(t&(1<<32-1))),
	).UTC()
}

// FromTime converts the given time to an OSC timetag.
func FromTime(t time.Time) Timetag {
	if t.IsZero() {
		return 1
	}

	seconds := uint64(t.Unix() + SecondsFrom1900To1970)
	secondFraction := float64(t.UTC().Nanosecond()) / nanosecondsPerFraction
	return Timetag((seconds << 32) + uint64(uint32(secondFraction)))
}

// ReadTimetag parses a timetag from a byte slice.
func ReadTimetag(data []byte) (Timetag, error) {
	if len(data) < TimetagSize {
		return Timetag(0), errors.New("timetags must be 64-bit")
	}
	var tt uint64
	_ = binary.Read(bytes.NewReader(data), binary.BigEndian, &tt)
	return Timetag(tt), nil
}
