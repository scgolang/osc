package osc

import (
	"errors"
	"time"
)

var (
	ErrInvalidAddress = errors.New("invalid OSC address")
)

// oscDispatcher dispatches OSC packets.
type oscDispatcher map[string]HandlerFunc

// dispatch dispatches OSC packets. Implements the Dispatcher interface.
func (disp oscDispatcher) dispatch(packet Packet) {
	switch packet.(type) {
	default:
		return

	case *Message:
		msg, _ := packet.(*Message)
		for address, handler := range disp {
			if msg.Match(address) {
				handler.HandleMessage(msg)
			}
		}

	case *Bundle:
		bundle, _ := packet.(*Bundle)
		timer := time.NewTimer(bundle.Timetag.ExpiresIn())

		go func() {
			<-timer.C
			for _, message := range bundle.Messages {
				for address, handler := range disp {
					if message.Match(address) {
						handler.HandleMessage(message)
					}
				}
			}

			// Process all bundles
			for _, b := range bundle.Bundles {
				disp.dispatch(b)
			}
		}()
	}
}
