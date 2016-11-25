package osc

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// Argument represents an OSC argument.
// An OSC argument can have many different types, which is why
// we choose to represent them with an interface.
type Argument interface {
	io.WriterTo

	Bytes() []byte
	Equal(Argument) bool
	ReadInt32() (int32, error)
	ReadFloat32() (float32, error)
	ReadBool() (bool, error)
	ReadString() (string, error)
	ReadBlob() ([]byte, error)
	String() string
	Typetag() byte
}

// ReadArguments reads all arguments from the reader and adds it to the OSC message.
func ReadArguments(typetags, data []byte) ([]Argument, error) {
	args := []Argument{}

	// Strip off the prefix.
	if len(typetags) > 0 && typetags[0] == TypetagPrefix {
		typetags = typetags[1:]
	}

	for i, tt := range typetags {
		arg, idx, err := ReadArgument(tt, data)
		if err != nil {
			return nil, errors.Wrapf(err, "read argument %d", i)
		}
		args = append(args, arg)
		data = data[idx:]
	}
	return args, nil
}

// ReadArgument parses an OSC message argument given a type tag and some data.
func ReadArgument(tt byte, data []byte) (Argument, int64, error) {
	switch tt {
	case TypetagInt:
		return ReadIntFrom(data)
	case TypetagFloat:
		return ReadFloatFrom(data)
	case TypetagTrue:
		return Bool(true), 0, nil
	case TypetagFalse:
		return Bool(false), 0, nil
	case TypetagString:
		s, idx := ReadString(data)
		return String(s), idx, nil
	case TypetagBlob:
		return ReadBlobFrom(data)
	default:
		return nil, 0, errors.Wrapf(ErrInvalidTypeTag, "typetag %q", string(tt))
	}
}

// Int represents a 32-bit integer.
type Int int32

// ReadIntFrom reads a 32-bit integer from a byte slice.
func ReadIntFrom(data []byte) (Argument, int64, error) {
	var i Int
	if err := binary.Read(bytes.NewReader(data), byteOrder, &i); err != nil {
		return nil, 0, errors.Wrap(err, "read int argument")
	}
	return i, 4, nil
}

// Bytes converts the arg to a byte slice suitable for adding to the binary representation of an OSC message.
func (i Int) Bytes() []byte {
	return []byte{
		byte(int32(i) >> 24),
		byte(int32(i) >> 16),
		byte((int32(i) >> 8)),
		byte(i),
	}
}

// Equal returns true if the argument equals the other one, false otherwise.
func (i Int) Equal(other Argument) bool {
	if other.Typetag() != TypetagInt {
		return false
	}
	i2 := other.(Int)
	return i == i2
}

// ReadInt32 reads a 32-bit integer from the arg.
func (i Int) ReadInt32() (int32, error) { return int32(i), nil }

// ReadFloat32 reads a 32-bit float from the arg.
func (i Int) ReadFloat32() (float32, error) { return 0, ErrInvalidTypeTag }

// ReadBool bool reads a boolean from the arg.
func (i Int) ReadBool() (bool, error) { return false, ErrInvalidTypeTag }

// ReadString string reads a string from the arg.
func (i Int) ReadString() (string, error) { return "", ErrInvalidTypeTag }

// ReadBlob reads a slice of bytes from the arg.
func (i Int) ReadBlob() ([]byte, error) { return nil, ErrInvalidTypeTag }

// String converts the arg to a string.
func (i Int) String() string { return fmt.Sprintf("Int(%d)", i) }

// Typetag returns the argument's type tag.
func (i Int) Typetag() byte { return TypetagInt }

// WriteTo writes the arg to an io.Writer.
func (i Int) WriteTo(w io.Writer) (int64, error) {
	written, err := fmt.Fprintf(w, "%d", i)
	return int64(written), err
}

// Float represents a 32-bit float.
type Float float32

// ReadFloatFrom reads a 32-bit float from a byte slice.
func ReadFloatFrom(data []byte) (Argument, int64, error) {
	var f Float
	if err := binary.Read(bytes.NewReader(data), byteOrder, &f); err != nil {
		return nil, 0, errors.Wrap(err, "read float argument")
	}
	return f, 4, nil
}

// Bytes converts the arg to a byte slice suitable for adding to the binary representation of an OSC message.
func (f Float) Bytes() []byte {
	var (
		buf = &bytes.Buffer{}
		_   = binary.Write(buf, byteOrder, float32(f)) // Never fails
	)
	return buf.Bytes()
}

// Equal returns true if the argument equals the other one, false otherwise.
func (f Float) Equal(other Argument) bool {
	if other.Typetag() != TypetagFloat {
		return false
	}
	f2 := other.(Float)
	return f == f2
}

// ReadInt32 reads a 32-bit integer from the arg.
func (f Float) ReadInt32() (int32, error) { return 0, ErrInvalidTypeTag }

// ReadFloat32 reads a 32-bit float from the arg.
func (f Float) ReadFloat32() (float32, error) { return float32(f), nil }

// ReadBool bool reads a boolean from the arg.
func (f Float) ReadBool() (bool, error) { return false, ErrInvalidTypeTag }

// ReadString string reads a string from the arg.
func (f Float) ReadString() (string, error) { return "", ErrInvalidTypeTag }

// ReadBlob reads a slice of bytes from the arg.
func (f Float) ReadBlob() ([]byte, error) { return nil, ErrInvalidTypeTag }

// String converts the arg to a string.
func (f Float) String() string { return fmt.Sprintf("Float(%f)", f) }

// Typetag returns the argument's type tag.
func (f Float) Typetag() byte { return TypetagFloat }

// WriteTo writes the arg to an io.Writer.
func (f Float) WriteTo(w io.Writer) (int64, error) {
	written, err := fmt.Fprintf(w, "%f", f)
	return int64(written), err
}

// Bool represents a boolean value.
type Bool bool

// Bytes converts the arg to a byte slice suitable for adding to the binary representation of an OSC message.
func (b Bool) Bytes() []byte {
	return []byte{}
}

// Equal returns true if the argument equals the other one, false otherwise.
func (b Bool) Equal(other Argument) bool {
	if other.Typetag() != TypetagFalse && other.Typetag() != TypetagTrue {
		return false
	}
	b2 := other.(Bool)
	return b == b2
}

// ReadInt32 reads a 32-bit integer from the arg.
func (b Bool) ReadInt32() (int32, error) { return 0, ErrInvalidTypeTag }

// ReadFloat32 reads a 32-bit float from the arg.
func (b Bool) ReadFloat32() (float32, error) { return 0, ErrInvalidTypeTag }

// ReadBool bool reads a boolean from the arg.
func (b Bool) ReadBool() (bool, error) { return bool(b), nil }

// ReadString string reads a string from the arg.
func (b Bool) ReadString() (string, error) { return "", ErrInvalidTypeTag }

// ReadBlob reads a slice of bytes from the arg.
func (b Bool) ReadBlob() ([]byte, error) { return nil, ErrInvalidTypeTag }

// String converts the arg to a string.
func (b Bool) String() string { return fmt.Sprintf("Bool(%t)", b) }

// Typetag returns the argument's type tag.
func (b Bool) Typetag() byte {
	if bool(b) {
		return TypetagTrue
	}
	return TypetagFalse
}

// WriteTo writes the arg to an io.Writer.
func (b Bool) WriteTo(w io.Writer) (int64, error) {
	written, err := fmt.Fprintf(w, "%t", b)
	return int64(written), err
}

// String is a string.
type String string

// Bytes converts the arg to a byte slice suitable for adding to the binary representation of an OSC message.
func (s String) Bytes() []byte {
	return ToBytes(string(s))
}

// Equal returns true if the argument equals the other one, false otherwise.
func (s String) Equal(other Argument) bool {
	if other.Typetag() != TypetagString {
		return false
	}
	s2 := other.(String)
	return s == s2
}

// ReadInt32 reads a 32-bit integer from the arg.
func (s String) ReadInt32() (int32, error) { return 0, ErrInvalidTypeTag }

// ReadFloat32 reads a 32-bit float from the arg.
func (s String) ReadFloat32() (float32, error) { return 0, ErrInvalidTypeTag }

// ReadBool bool reads a boolean from the arg.
func (s String) ReadBool() (bool, error) { return false, ErrInvalidTypeTag }

// ReadString string reads a string from the arg.
func (s String) ReadString() (string, error) { return string(s), nil }

// ReadBlob reads a slice of bytes from the arg.
func (s String) ReadBlob() ([]byte, error) { return nil, ErrInvalidTypeTag }

// String converts the arg to a string.
func (s String) String() string { return string(s) }

// Typetag returns the argument's type tag.
func (s String) Typetag() byte { return TypetagString }

// WriteTo writes the arg to an io.Writer.
func (s String) WriteTo(w io.Writer) (int64, error) {
	written, err := fmt.Fprintf(w, "%s", s)
	return int64(written), err
}

// Blob is a slice of bytes.
type Blob []byte

// ReadBlobFrom reads a binary blob from the provided data.
func ReadBlobFrom(data []byte) (Argument, int64, error) {
	var length int32
	if err := binary.Read(bytes.NewReader(data), byteOrder, &length); err != nil {
		return nil, 0, errors.Wrap(err, "read blob argument")
	}
	b, bl := ReadBlob(length, data[4:])
	return Blob(b), bl + 4, nil
}

// Bytes converts the arg to a byte slice suitable for adding to the binary representation of an OSC message.
func (b Blob) Bytes() []byte {
	return Pad(bytes.Join([][]byte{
		Int(len(b)).Bytes(),
		[]byte(b),
	}, []byte{}))
}

// Equal returns true if the argument equals the other one, false otherwise.
func (b Blob) Equal(other Argument) bool {
	if other.Typetag() != TypetagBlob {
		return false
	}
	b2 := other.(Blob)
	if len(b) != len(b2) {
		return false
	}
	return bytes.Equal(b, b2)
}

// ReadInt32 reads a 32-bit integer from the arg.
func (b Blob) ReadInt32() (int32, error) { return 0, ErrInvalidTypeTag }

// ReadFloat32 reads a 32-bit float from the arg.
func (b Blob) ReadFloat32() (float32, error) { return 0, ErrInvalidTypeTag }

// ReadBool bool reads a boolean from the arg.
func (b Blob) ReadBool() (bool, error) { return false, ErrInvalidTypeTag }

// ReadString string reads a string from the arg.
func (b Blob) ReadString() (string, error) { return "", ErrInvalidTypeTag }

// ReadBlob reads a slice of bytes from the arg.
func (b Blob) ReadBlob() ([]byte, error) { return []byte(b), nil }

// String converts the arg to a string.
func (b Blob) String() string { return base64.StdEncoding.EncodeToString([]byte(b)) }

// Typetag returns the argument's type tag.
func (b Blob) Typetag() byte { return TypetagBlob }

// WriteTo writes the arg to an io.Writer.
func (b Blob) WriteTo(w io.Writer) (int64, error) {
	written, err := w.Write([]byte(b))
	return int64(written), err
}

// Arguments is a slice of Argument.
type Arguments []Argument
