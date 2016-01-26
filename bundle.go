package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
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
	Sender  net.Addr
}

// NewBundle returns an OSC Bundle.
func NewBundle(t time.Time, packets ...Packet) *Bundle {
	return &Bundle{
		Timetag: FromTime(t),
		Packets: packets,
	}
}

// Contents returns the contents of the bundle as a
// slice of bytes.
func (b *Bundle) Contents() ([]byte, error) {
	var (
		buf          = &bytes.Buffer{}
		bytesWritten int64
	)

	// Add the '#bundle' string
	if bw, err := buf.Write(bundlePrefix); err != nil {
		return nil, err
	} else {
		bytesWritten += int64(bw)
	}
	for i := bytesWritten; i%4 != 0; i++ {
		if err := buf.WriteByte(0); err != nil {
			return nil, err
		}
		bytesWritten++
	}

	// Add the timetag
	if err := binary.Write(buf, byteOrder, b.Timetag); err != nil {
		return nil, err
	}
	bytesWritten += 8

	// Process all OSC Messages
	for _, p := range b.Packets {
		contents, err := p.Contents()
		if err != nil {
			return nil, err
		}

		size := int32(len(contents))
		if err := binary.Write(buf, byteOrder, size); err != nil {
			return nil, err
		}

		if bw, err := buf.Write(contents); err != nil {
			return nil, err
		} else {
			bytesWritten += int64(bw)
		}
	}

	return buf.Bytes(), nil
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
		if msg, ok := p.(*Message); ok {
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
func parseBundle(data []byte, sender net.Addr) (*Bundle, error) {
	var (
		i = 0
		b = &Bundle{Sender: sender}
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
			pkt, err := parseMessage(data[i:], sender)
			if err != nil {
				return nil, err
			}
			b.Packets = append(b.Packets, pkt)
		case bundleChar:
			pkt, err := parseBundle(data[i:], sender)
			if err != nil {
				return nil, err
			}
			b.Packets = append(b.Packets, pkt)
		case 0:
			break ReadPackets
		default:
			return nil, ErrParse
		}
		j += 1
		i += int(size)
		r = bytes.NewReader(data[i:])
	}

	return b, nil
}
