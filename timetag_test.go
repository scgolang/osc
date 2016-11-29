package osc

import (
	"bytes"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestFromTime(t *testing.T) {
	// Test converting to/from time.Time
	for _, testcase := range []struct {
		Input    Timetag
		Expected time.Time
	}{
		{
			Input:    FromTime(time.Unix(0, 0)),
			Expected: time.Unix(0, 0),
		},
	} {
		if expected, got := testcase.Expected, testcase.Input.Time(); !expected.Equal(got) {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestTimetagBytes(t *testing.T) {
	for _, testcase := range []struct {
		Input    Timetag
		Expected []byte
	}{
		{
			Input:    Timetag(0),
			Expected: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			Input:    Timetag(10),
			Expected: []byte{0, 0, 0, 0, 0, 0, 0, 0x0A},
		},
	} {
		if expected, got := testcase.Expected, testcase.Input.Bytes(); !bytes.Equal(expected, got) {
			t.Fatalf("expected, %q, got %q", expected, got)
		}
	}
}

func TestTimetagString(t *testing.T) {
	for _, testcase := range []struct {
		Input    Timetag
		Expected string
	}{
		{Input: Timetag(10), Expected: "1900-01-01T00:00:00Z"},
	} {
		if expected, got := testcase.Expected, testcase.Input.String(); expected != got {
			t.Fatalf("expected, %s, got %s", expected, got)
		}
	}
}

func TestReadTimetag(t *testing.T) {
	type Output struct {
		TT  Timetag
		Err error
	}
	for _, testcase := range []struct {
		Input    []byte
		Expected Output
	}{
		{
			Input:    []byte{},
			Expected: Output{Err: errors.New("timetags must be 64-bit")},
		},
		{
			Input:    []byte{0, 0, 0, 0, 0, 0, 0, 1},
			Expected: Output{TT: Timetag(1)},
		},
	} {
		tt, err := ReadTimetag(testcase.Input)

		if testcase.Expected.Err == nil {
			if err != nil {
				t.Fatal(err)
			}
			if expected, got := testcase.Expected.TT, tt; expected != got {
				t.Fatalf("expected %s, got %s", expected, got)
			}
		} else {
			if expected, got := testcase.Expected.Err.Error(), err.Error(); expected != got {
				t.Fatalf("expected %s, got %s", expected, got)
			}
		}
	}
}
