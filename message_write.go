package osc

import (
	"encoding/binary"
	"io"
)

// WriteTo writes the message to w.
func (msg *Message) WriteTo(w io.Writer) (n int64, err error) {
	// Write address
	nw, err := w.Write(padString(msg.Address))
	if err != nil {
		return 0, err
	}
	n += int64(nw)

	// Write typetag
	nw, err = w.Write(padBytes(append([]byte{typetagPrefix}, msg.Typetag...)))
	if err != nil {
		return 0, err
	}
	n += int64(nw)

	// Write arguments
	nw64, err := msg.writeArgs(w)
	if err != nil {
		return 0, err
	}
	n += int64(nw64)

	return n, nil
}

// writeArgs writes the messages arguments to w.
func (msg *Message) writeArgs(w io.Writer) (n int64, err error) {
	for i, tt := range msg.Typetag {
		if tt == typetagTrue || tt == typetagFalse {
			continue
		}
		arg := msg.Args[i]
		if arg == nil {
			continue
		}
		// Write blob length.
		if tt == typetagBlob {
			if err := binary.Write(w, byteOrder, int32(len(msg.Args[i]))); err != nil {
				return 0, err
			}
		}
		nw, err := w.Write(arg)
		if err != nil {
			return 0, err
		}
		n += int64(nw)

		// Write padding for strings and blobs.
		if tt == typetagString || tt == typetagBlob {
			if _, err := w.Write([]byte{0}); err != nil {
				return 0, err
			}
			n++
			for j := nw + 1; j%4 != 0; j++ {
				if _, err := w.Write([]byte{0}); err != nil {
					return 0, err
				}
				n++
			}
		}
	}
	return n, err
}

// ensureSize expands the slices that store the typetags and args.
func (msg *Message) ensureSize(size int) {
	msg.ensureTypetagSize(size)
	msg.ensureArgsSize(size)
}

// ensureTypetagSize makes sure that the Typetag slice has a certain size.
func (msg *Message) ensureTypetagSize(size int) {
	if msg.Typetag == nil {
		msg.Typetag = make([]byte, size)
	}
	if len(msg.Typetag) >= size {
		return
	}
	newtt := make([]byte, size)
	copy(newtt, msg.Typetag)
	msg.Typetag = newtt
}

// ensureArgsSize ensures that the args slice is at least the
// specified size.
func (msg *Message) ensureArgsSize(size int) {
	if msg.Args == nil {
		msg.Args = make([][]byte, size)
	}
	if len(msg.Args) >= size {
		return
	}
	newargs := make([][]byte, size)
	copy(newargs, msg.Args)
	msg.Args = newargs
}

// WriteInt32 writes an int32 value to an OSC message.
func (msg *Message) WriteInt32(index int, val int32) error {
	msg.ensureSize(index + 1)
	msg.Typetag[index] = typetagInt
	msg.Args[index] = toBytes(val)
	return nil
}

// WriteFloat32 writes a float32 value to an OSC message.
func (msg *Message) WriteFloat32(index int, val float32) error {
	msg.ensureSize(index + 1)
	msg.Typetag[index] = typetagFloat
	msg.Args[index] = toBytes(val)
	return nil
}

// WriteBool writes a boolean value to an OSC message.
func (msg *Message) WriteBool(index int, val bool) error {
	msg.ensureSize(index + 1)
	if val {
		msg.Typetag[index] = typetagTrue
	} else {
		msg.Typetag[index] = typetagFalse
	}
	msg.Args[index] = nil
	return nil
}

// WriteString writes a string value to an OSC message.
func (msg *Message) WriteString(index int, val string) error {
	msg.ensureSize(index + 1)
	msg.Typetag[index] = typetagString
	msg.Args[index] = []byte(val)
	return nil
}

// WriteBlob writes a binary blob to an OSC message.
func (msg *Message) WriteBlob(index int, blob []byte) error {
	msg.ensureSize(index + 1)
	msg.Typetag[index] = typetagBlob
	if len(blob) == 0 {
		msg.Args[index] = nil
		return nil
	}
	msg.Args[index] = blob
	return nil
}
