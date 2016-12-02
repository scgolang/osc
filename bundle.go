package osc

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/pkg/errors"
)

const (
	// BundleTag is the tag on an OSC bundle message.
	BundleTag = "#bundle"
)

// Common errors.
var (
	ErrEarlyTimetag = errors.New("enclosing bundle's timetag was later than the nested bundle's")
	ErrEndOfPackets = errors.New("end of packets")
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
	return parseBundle(data, sender, -1)
}

// parseBundle parses a bundle from a byte slice.
// It will stop after reading limit bytes.
// If you wish to have it consume as many bytes as possible, pass -1 as the limit.
func parseBundle(data []byte, sender net.Addr, limit int32) (Bundle, error) {
	b := Bundle{}

	// If 0 <= limit < 16 this is an error.
	// We have to be able to read at least the bundle tag and a timetag.
	if (limit >= 0) && (limit < int32(len(BundleTag)+1+TimetagSize)) {
		return b, errors.New("limit must be >= 16 or < 0")
	}

	data, err := sliceBundleTag(data)
	if err != nil {
		return b, errors.Wrap(err, "slice bundle tag")
	}

	tt, err := ReadTimetag(data)
	if err != nil {
		return b, errors.Wrap(err, "read timetag")
	}
	b.Timetag = tt
	data = data[8:]

	// We are limited to only reading the bundle tag and the timetag.
	if limit == 16 {
		return b, nil
	}

	// We take away 16 from limit so that readPackets doesn't have to know we have already read 16 bytes.
	packets, err := readPackets(data, sender, limit-16)
	if err != nil {
		return b, errors.Wrap(err, "read packets")
	}
	b.Packets = packets

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

// sliceBundleTag slices the bundle tag off the data.
// If the bundle tag is not present or is not correct, an error is returned.
func sliceBundleTag(data []byte) ([]byte, error) {
	bundleTag := append([]byte(BundleTag), 0)
	if len(data) < len(bundleTag) {
		return nil, errors.Errorf("expected %q, got %q", bundleTag, data)
	}
	idx := bytes.Index(data, bundleTag)
	if idx == -1 {
		return nil, errors.Errorf("expected %q, got %q", bundleTag, data[:len(bundleTag)])
	}
	return data[len(bundleTag):], nil
}

// readPackets reads bundle packets from a byte slice.
func readPackets(data []byte, sender net.Addr, limit int32) ([]Packet, error) {
	ps := []Packet{}

	var (
		p   Packet
		l   int32
		err error
	)
	for {
		p, l, err = readPacket(data, sender)
		if err == ErrEndOfPackets {
			return ps, nil
		}
		if err != nil {
			return nil, errors.Wrap(err, "read packet")
		}
		ps = append(ps, p)
		if l+4 == int32(len(data)) {
			break
		}
		if (limit >= 0) && (limit <= l+4) {
			break
		}
		data = data[l+4:]
	}
	return ps, nil
}

// readPacket reads an OSC bundle packet from a byte slice.
// The packet and the packet length are returned along with nil if there was no error.
// If the packet length is 0 then ErrEndOfPackets is returned as the error.
// If ErrEndOfPackets is returned then Packet will always be nil.
// The returned packet length includes the length of the packet length integer itself,
// so it is actually packet_length + 4.
func readPacket(data []byte, sender net.Addr) (Packet, int32, error) {
	if len(data) < 4 {
		return nil, int32(len(data)), ErrEndOfPackets
	}
	var l int32
	_ = binary.Read(bytes.NewReader(data), byteOrder, &l) // Never fails
	if l == int32(0) {
		return nil, 0, ErrEndOfPackets
	}

	data = data[4:]

	if int32(len(data)) < l {
		return nil, 0, errors.Errorf("packet length %d is greater than data length %d", l, len(data))
	}

	switch data[0] {
	case MessageChar:
		msg, err := ParseMessage(data, sender)
		if err != nil {
			return nil, 0, errors.Wrap(err, "parse message from packet")
		}
		return msg, l, nil // The returned length includes the packet length integer.
	case BundleTag[0]:
		bundle, err := parseBundle(data, sender, l)
		if err != nil {
			return nil, 0, errors.Wrap(err, "parse bundle from packet")
		}
		return bundle, l, nil // The returned length includes the packet length integer.
	default:
		return nil, 0, errors.Errorf("packet should never start with %c", data[0])
	}
}
