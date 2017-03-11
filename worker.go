package osc

import (
	"github.com/pkg/errors"
)

type Worker struct {
	DataChan   chan Incoming
	Dispatcher Dispatcher
	ErrChan    chan error
	Ready      chan<- Worker
}

func (w Worker) Run() error {
	w.Ready <- w

	for incoming := range w.DataChan {
		data := incoming.Data

		switch data[0] {
		case BundleTag[0]:
			go func() {
				bundle, err := ParseBundle(data, incoming.Sender)
				if err != nil {
					w.ErrChan <- err
				}
				if err := w.Dispatcher.Dispatch(bundle); err != nil {
					w.ErrChan <- errors.Wrap(err, "dispatch bundle")
				}
			}()
		case MessageChar:
			go func() {
				msg, err := ParseMessage(data, incoming.Sender)
				if err != nil {
					w.ErrChan <- err
				}
				if err := w.Dispatcher.Invoke(msg); err != nil {
					w.ErrChan <- errors.Wrap(err, "dispatch message")
				}
			}()
		default:
			w.ErrChan <- ErrParse
		}
		// Announce the worker is ready again.
		w.Ready <- w
	}
	return nil
}
