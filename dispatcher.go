package osc

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Dispatcher for OSC packets.
type OscDispatcher struct {
	handlers map[string]Handler
}

// NewOscDispatcher returns an OscDispatcher.
func NewOscDispatcher() (dispatcher *OscDispatcher) {
	return &OscDispatcher{handlers: make(map[string]Handler)}
}

// AddMsgHandler adds a new message handler for the given OSC address.
func (self *OscDispatcher) AddMsgHandler(address string, handler HandlerFunc) error {
	for _, chr := range "*?,[]{}# " {
		if strings.Contains(address, fmt.Sprintf("%c", chr)) {
			return errors.New("OSC Address string may not contain any characters in \"*?,[]{}# \n")
		}
	}

	if existsAddress(address, self.handlers) {
		return errors.New("OSC address exists already")
	}

	self.handlers[address] = handler

	return nil
}

// Dispatch dispatches OSC packets. Implements the Dispatcher interface.
func (self *OscDispatcher) Dispatch(packet Packet) {
	switch packet.(type) {
	default:
		return

	case *Message:
		msg, _ := packet.(*Message)
		for address, handler := range self.handlers {
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
				for address, handler := range self.handlers {
					if message.Match(address) {
						handler.HandleMessage(message)
					}
				}
			}

			// Process all bundles
			for _, b := range bundle.Bundles {
				self.Dispatch(b)
			}
		}()
	}
}
