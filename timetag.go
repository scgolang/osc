package osc

import (
	"bytes"
	"encoding/binary"
	"time"
)

// Timetag represents an OSC Time Tag.
// An OSC Time Tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
type Timetag uint64

// NewTimetag returns a new OSC timetag object.
func NewTimetag(timeStamp time.Time) Timetag {
	return Timetag(timeToTimetag(timeStamp))
}

// FractionalSecond returns the last 32 bits of the Osc Time Tag. Specifies the
// fractional part of a second.
func (self Timetag) FractionalSecond() uint32 {
	return uint32(uint64(self) << 32)
}

// SecondsSinceEpoch returns the first 32 bits (the number of seconds since the
// midnight 1900) from the OSC timetag.
func (self Timetag) SecondsSinceEpoch() uint32 {
	return uint32(uint64(self) >> 32)
}

// ToByteArray converts the OSC Time Tag to a byte array.
func (self Timetag) ToByteArray() []byte {
	var data = new(bytes.Buffer)
	binary.Write(data, binary.BigEndian, uint64(self))
	return data.Bytes()
}

// ExpiresIn calculates the number of seconds until the current time is the
// same as the value of the timetag. It returns zero if the value of the
// timetag is in the past.
func (self Timetag) ExpiresIn() time.Duration {
	if int(self) <= 1 {
		return 0
	}

	tt := timetagToTime(uint64(self))
	seconds := tt.Sub(time.Now())

	if seconds <= 0 {
		return 0
	}

	return seconds
}
