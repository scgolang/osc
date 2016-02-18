package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	// TimetagSize is the number of 8-bit bytes
	// in an OSC timetag.
	TimetagSize = 8

	// Immediately is a special timetag value that
	// means "immediately".
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

// Time converts an OSC timetag to a time.Time.
func (tt Timetag) Time() time.Time {
	secs := (int64(tt) >> 32) - secondsFrom1900To1970
	return time.Unix(secs, int64(tt)&0xFFFFFFFF)
}

// FromTime converts the given time to an OSC timetag.
func FromTime(t time.Time) Timetag {
	secs := uint64((secondsFrom1900To1970 + t.Unix()) << 32)
	return Timetag(secs + uint64(uint32(t.Nanosecond())))
}

// parseTimetag parses a timetag from a byte slice.
func parseTimetag(data []byte) (Timetag, error) {
	if len(data) < 8 {
		return Timetag(0), fmt.Errorf("timetags must be 64-bit")
	}
	var (
		buf1  = bytes.NewBuffer(data[:TimetagSize/2])
		buf2  = bytes.NewBuffer(data[TimetagSize/2:])
		secs  uint64
		nsecs uint64
	)
	if err := binary.Read(buf1, byteOrder, &secs); err != nil {
		return Timetag(0), nil
	}
	if err := binary.Read(buf2, byteOrder, &nsecs); err != nil {
		return Timetag(0), nil
	}
	return Timetag((secs << 32) + nsecs), nil
}
