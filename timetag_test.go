package osc

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestTimetag(t *testing.T) {
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

	// Test String method
	for _, testcase := range []struct {
		Input    Timetag
		Expected string
	}{
		{Input: Timetag(10), Expected: "a"},
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
