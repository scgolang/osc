package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Common errors.
var (
	ErrEarlyTimetag = errors.New("enclosing bundle's timetag was later than the nested bundle's")
)

var (
	bundlePrefix    = []byte{'#', 'b', 'u', 'n', 'd', 'l', 'e', 0}
	bundlePrefixLen = len(bundlePrefix)
)

// Bundle is an OSC bundle.
// An OSC Bundle consists of the OSC-string "#bundle" followed by an OSC Time Tag,
// followed by zero or more bundle elements. The OSC-timetag is a 64-bit fixed
// point time tag. See http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag Timetag
	Packets []Packet
}

// NewBundle returns an OSC Bundle.
func NewBundle(t time.Time, packets ...Packet) *Bundle {
	return &Bundle{
		Timetag: FromTime(t),
		Packets: packets,
	}
}

// WriteTo writes the bundle to w.
func (b *Bundle) WriteTo(w io.Writer) (n int64, err error) {
	// Add the '#bundle' string
	nw, err := w.Write(bundlePrefix)
	if err != nil {
		return 0, err
	}
	n += int64(nw)

	for i := n; i%4 != 0; i++ {
		if _, err := w.Write([]byte{0}); err != nil {
			return 0, err
		}
		n++
	}

	// Add the timetag
	if err := binary.Write(w, byteOrder, b.Timetag); err != nil {
		return 0, err
	}
	n += 8

	// Process all OSC Messages
	for _, p := range b.Packets {
		// TODO: would be nice to be able to get the length
		buf := &bytes.Buffer{}
		nw64, err := p.WriteTo(buf)
		if err != nil {
			return 0, err
		}

		size := int32(nw64)
		if err := binary.Write(w, byteOrder, size); err != nil {
			return 0, err
		}

		nw, err := w.Write(buf.Bytes())
		if err != nil {
			return 0, err
		}
		n += int64(nw)
	}

	return n, nil
}

// Invoke invokes an OSC method for each element of a
// bundle recursively.
// If the timetag of the receiver is in the future, then this
// method will block until its time has come.
func (b *Bundle) Invoke(address string, method Method) error {
	until := b.Timetag.Time().Sub(time.Now())

	if until > 0 {
		time.Sleep(until)
	}

	for _, p := range b.Packets {
		if msg, ok := p.(*Message); !ok {
			matched, err := msg.Match(address)
			if err != nil {
				return err
			}
			if matched {
				if err := method(msg); err != nil {
					return err
				}
			}
			continue
		}
		if bundle, ok := p.(*Bundle); ok {
			if bundle.Timetag < b.Timetag {
				return ErrEarlyTimetag
			}
			if err := bundle.Invoke(address, method); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseBundle parses an OSC bundle from a slice of bytes.
func parseBundle(data []byte) (*Bundle, error) {
	var (
		i = 0
		b = &Bundle{}
	)
	if len(data) < len(bundlePrefix) {
		return nil, fmt.Errorf("invalid bundle: %q", data)
	}
	if prefix := data[0:bundlePrefixLen]; 0 != bytes.Compare(prefix, bundlePrefix) {
		return nil, fmt.Errorf("invalid bundle prefix: %q", prefix)
	}
	i = bundlePrefixLen

	timetag, err := parseTimetag(data[i:])
	if err != nil {
		return nil, err
	}
	b.Timetag = timetag
	i += TimetagSize

	var (
		r    = bytes.NewReader(data[i:])
		size int32
	)
	j := 0

ReadPackets:
	for err := binary.Read(r, byteOrder, &size); err == nil; err = binary.Read(r, byteOrder, &size) {
		i += 4
		switch data[i] {
		case messageChar:
			pkt, err := parseMessage(data[i:])
			if err != nil {
				return nil, err
			}
			b.Packets = append(b.Packets, pkt)
		case bundleChar:
			pkt, err := parseBundle(data[i:])
			if err != nil {
				return nil, err
			}
			b.Packets = append(b.Packets, pkt)
		case 0:
			break ReadPackets
		default:
			return nil, ErrParse
		}
		j++
		i += int(size)
		r = bytes.NewReader(data[i:])
	}

	return b, nil
}
