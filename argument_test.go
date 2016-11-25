package osc

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/pkg/errors"
)

type equalTest struct {
	arg      Argument
	equal    []Argument
	notEqual []Argument
}

func (e equalTest) run(t *testing.T) {
	for i, eq := range e.equal {
		if !e.arg.Equal(eq) {
			t.Fatalf("equals[%d] expected %s to equal %s", i, e.arg, eq)
		}
	}
	for i, ne := range e.notEqual {
		if e.arg.Equal(ne) {
			t.Fatalf("notEquals[%d] expected %s to not equal %s", i, e.arg, ne)
		}
	}
}

func TestIntBytes(t *testing.T) {
	for _, testcase := range []struct {
		Int      Int
		Expected []byte
	}{
		{
			Int:      Int(1),
			Expected: []byte{0, 0, 0, 1},
		},
		{
			Int:      Int(256),
			Expected: []byte{0, 0, 1, 0},
		},
	} {
		if expected, got := testcase.Expected, testcase.Int.Bytes(); !bytes.Equal(expected, got) {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
}

func TestIntEqual(t *testing.T) {
	equalTest{
		arg:      Int(0),
		equal:    []Argument{Int(0)},
		notEqual: []Argument{Int(2), String("Foo")},
	}.run(t)
}

func TestIntReadInt32(t *testing.T) {
	arg := Int(0)
	i, err := arg.ReadInt32()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := int32(0), i; expected != got {
		t.Fatalf("expected %d, got %d", expected, got)
	}
}

func TestIntReadOther(t *testing.T) {
	i := Int(0)
	if _, err := i.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := i.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := i.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := i.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
}

func TestIntString(t *testing.T) {
	arg := Int(0)
	if expected, got := "Int(0)", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
}

func TestIntTypetag(t *testing.T) {
	arg := Int(0)
	if expected, got := TypetagInt, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
}

func TestIntWriteTo(t *testing.T) {
	arg := Int(0)
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestFloatBytes(t *testing.T) {
	if expected, got := []byte{0x40, 0x48, 0xf5, 0xc3}, Float(3.14).Bytes(); !bytes.Equal(expected, got) {
		t.Fatalf("expected %x, got %x", expected, got)
	}
}

func TestFloatEqual(t *testing.T) {
	equalTest{
		arg:      Float(0),
		equal:    []Argument{Float(0)},
		notEqual: []Argument{Float(3.14), String("foo")},
	}.run(t)
}

func TestFloatReadFloat(t *testing.T) {
	arg := Float(0)
	f, err := arg.ReadFloat32()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := float32(0), f; expected != got {
		t.Fatalf("expected %f, got %f", expected, got)
	}
}

func TestFloatReadOther(t *testing.T) {
	f := Float(0)
	if _, err := f.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := f.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := f.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := f.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
}

func TestFloatString(t *testing.T) {
	arg := Float(0)
	if expected, got := "Float(0.000000)", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
}

func TestFloatTypetag(t *testing.T) {
	arg := Float(0)
	if expected, got := TypetagFloat, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
}

func TestFloatWriteTo(t *testing.T) {
	arg := Float(0)
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestBoolBytes(t *testing.T) {
	arg := Bool(false)
	if expected, got := []byte{}, arg.Bytes(); !bytes.Equal(expected, got) {
		t.Fatalf("expected %x, got %x", expected, got)
	}
}

func TestBoolEqual(t *testing.T) {
	arg := Bool(false)
	if other := Int(3); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
	if other := Bool(false); !arg.Equal(other) {
		t.Fatalf("expected %s to equal %s", arg, other)
	}
}

func TestBoolReadBool(t *testing.T) {
	arg := Bool(false)
	b, err := arg.ReadBool()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := false, b; expected != got {
		t.Fatalf("expected %t, got %t", expected, got)
	}
}

func TestBoolReadOther(t *testing.T) {
	arg := Bool(false)
	if _, err := arg.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
}

func TestBoolString(t *testing.T) {
	arg := Bool(false)
	if expected, got := "Bool(false)", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
}

func TestBoolTypetag(t *testing.T) {
	arg := Bool(false)
	if expected, got := TypetagFalse, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
	argTrue := Bool(true)
	if expected, got := TypetagTrue, argTrue.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
}

func TestBoolWriteTo(t *testing.T) {
	arg := Bool(false)
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestStringBytes(t *testing.T) {
	arg := String("foo")
	if expected, got := []byte{'f', 'o', 'o', 0}, arg.Bytes(); !bytes.Equal(expected, got) {
		t.Fatalf("expected %x, got %x", expected, got)
	}
}

func TestStringEqual(t *testing.T) {
	arg := String("foo")
	if other := String("foo"); !arg.Equal(other) {
		t.Fatalf("expected %s to equal %s", arg, other)
	}
	if other := Int(4); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
}

func TestStringReadString(t *testing.T) {
	arg := String("foo")
	s, err := arg.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := "foo", s; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestStringReadOther(t *testing.T) {
	arg := String("foo")
	if _, err := arg.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBlob(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
}

func TestStringString(t *testing.T) {
	arg := String("foo")
	if expected, got := "foo", arg.String(); expected != got {
		t.Fatalf("expected %s to equal %s", expected, got)
	}
}

func TestStringTypetag(t *testing.T) {
	arg := String("foo")
	if expected, got := TypetagString, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
}

func TestStringWriteTo(t *testing.T) {
	arg := String("foo")
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestBlobBytes(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	if expected, got := []byte{0, 0, 0, 3, 'f', 'o', 'o', 0}, arg.Bytes(); !bytes.Equal(expected, got) {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestBlobEqual(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	if other := Blob([]byte{'f', 'o', 'o'}); !arg.Equal(other) {
		t.Fatalf("expected %s to equal %s", arg, other)
	}
	if other := Blob([]byte{'f'}); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
	if other := String("bar"); arg.Equal(other) {
		t.Fatalf("expected %s to not equal %s", arg, other)
	}
}

func TestBlobReadBlob(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	b, err := arg.ReadBlob()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := []byte{'f', 'o', 'o'}, b; !bytes.Equal(expected, got) {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestBlobReadOther(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	if _, err := arg.ReadInt32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadFloat32(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadBool(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
	if _, err := arg.ReadString(); err != ErrInvalidTypeTag {
		t.Fatalf("expected ErrInvalidTypeTag, got %+v", err)
	}
}

func TestBlobString(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	b, err := arg.ReadBlob()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := base64.StdEncoding.EncodeToString(b), arg.String(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestBlobTypetag(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	if expected, got := TypetagBlob, arg.Typetag(); expected != got {
		t.Fatalf("expected %c, got %c", expected, got)
	}
}

func TestBlobWriteTo(t *testing.T) {
	arg := Blob([]byte{'f', 'o', 'o'})
	if _, err := arg.WriteTo(ioutil.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestReadArgument(t *testing.T) {
	type Input struct {
		tt   byte
		data []byte
	}
	type Output struct {
		Argument Argument
		Consumed int64 // number of bytes consumed
		Err      error
	}
	for _, testcase := range []struct {
		Input    Input
		Expected Output
	}{
		{
			Input:    Input{tt: TypetagInt, data: []byte{0, 0, 0, 1}},
			Expected: Output{Argument: Int(1), Consumed: 4},
		},
		{
			Input:    Input{tt: TypetagInt, data: []byte{}},
			Expected: Output{Err: errors.New("read int argument: EOF")},
		},
		{
			Input:    Input{tt: TypetagFloat, data: []byte{0x40, 0x48, 0xf5, 0xc3}},
			Expected: Output{Argument: Float(3.14), Consumed: 4},
		},
		{
			Input:    Input{tt: TypetagFloat, data: []byte{}},
			Expected: Output{Err: errors.New("read float argument: EOF")},
		},
		{
			Input:    Input{tt: TypetagTrue},
			Expected: Output{Argument: Bool(true)},
		},
		{
			Input:    Input{tt: TypetagFalse},
			Expected: Output{Argument: Bool(false)},
		},
		{
			Input:    Input{tt: TypetagString, data: []byte{'a', 'b', 'c', 'd', 'e'}},
			Expected: Output{Argument: String("abcde"), Consumed: 8},
		},
		{
			// Length followed by blob
			Input:    Input{tt: TypetagBlob, data: []byte{0, 0, 0, 5, 'a', 'b', 'c', 'd', 'e'}},
			Expected: Output{Argument: Blob([]byte{'a', 'b', 'c', 'd', 'e', 0, 0, 0}), Consumed: 12},
		},
		{
			Input:    Input{tt: TypetagBlob, data: []byte{}},
			Expected: Output{Err: errors.New("read blob argument: EOF")},
		},
		{
			Input:    Input{tt: 'Q'},
			Expected: Output{Err: errors.Wrap(ErrInvalidTypeTag, `typetag "Q"`)},
		},
	} {
		a, consumed, err := ReadArgument(testcase.Input.tt, testcase.Input.data)
		if testcase.Expected.Err == nil {
			if expected, got := testcase.Expected.Err, err; expected != got {
				t.Fatalf("expected %s, got %s", expected, got)
			}
			if expected, got := testcase.Expected.Consumed, consumed; expected != got {
				t.Fatalf("expected %d, got %d", expected, got)
			}
			if expected, got := testcase.Expected.Argument, a; !expected.Equal(got) {
				t.Fatalf("expected %+v, got %+v", expected, got)
			}
		} else {
			if expected, got := testcase.Expected.Err.Error(), err.Error(); expected != got {
				t.Fatalf("expected %s, got %s", expected, got)
			}
		}
	}
}

func TestReadArguments(t *testing.T) {
	type Input struct {
		Typetags []byte
		Data     []byte
	}
	type Output struct {
		Arguments []Argument
		Err       error
	}
	for _, testcase := range []struct {
		Input    Input
		Expected Output
	}{
		{
			Input:    Input{Typetags: []byte{}, Data: []byte{}},
			Expected: Output{Arguments: []Argument{}},
		},
		{
			Input:    Input{Typetags: []byte{TypetagInt}, Data: []byte{}},
			Expected: Output{Err: errors.New("read argument 0: read int argument: EOF")},
		},
		{
			Input:    Input{Typetags: []byte{TypetagInt}, Data: []byte{0, 0, 0, 1}},
			Expected: Output{Arguments: []Argument{Int(1)}},
		},
		{
			Input: Input{Typetags: []byte{TypetagBlob}, Data: []byte{0, 0, 1, 1, 4, 5, 6, 7}},
			Expected: Output{
				Arguments: []Argument{
					Blob([]byte{4, 5, 6, 7}),
				},
			},
		},
	} {
		args, err := ReadArguments(testcase.Input.Typetags, testcase.Input.Data)

		if testcase.Expected.Err == nil {
			if err != nil {
				t.Fatal(err)
			}
			if expected, got := len(testcase.Expected.Arguments), len(args); expected != got {
				t.Fatalf("expected %d arguments, got %d", expected, got)
			}
			for i, arg := range args {
				if expected, got := testcase.Expected.Arguments[i], arg; !expected.Equal(got) {
					t.Fatalf("(argument %d) expected %s, got %s", i, expected, got)
				}
			}
		} else {
			if expected, got := testcase.Expected.Err.Error(), err.Error(); expected != got {
				t.Fatalf("expected %s, got %s", expected, got)
			}
		}
	}
}
