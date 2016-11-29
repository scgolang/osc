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
	b := Bundle{}

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

	packets, err := readPackets(data, sender)
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
	idx := bytes.Index(data, []byte(bundleTag))
	if idx == -1 {
		return nil, errors.Errorf("expected %q, got %q", []byte(bundleTag), data[:len(bundleTag)])
	}
	return data[len(bundleTag):], nil
}

// readPackets reads bundle packets from a byte slice.
func readPackets(data []byte, sender net.Addr) ([]Packet, error) {
	ps := []Packet{}

	p, l, err := readPacket(data, sender)
	if err != nil {
		if err == ErrEndOfPackets {
			return ps, nil
		}
		return nil, err
	}
	ps = append(ps, p)

	for data = data[l:]; err != nil; data = data[l:] {
		ps = append(ps, p)
		p, l, err = readPacket(data, sender)
		if err == ErrEndOfPackets {
			return ps, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return ps, nil
}

// readPacket reads an OSC bundle packet from a byte slice.
// The packet and the packet length are returned along with nil if there was no error.
// If the packet length is 0 then ErrEndOfPackets is returned as the error.
func readPacket(data []byte, sender net.Addr) (Packet, int32, error) {
	if len(data) < 4 {
		return nil, int32(len(data)), ErrEndOfPackets
	}
	var l int32
	if err := binary.Read(bytes.NewReader(data), byteOrder, &l); err != nil {
		return nil, 0, errors.Wrap(err, "read packet length")
	}
	if l == int32(0) {
		return nil, 0, ErrEndOfPackets
	}

	data = data[4:]

	if int32(len(data)) < l {
		return nil, 0, errors.New("packet length is greater than data length")
	}

	switch data[0] {
	case MessageChar:
		msg, err := ParseMessage(data, sender)
		if err != nil {
			return nil, 0, errors.Wrap(err, "parse message from packet")
		}
		return msg, l, nil
	case BundleTag[0]:
		bundle, err := ParseBundle(data, sender)
		if err != nil {
			return nil, 0, errors.Wrap(err, "parse bundle from packet")
		}
		return bundle, l, nil
	default:
		return nil, 0, errors.Errorf("packet should never start with %c", data[0])
	}
}
