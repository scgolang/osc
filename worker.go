package osc

import (
	"github.com/pkg/errors"
)

// worker is a worker who can process OSC messages.
type worker struct {
	DataChan   chan Incoming
	Dispatcher Dispatcher
	ErrChan    chan error
	Ready      chan<- worker
	ExactMatch bool
}

// run runs the worker.
func (w worker) run() {
	w.Ready <- w

DataLoop:
	for incoming := range w.DataChan {
		data := incoming.Data

		switch data[0] {
		case BundleTag[0]:
			bundle, err := ParseBundle(data, incoming.Sender)
			if err != nil {
				w.ErrChan <- err
			}
			if err := w.Dispatcher.Dispatch(bundle, w.ExactMatch); err != nil {
				w.ErrChan <- errors.Wrap(err, "dispatch bundle")
			}
		case MessageChar:
			msg, err := ParseMessage(data, incoming.Sender)
			if err != nil {
				w.ErrChan <- err
				continue DataLoop
			}
			if err := ValidateAddress(msg.Address); err != nil {
				w.ErrChan <- err
				continue DataLoop
			}
			if err := w.Dispatcher.Invoke(msg, w.ExactMatch); err != nil {
				w.ErrChan <- errors.Wrap(err, "dispatch message")
				continue DataLoop
			}
		default:
			w.ErrChan <- ErrParse
		}
		// Announce the worker is ready again.
		w.Ready <- w
	}
}
