package osc

import (
	"bytes"
	"testing"
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
				[]byte{0, 0, 0, 0, 0, 0, 0, 0x0A},
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
						},
					},
				},
			},
			Output: bytes.Join([][]byte{
				ToBytes(BundleTag),
				[]byte{0, 0, 0, 0, 0, 0, 0, 0x0A},
				[]byte{0, 0, 0, 0x10},
				[]byte{'/', 'f', 'o', 'o', 0, 0, 0, 0},
				[]byte{',', TypetagInt, 0, 0},
				[]byte{0, 0, 0, 2},
			}, []byte{}),
		},
	} {
		if expected, got := testcase.Output, testcase.Input.Bytes(); !bytes.Equal(expected, got) {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
}
