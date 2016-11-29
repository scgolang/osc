package osc

import (
	"bytes"
	"errors"
	"net"
)

const (
	// BundleTag is the tag on an OSC bundle message.
	BundleTag = "#bundle"
)

// Common errors.
var (
	ErrEarlyTimetag = errors.New("enclosing bundle's timetag was later than the nested bundle's")
)

// Bundle is an OSC bundle.
// An OSC Bundle consists of the OSC-string "#bundle" followed by an OSC Time Tag,
// followed by zero or more bundle elements. The OSC-timetag is a 64-bit fixed
// point time tag. See http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag Timetag
	Packets []Packet
	Sender  net.Addr
}

// ParseBundle parses a bundle from a byte slice.
func ParseBundle(data []byte, sender net.Addr) (Bundle, error) {
	b := Bundle{}
	return b, nil
}

// Bytes returns the contents of the bundle as a slice of bytes.
func (b Bundle) Bytes() []byte {
	bss := [][]byte{
		ToBytes(BundleTag),
		b.Timetag.Bytes(),
	}
	for _, p := range b.Packets {
		var (
			bs     = p.Bytes()
			length = Int(int32(len(bs)))
		)
		bss = append(bss, length.Bytes(), p.Bytes())
	}
	return bytes.Join(bss, []byte{})
}

// Equal returns true if one bundle equals another, and false otherwise.
func (b Bundle) Equal(other Packet) bool {
	b2, ok := other.(Bundle)
	if !ok {
		return false
	}
	if b.Timetag != b2.Timetag {
		return false
	}
	if len(b.Packets) != len(b2.Packets) {
		return false
	}
	for i, p := range b.Packets {
		if !p.Equal(b2.Packets[i]) {
			return false
		}
	}
	return true
}
