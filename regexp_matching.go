package osc

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// RegexpMatching is a dispatcher that simply uses a regexp to match a message.
// If you are looking for the official OSC Pattern Matching dispatcher, see [PatternMatching].

type RegexpMatching map[string]MessageHandler

// Dispatch invokes an OSC bundle's messages.
func (h RegexpMatching) Dispatch(b Bundle, exactMatch bool) error {
	var (
		now = time.Now()
		tt  = b.Timetag.Time()
	)
	if tt.Before(now) {
		return h.immediately(b, exactMatch)
	}
	<-time.After(tt.Sub(now))
	return h.immediately(b, exactMatch)
}

// immediately invokes an OSC bundle immediately.
func (h RegexpMatching) immediately(b Bundle, exactMatch bool) error {
	for _, p := range b.Packets {
		errs := []any{}
		if err := h.invoke(p, exactMatch); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("failed to invoke osc bundle "+strings.Repeat(": %w", len(errs)), errs...)
		}
	}
	return nil
}

// invoke invokes an OSC packet, which could be a message or a bundle of messages.
func (h RegexpMatching) invoke(p Packet, exactMatch bool) error {
	switch x := p.(type) {
	case Message:
		return h.Invoke(x, exactMatch)
	case Bundle:
		return h.immediately(x, exactMatch)
	default:
		return errors.Errorf("unsupported type for dispatcher: %T", p)
	}
}

// Invoke invokes an OSC message.
func (h RegexpMatching) Invoke(msg Message, exactMatch bool) error {
	for addressPattern, handler := range h {

		re, err := regexp.Compile(addressPattern)
		if err != nil {
			return err
		}

		if re.MatchString(msg.Address) {
			return handler.Handle(msg)
		}
	}
	return nil
}
