package osc

import (
	"bufio"
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
type Timetag struct {
	timeTag  uint64 // The acutal time tag
	time     time.Time
	MinValue uint64 // Minimum value of an OSC Time Tag. Is always 1.
}

// NewTimetag returns a new OSC timetag object.
func NewTimetag(timeStamp time.Time) (timetag *Timetag) {
	return &Timetag{
		time:     timeStamp,
		timeTag:  timeToTimetag(timeStamp),
		MinValue: uint64(1)}
}

// NewTimetagFromTimetag creates a new Timetag from the given time tag.
func NewTimetagFromTimetag(timetag uint64) (t *Timetag) {
	time := timetagToTime(timetag)
	return NewTimetag(time)
}

// Time returns the time.
func (self *Timetag) Time() time.Time {
	return self.time
}

// FractionalSecond returns the last 32 bits of the Osc Time Tag. Specifies the
// fractional part of a second.
func (self *Timetag) FractionalSecond() uint32 {
	return uint32(self.timeTag << 32)
}

// SecondsSinceEpoch returns the first 32 bits (the number of seconds since the
// midnight 1900) from the OSC timetag.
func (self *Timetag) SecondsSinceEpoch() uint32 {
	return uint32(self.timeTag >> 32)
}

// TimeTag returns the time tag value
func (self *Timetag) TimeTag() uint64 {
	return self.timeTag
}

// ToByteArray converts the OSC Time Tag to a byte array.
func (self *Timetag) ToByteArray() []byte {
	var data = new(bytes.Buffer)
	binary.Write(data, binary.BigEndian, self.timeTag)
	return data.Bytes()
}

// SetTime sets the value of the OSC Time Tag.
func (self *Timetag) SetTime(time time.Time) {
	self.time = time
	self.timeTag = timeToTimetag(time)
}

// ExpiresIn calculates the number of seconds until the current time is the
// same as the value of the timetag. It returns zero if the value of the
// timetag is in the past.
func (self *Timetag) ExpiresIn() time.Duration {
	if self.timeTag <= 1 {
		return 0
	}

	tt := timetagToTime(self.timeTag)
	seconds := tt.Sub(time.Now())

	if seconds <= 0 {
		return 0
	}

	return seconds
}

////
// De/Encoding functions
////

// readBlob reads an OSC Blob from the blob byte array. Padding bytes are removed
// from the reader and not returned.
func readBlob(reader *bufio.Reader) (blob []byte, n int, err error) {
	// First, get the length
	var blobLen int
	if err = binary.Read(reader, binary.BigEndian, &blobLen); err != nil {
		return nil, 0, err
	}
	n = 4 + blobLen

	// Read the data
	blob = make([]byte, blobLen)
	if _, err = reader.Read(blob); err != nil {
		return nil, 0, err
	}

	// Remove the padding bytes
	numPadBytes := padBytesNeeded(blobLen)
	if numPadBytes > 0 {
		n += numPadBytes
		dummy := make([]byte, numPadBytes)
		if _, err = reader.Read(dummy); err != nil {
			return nil, 0, err
		}
	}

	return blob, n, nil
}
