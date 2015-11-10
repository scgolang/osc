package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// An OSC Bundle consists of the OSC-string "#bundle" followed by an OSC Time Tag,
// followed by zero or more OSC bundle/message elements. The OSC-timetag is a 64-bit fixed
// point time tag. See http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag       Timetag
	Messages      []*Message
	Bundles       []*Bundle
	senderAddress net.Addr
}

// NewBundle returns an OSC Bundle. Use this function to create a new OSC Bundle.
func NewBundle(time time.Time) (bundle *Bundle) {
	return &Bundle{Timetag: NewTimetag(time)}
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
	bundle := &Bundle{Timetag: Timetag(timeTag), senderAddress: senderAddress}

	// Read until the end of the buffer
	// for *start < end {
	// 	// Read the size of the bundle element
	// 	var length int32
	// 	err = binary.Read(r, binary.BigEndian, &length)
	// 	*start += 4
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	var packet Packet
	// 	packet, err = self.readPacket(r, start, end)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	bundle.Append(packet)
	// }

	return bundle, nil
}

// Append appends an OSC bundle or OSC message to the bundle.
func (self *Bundle) Append(pkt Packet) (err error) {
	switch t := pkt.(type) {
	default:
		return errors.New(fmt.Sprintf("Unsupported OSC packet type: only Bundle and Message are supported.", t))

	case *Bundle:
		self.Bundles = append(self.Bundles, t)

	case *Message:
		self.Messages = append(self.Messages, t)
	}

	return nil
}

// ToByteArray serializes the OSC bundle to a byte array with the following format:
// 1. Bundle string: '#bundle'
// 2. OSC timetag
// 3. Length of first OSC bundle element
// 4. First bundle element
// 5. Length of n OSC bundle element
// 6. n bundle element
func (self *Bundle) ToByteArray() (buffer []byte, err error) {
	var data = &bytes.Buffer{}

	// Add the '#bundle' string
	_, err = writePaddedString("#bundle", data)
	if err != nil {
		return nil, err
	}

	// Add the timetag
	if _, err = data.Write(self.Timetag.ToByteArray()); err != nil {
		return nil, err
	}

	// Process all OSC Messages
	for _, m := range self.Messages {
		var msgLen int
		var msgBuf []byte

		msgBuf, err = m.ToByteArray()
		if err != nil {
			return nil, err
		}

		// Append the length of the OSC message
		msgLen = len(msgBuf)
		if err = binary.Write(data, binary.BigEndian, int32(msgLen)); err != nil {
			return nil, err
		}

		// Append the OSC message
		data.Write(msgBuf)
	}

	// Process all OSC Bundles
	for _, b := range self.Bundles {
		var bLen int
		var bBuf []byte

		bBuf, err = b.ToByteArray()
		if err != nil {
			return nil, err
		}

		// Write the size of the bundle
		bLen = len(bBuf)
		if err = binary.Write(data, binary.BigEndian, int32(bLen)); err != nil {
			return nil, err
		}

		// Append the bundle
		data.Write(bBuf)
	}

	return data.Bytes(), nil
}
