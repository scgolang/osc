package osc

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestImmediately(t *testing.T) {
	if !Immediately.Time().IsZero() {
		t.Fatalf("expected Immediately to convert to the zero time")
	}
}

func TestFromTime(t *testing.T) {
	// Test converting to/from time.Time
	for _, testcase := range []struct {
		Input    Timetag
		Expected time.Time
	}{
		{Input: FromTime(time.Unix(0, 0)), Expected: time.Unix(0, 0)},
		{Input: FromTime(time.Time{}), Expected: time.Time{}},
	} {
		if expected, got := testcase.Expected, testcase.Input.Time(); !expected.Equal(got) {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestTimetagString(t *testing.T) {
	for _, testcase := range []struct {
		Input    Timetag
		Expected string
	}{
		// 0s + 0 * 0.233ns
		{Input: Timetag(0), Expected: "1900-01-01T00:00:00Z"},
		// "immediately" special value
		{Input: Timetag(1), Expected: "0001-01-01T00:00:00Z"},
		// 0s + 2 * 0.233ns
		{Input: Timetag(2), Expected: "1900-01-01T00:00:00Z"},
		// 0s + (2^32-1)/(2^32) seconds
		{Input: Timetag(0xFFFFFFFF), Expected: "1900-01-01T00:00:00Z"},
		// 1s + 0 * 0.233ns
		{Input: Timetag(0x100000000), Expected: "1900-01-01T00:00:01Z"},
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
