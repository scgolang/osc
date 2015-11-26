package osc

import "testing"

func TestMatch(t *testing.T) {
	// matches where 1st string is pattern and second is address
	for _, pair := range [][2]string{
		{"/path/to/method", "/path/to/method"},
		{"/path/to/meth?d", "/path/to/method"},
		{"/path/to/*", "/path/to/method"},
		{"/path/to/method*", "/path/to/method"},
		{"/path/to/m[aei]thod", "/path/to/method"},
	} {
		msg, err := NewMessage(pair[0])
		if err != nil {
			t.Fatal(err)
		}
		match, err := msg.Match(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Fatalf("Expected %s to match %s", pair[1], pair[0])
		}
	}

	// misses where 1st string is pattern and second is address
	for _, pair := range [][2]string{
		{"/path/to/destruction", "/path/to/method"},
		{"/path/to/me?thod", "/path/to/method"},
		{"/path/to?method", "/path/to/method"},
		{"/path/to*", "/path/to/method"},
		{"/path/to/[domet]", "/path/to/method"},
	} {
		msg, err := NewMessage(pair[0])
		if err != nil {
			t.Fatal(err)
		}
		match, err := msg.Match(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if match {
			t.Fatalf("Expected %s to not match %s", pair[1], pair[0])
		}
	}
}
