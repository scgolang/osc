package osc

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestWorkerRun(t *testing.T) {
	var (
		data  = make(chan Incoming)
		errch = make(chan error)
		ready = make(chan Worker)
	)
	worker := Worker{
		DataChan:   data,
		Dispatcher: errorDispatcher{},
		ErrChan:    errch,
		Ready:      ready,
	}
	// Worker exits when the data chan is closed.
	defer close(data)

	// Run the worker goroutine.
	go worker.Run()

	// Wait for the worker to signal that it is ready.
	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout receiving on ready chan")
	}
	// Send some data.
	incoming := Incoming{
		Data: Message{Address: "/foo"}.Bytes(),
	}
	select {
	case data <- incoming:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout sending on data chan")
	}
	// Dispatcher will generate an error.
	select {
	case err := <-errch:
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout receiving on error chan")
	}
}

type errorDispatcher struct {
}

func (d errorDispatcher) Dispatch(bundle Bundle, exactMatch bool) error {
	return errors.New("fake Dispatch error")
}

func (d errorDispatcher) Invoke(msg Message, exactMatch bool) error {
	return errors.New("fake Invoke error")
}
