package osc

import "errors"

var (
	ErrInvalidAddress = errors.New("invalid OSC address")
)

// Method is an OSC method
type Method func(msg *Message) error

// Dispatcher dispatches OSC packets.
type Dispatcher map[string]Method

// DispatchMessage dispatches OSC message.
func (disp Dispatcher) DispatchMessage(msg *Message) error {
	for address, method := range disp {
		matched, err := msg.Match(address)
		if err != nil {
			return err
		}
		if matched {
			return method(msg)
		}
	}
	return nil
}

// DispatchBundle dispatches an OSC bundle.
func (disp Dispatcher) DispatchBundle(bun *Bundle) error {
	for address, method := range disp {
		bun.Invoke(address, method)
	}
	return nil
}
