package osc

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/pkg/errors"
)

const (
	// SecondsFrom1900To1970 is exactly what it sounds like.
	SecondsFrom1900To1970 = 2208988800
	// SecondsFrom1900To1970 = 2207520000

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
	secs := (uint64(tt) >> 32) - SecondsFrom1900To1970
	return time.Unix(int64(secs), int64(tt)&0xFFFFFFFF).UTC()
}

// FromTime converts the given time to an OSC timetag.
func FromTime(t time.Time) Timetag {
	t = t.UTC()
	secs := uint64((SecondsFrom1900To1970 + t.Unix()) << 32)
	return Timetag(secs + uint64(uint32(t.Nanosecond())))
}

// ReadTimetag parses a timetag from a byte slice.
func ReadTimetag(data []byte) (Timetag, error) {
	if len(data) < TimetagSize {
		return Timetag(0), errors.New("timetags must be 64-bit")
	}
	zero := []byte{0, 0, 0, 0}
	var (
		L     = append(zero, data[:TimetagSize/2]...)
		R     = append(zero, data[TimetagSize/2:]...)
		secs  uint64
		nsecs uint64
	)
	_ = binary.Read(bytes.NewReader(L), byteOrder, &secs)  // Never fails
	_ = binary.Read(bytes.NewReader(R), byteOrder, &nsecs) // Never fails
	return Timetag((secs << 32) + nsecs), nil
}
