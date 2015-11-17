package osc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

var (
	bundlePrefix = []byte{'#', 'b', 'u', 'n', 'd', 'l', 'e'}
)

// BundleElement is an OSC bundle element.
// A bundle element consists of an int32 that provides the bundle size
// (which will always be a multiple of 4), followed by the contents.
// A bundle element is either an OSC message or an OSC bundle.
type BundleElement struct {
	size     int32  // size will always be a multiple of 4
	contents []byte // contents of bundle element
}

// WriteTo is an implementation of the io.WriterTo interface.
func (el *BundleElement) WriteTo(w io.Writer) (int64, error) {
	var bytesWritten int64
	if err := binary.Write(w, byteOrder, el.size); err != nil {
		return bytesWritten, err
	}

	bytesWritten += 4
	if bw, err := w.Write(el.contents); err != nil {
		return bytesWritten, err
	} else {
		bytesWritten += int64(bw)
	}

	return bytesWritten, nil
}

// Elementer provides the Element method that returns a
// pointer to a BundleElement.
type Elementer interface {
	Element() *BundleElement
}

// An OSC Bundle consists of the OSC-string "#bundle" followed by an OSC Time Tag,
// followed by zero or more bundle elements. The OSC-timetag is a 64-bit fixed
// point time tag. See http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag       Timetag
	Elements      []*BundleElement
	SenderAddress net.Addr
}

// NewBundle returns an OSC Bundle.
func NewBundle() *Bundle {
	return &Bundle{Timetag: NewTimetag(time.Now())}
}

// parseBundle parses an OSC bundle from a slice of bytes.
func parseBundle(data []byte, senderAddress net.Addr) (*Bundle, error) {
	// Read the '#bundle' OSC string
	startTag, _ := readPaddedString(data)
	// *start += n

	if startTag != BundleTag {
		return nil, fmt.Errorf("Invalid bundle start tag: %s", startTag)
	}

	// Read the timetag
	var (
		timeTag uint64
		r       = bytes.NewReader(data)
	)
	if err := binary.Read(r, binary.BigEndian, &timeTag); err != nil {
		return nil, err
	}
	// *start += 8

	// Create a new bundle
	bundle := &Bundle{Timetag: Timetag(timeTag), SenderAddress: senderAddress}

	return bundle, nil
}

// Element is an implementation of the Elementer interface.
func (bun *Bundle) Element() *BundleElement {
	return nil
}

// WriteTo is an implementation of the io.WriterTo interface.
func (bun *Bundle) WriteTo(w io.Writer) (int64, error) {
	var bytesWritten int64

	// Add the '#bundle' string
	if bw, err := w.Write(bundlePrefix); err != nil {
		return bytesWritten, err
	} else {
		bytesWritten += int64(bw)
	}
	for i := bytesWritten; i%4 != 0; i++ {
		if _, err := w.Write([]byte{0}); err != nil {
			return bytesWritten, err
		}
		bytesWritten++
	}

	// Add the timetag
	if err := binary.Write(w, byteOrder, bun.Timetag); err != nil {
		return bytesWritten, err
	}
	bytesWritten += 8

	// Process all OSC Messages
	for _, element := range bun.Elements {
		if bw, err := element.WriteTo(w); err != nil {
			return bytesWritten, err
		} else {
			bytesWritten += int64(bw)
		}
	}

	return bytesWritten, nil
}

// Invoke invokes an OSC method for each element of a
// bundle recursively.
func (bun *Bundle) Invoke(address string, method Method) error {
	// TODO: implement
	return nil
}
