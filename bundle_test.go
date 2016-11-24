package osc

import (
	"testing"
	"time"
)

func TestBundle(t *testing.T) {
	msg, err := NewMessage("/foo")
	if err != nil {
		t.Fatal(err)
	}
	b := NewBundle(time.Now(), msg)
	if _, err := b.Contents(); err != nil {
		t.Fatal(err)
	}
}
