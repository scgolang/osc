package osc

import (
	"bytes"
	"testing"

	"github.com/pkg/errors"
)

func TestBundleBytes(t *testing.T) {
	for _, testcase := range []struct {
		Input  Bundle
		Output []byte
	}{
		{
			Input: Bundle{Timetag: 10},
			Output: bytes.Join([][]byte{
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x0A},
			}, []byte{}),
		},
		{
			Input: Bundle{
				Timetag: 10,
				Packets: []Packet{
					Message{
						Address: "/foo",
						Arguments: Arguments{
							Int(2),
							Float(5.2314),
						},
					},
				},
			},
			Output: bytes.Join([][]byte{
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x0A},
				{0, 0, 0, 0x14},
				{'/', 'f', 'o', 'o', 0, 0, 0, 0},
				{',', TypetagInt, TypetagFloat, 0},
				{0, 0, 0, 2},
				{0x40, 0xA7, 0x67, 0xA1},
			}, []byte{}),
		},
		// Bundle within a bundle.
		{
			Input: Bundle{
				Timetag: 10,
				Packets: []Packet{
					Bundle{
						Timetag: 20,
						Packets: []Packet{
							Message{
								Address: "/foobar",
								Arguments: Arguments{
									Float(1),
								},
							},
						},
					},
					Message{
						Address: "/foo",
						Arguments: Arguments{
							Int(2),
							Float(5.2314),
						},
					},
				},
			},
			Output: bytes.Join([][]byte{
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x0A}, // Timetag
				{0, 0, 0, 0x24},             // Length of first bundle element
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x14}, // Timetag
				{0, 0, 0, 0x10},             // Length of first element of bundle within bundle.
				{'/', 'f', 'o', 'o', 'b', 'a', 'r', 0},
				{TypetagPrefix, TypetagFloat, 0, 0},
				{0x3F, 0x80, 0x00, 0x00},
				{0, 0, 0, 0x14}, // Length of second bundle element
				{'/', 'f', 'o', 'o', 0, 0, 0, 0},
				{TypetagPrefix, TypetagInt, TypetagFloat, 0},
				{0, 0, 0, 2},
				{0x40, 0xA7, 0x67, 0xA1},
			}, []byte{}),
		},
	} {
		if expected, got := testcase.Output, testcase.Input.Bytes(); !bytes.Equal(expected, got) {
			t.Fatalf("expected %q\n                got %q", expected, got)
		}
	}
}

func TestBundleEqual(t *testing.T) {
	for _, testcase := range []struct {
		b  Bundle
		e  []Bundle
		ne []Packet
	}{
		{
			b: Bundle{Timetag: 5},
			e: []Bundle{
				{Timetag: 5},
			},
			ne: []Packet{
				Message{},
				Bundle{Timetag: 2},
				Bundle{
					Timetag: 5,
					Packets: []Packet{
						Message{Address: "/foo"},
					},
				},
			},
		},
		{
			b: Bundle{
				Timetag: 5,
				Packets: []Packet{
					Message{Address: "/bar"},
				},
			},
			ne: []Packet{
				Bundle{
					Timetag: 5,
					Packets: []Packet{
						Message{Address: "/foo"},
					},
				},
			},
		},
	} {
		b := testcase.b
		for i, e := range testcase.e {
			if !b.Equal(e) {
				t.Fatalf("(testcase %d) expected %q to equal %q", i, b, e)
			}
		}
		for i, ne := range testcase.ne {
			if b.Equal(ne) {
				t.Fatalf("(testcase %d) expected %q to not equal %q", i, b, ne)
			}
		}
	}
}

func TestParseBundle(t *testing.T) {
	type Output struct {
		bundle Bundle
		err    error
	}
	for i, testcase := range []struct {
		Input    []byte
		Expected Output
	}{
		// testcase 0
		{
			Input: []byte{},
			Expected: Output{
				err: errors.New(`slice bundle tag: expected "#bundle\x00", got ""`),
			},
		},
		// testcase 1
		{
			Input: append([]byte("#fundle"), 0),
			Expected: Output{
				err: errors.New(`slice bundle tag: expected "#bundle\x00", got "#fundle\x00"`),
			},
		},
		// testcase 2
		{
			Input: bytes.Join(
				[][]byte{
					append([]byte("#bundle"), 0),
					{1, 2, 3, 4, 5, 6, 7},
				},
				[]byte{},
			),
			Expected: Output{
				err: errors.New(`read timetag: timetags must be 64-bit`),
			},
		},
		// testcase 3
		{
			Input: bytes.Join(
				[][]byte{
					append([]byte("#bundle"), 0),
					{1, 2, 3, 4, 5, 6, 7, 8},
					{0, 0, 0, 4},
					{'%', 'n', 'o', 0},
				},
				[]byte{},
			),
			Expected: Output{
				err: errors.New(`read packets: read packet: packet should never start with %`),
			},
		},
		// testcase 4
		{
			Input: bytes.Join(
				[][]byte{
					append([]byte("#bundle"), 0),
					Timetag(50).Bytes(),
					{0, 0, 0, 0x10},
					{'/', 'f', 'o', 'o', 'b', 'a', 'r', 0},
					{TypetagPrefix, TypetagInt, 0, 0},
					{0, 0, 0, 7},
				},
				[]byte{},
			),
			Expected: Output{
				bundle: Bundle{
					Timetag: Timetag(50),
					Packets: []Packet{
						Message{
							Address:   "/foobar",
							Arguments: Arguments{Int(7)},
						},
					},
				},
			},
		},
		// testcase 5
		{
			Input: bytes.Join(
				[][]byte{
					append([]byte("#bundle"), 0),
					Timetag(50).Bytes(),
					{0, 0, 0x10},
				},
				[]byte{},
			),
			Expected: Output{
				bundle: Bundle{Timetag: Timetag(50)},
			},
		},
		// testcase 6
		{
			Input: bytes.Join(
				[][]byte{
					append([]byte("#bundle"), 0),
					Timetag(50).Bytes(),
					{0, 0, 0, 0x20},
					{'/', 'f', 'o', 'o', 'b', 'a', 'r', 0},
					{TypetagPrefix, TypetagInt, 0, 0},
					{0, 0, 0, 7},
				},
				[]byte{},
			),
			Expected: Output{
				err: errors.New("read packets: read packet: packet length 32 is greater than data length 16"),
			},
		},
		// testcase 7
		{
			Input: bytes.Join(
				[][]byte{
					append([]byte("#bundle"), 0),
					Timetag(50).Bytes(),
					{0, 0, 0, 0x10},
					{'/', 'f', 'o', 'o', 'b', 'a', 'r', 0},
					{TypetagPrefix, TypetagInt, 0, 0},
					{0, 0, 0, 7},
					{0, 0, 0, 0x10},
					{'/', 'b', 'o', 'r', 'k', 0, 0, 0},
					{TypetagPrefix, TypetagFloat, 0, 0},
					{0x40, 0x48, 0xf5, 0xc3},
				},
				[]byte{},
			),
			Expected: Output{
				bundle: Bundle{
					Timetag: Timetag(50),
					Packets: []Packet{
						Message{
							Address:   "/foobar",
							Arguments: Arguments{Int(7)},
						},
						Message{
							Address:   "/bork",
							Arguments: Arguments{Float(3.14)},
						},
					},
				},
			},
		},
		// testcase 8
		{
			Input: bytes.Join(
				[][]byte{
					append([]byte("#bundle"), 0),
					Timetag(50).Bytes(),
					{0, 0, 0, 0x10},
					{'/', 'f', 'o', 'o', 'b', 'a', 'r', 0},
					{TypetagPrefix, TypetagInt, 'Q', 0},
					{0, 0, 0, 7},
				},
				[]byte{},
			),
			Expected: Output{
				err: errors.New(`read packets: read packet: parse message from packet: parse message: read argument 1: typetag "Q": invalid type tag`),
			},
		},
		// testcase 9
		{
			Input: bytes.Join([][]byte{
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x0A}, // Timetag
				{0, 0, 0, 0x24},             // Length of first bundle element
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x14}, // Timetag
				{0, 0, 0, 0x10},             // Length of first element of bundle within bundle.
				{'/', 'f', 'o', 'o', 'b', 'a', 'r', 0},
				{TypetagPrefix, TypetagFloat, 0, 0},
				{0x3F, 0x80, 0x00, 0x00},
				{0, 0, 0, 0x14}, // Length of second bundle element
				{'/', 'f', 'o', 'o', 0, 0, 0, 0},
				{TypetagPrefix, TypetagInt, TypetagFloat, 0},
				{0, 0, 0, 2},
				{0x40, 0xA7, 0x67, 0xA1},
				{0, 0, 0, 0x10}, // Length of third bundleelement
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x28}, // Timetag
			}, []byte{}),
			Expected: Output{
				bundle: Bundle{
					Timetag: 10,
					Packets: []Packet{
						Bundle{
							Timetag: 20,
							Packets: []Packet{
								Message{
									Address: "/foobar",
									Arguments: Arguments{
										Float(1),
									},
								},
							},
						},
						Message{
							Address: "/foo",
							Arguments: Arguments{
								Int(2),
								Float(5.2314),
							},
						},
						Bundle{Timetag: 40},
					},
				},
			},
		},
		// testcase 10
		{
			Input: bytes.Join([][]byte{
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x0A}, // Timetag
				{0, 0, 0, 0x24},             // Length of first bundle element
				ToBytes(BundleTag),
				{0, 0, 0, 0, 0, 0, 0, 0x14}, // Timetag
				{0, 0, 0, 0x10},             // Length of first element of bundle within bundle.
				{'/', 'f', 'o', 'o', 'b', 'a', 'r', 0},
				{TypetagPrefix, 'Q', 0, 0},
				{0x3F, 0x80, 0x00, 0x00},
			}, []byte{}),
			Expected: Output{
				err: errors.New(`read packets: read packet: parse bundle from packet: read packets: read packet: parse message from packet: parse message: read argument 0: typetag "Q": invalid type tag`),
			},
		},
	} {
		b, err := ParseBundle(testcase.Input, nil)
		if testcase.Expected.err == nil {
			if err != nil {
				t.Fatalf("(testcase %d) %s", i, err)
			}
			if expected, got := testcase.Expected.bundle, b; !expected.Equal(got) {
				t.Fatalf("(testcase %d) expected %q\n                              got %q", i, expected, got)
			}
		} else {
			if expected, got := testcase.Expected.err.Error(), err.Error(); expected != got {
				t.Fatalf("(testcase %d) expected %s, got %s", i, expected, got)
			}
		}
	}
}

func TestParseBundleLimit(t *testing.T) {
	// Test the limit parameter of parseBundle.
	_, limitErr := parseBundle(nil, nil, 10)
	if expected, got := errors.New("limit must be >= 16 or < 0"), limitErr; got == nil || (expected.Error() != got.Error()) {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
