package osc

import "errors"

var (
	ErrInvalidAddress = errors.New("invalid OSC address")
)

// oscDispatcher dispatches OSC packets.
type oscDispatcher map[string]Method

// dispatchMessage dispatches OSC message.
func (disp oscDispatcher) dispatchMessage(msg *Message) error {
	for address, method := range disp {
		matched, err := msg.Match(address)
		if err != nil {
			return err
		}
		if matched {
			method(msg)
		}
	}
	return nil
}

// dispatchBundle dispatches an OSC bundle.
func (disp oscDispatcher) dispatchBundle(bun *Bundle) error {
	for address, method := range disp {
		bun.Invoke(address, method)
	}
	return nil
}
