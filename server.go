package osc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

// Common errors.
var (
	ErrNoDispatcher   = errors.New("no dispatcher defined")
	ErrPrematureClose = errors.New("server cannot be closed before calling Listen")
	ErrInvalidTypeTag = errors.New("invalid type tag")
)

// Server is an OSC server.
type Server struct {
	Listening  chan struct{} // Listening is a channel used to indicate when the server is running.
	dispatcher oscDispatcher // Dispatcher that dispatches OSC packets/messages.
	conn       *net.UDPConn  // conn is a UDP connection object.
}

// NewServer returns a new OSC Server.
func NewServer(addr string, handlers map[string]HandlerFunc) (*Server, error) {
	for addr, _ := range handlers {
		if err := validateAddress(addr); err != nil {
			return nil, err
		}
	}

	netAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", netAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		Listening:  make(chan struct{}),
		dispatcher: oscDispatcher(handlers),
		conn:       conn,
	}, nil
}

// LocalAddr returns the local network address.
func (self *Server) LocalAddr() net.Addr {
	return self.conn.LocalAddr()
}

// Close stops the OSC server and closes the connection.
func (self *Server) Close() error {
	if self.conn == nil {
		return nil
	}
	return self.conn.Close()
}

// Listen retrieves incoming OSC packets and dispatches the retrieved OSC packets.
func (self *Server) Listen() error {
	if self.dispatcher == nil {
		return ErrNoDispatcher
	}

	self.Listening <- struct{}{}

	msg, err := self.readFromConnection()
	if err != nil {
		return err
	}
	self.dispatcher.dispatch(msg)

	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (self *Server) SendTo(addr net.Addr, packet Packet) (err error) {
	data, err := packet.ToByteArray()
	if err != nil {
		self.conn.Close()
		return err
	}

	written, err := self.conn.WriteTo(data, addr)
	if err != nil {
		fmt.Println("could not write packet")
		self.conn.Close()
		return err
	}
	if written != len(data) {
		errmsg := "only wrote %d bytes of osc packet with length %d"
		return fmt.Errorf(errmsg, written, len(data))
	}

	return nil
}

// readFromConnection retrieves OSC packets.
func (self *Server) readFromConnection() (packet Packet, err error) {
	data := make([]byte, 65535)
	var n, start int
	n, _, err = self.conn.ReadFromUDP(data)
	packet, err = self.readPacket(bufio.NewReader(bytes.NewBuffer(data)), &start, n)

	return packet, nil
}

// receivePacket receives an OSC packet from the given reader.
func (self *Server) readPacket(reader *bufio.Reader, start *int, end int) (packet Packet, err error) {
	var buf []byte
	buf, err = reader.Peek(1)
	if err != nil {
		return nil, err
	}

	// An OSC Message starts with a '/'
	if buf[0] == '/' {
		packet, err = self.readMessage(reader, start)
		if err != nil {
			return nil, err
		}
	} else if buf[0] == '#' { // An OSC bundle starts with a '#'
		packet, err = self.readBundle(reader, start, end)
		if err != nil {
			return nil, err
		}
	}

	return packet, nil
}

// readBundle reads an Bundle from reader.
func (self *Server) readBundle(reader *bufio.Reader, start *int, end int) (bundle *Bundle, err error) {
	// Read the '#bundle' OSC string
	var startTag string
	var n int
	startTag, n, err = readPaddedString(reader)
	if err != nil {
		return nil, err
	}
	*start += n

	if startTag != BundleTag {
		return nil, fmt.Errorf("Invalid bundle start tag: %s", startTag)
	}

	// Read the timetag
	var timeTag uint64
	if err := binary.Read(reader, binary.BigEndian, &timeTag); err != nil {
		return nil, err
	}
	*start += 8

	// Create a new bundle
	bundle = NewBundle(timetagToTime(timeTag))

	// Read until the end of the buffer
	for *start < end {
		// Read the size of the bundle element
		var length int32
		err = binary.Read(reader, binary.BigEndian, &length)
		*start += 4
		if err != nil {
			return nil, err
		}

		var packet Packet
		packet, err = self.readPacket(reader, start, end)
		if err != nil {
			return nil, err
		}
		bundle.Append(packet)
	}

	return bundle, nil
}

// readMessage reads one OSC Message from reader.
func (self *Server) readMessage(reader *bufio.Reader, start *int) (msg *Message, err error) {
	// First, read the OSC address
	var n int
	address, n, err := readPaddedString(reader)
	if err != nil {
		return nil, err
	}
	*start += n

	// Create a new message
	msg = &Message{address: address}

	// Read all arguments
	if err = self.readArguments(msg, reader, start); err != nil {
		return nil, err
	}

	return msg, nil
}

// readArguments reads all arguments from the reader and adds it to the OSC message.
func (self *Server) readArguments(msg *Message, reader *bufio.Reader, start *int) error {
	// Read the type tag string
	var n int
	typetags, n, err := readPaddedString(reader)
	if err != nil {
		return err
	}
	*start += n

	// If the typetag doesn't start with ',', it's not valid
	if typetags[0] != ',' {
		return ErrInvalidTypeTag
	}

	// Remove ',' from the type tag
	typetags = typetags[1:]

	for _, c := range typetags {
		switch c {
		default:
			return fmt.Errorf("Unsupported type tag: %c", c)

		// int32
		case 'i':
			var i int32
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			*start += 4
			msg.Append(i)

		// int64
		case 'h':
			var i int64
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			*start += 8
			msg.Append(i)

		// float32
		case 'f':
			var f float32
			if err = binary.Read(reader, binary.BigEndian, &f); err != nil {
				return err
			}
			*start += 4
			msg.Append(f)

		// float64/double
		case 'd':
			var d float64
			if err = binary.Read(reader, binary.BigEndian, &d); err != nil {
				return err
			}
			*start += 8
			msg.Append(d)

		// string
		case 's':
			// TODO: fix reading string value
			var s string
			if s, _, err = readPaddedString(reader); err != nil {
				return err
			}
			*start += len(s) + padBytesNeeded(len(s))
			msg.Append(s)

		// blob
		case 'b':
			var buf []byte
			var n int
			if buf, n, err = readBlob(reader); err != nil {
				return err
			}
			*start += n
			msg.Append(buf)

		// OSC Time Tag
		case 't':
			var tt uint64
			if err = binary.Read(reader, binary.BigEndian, &tt); err != nil {
				return nil
			}
			*start += 8
			msg.Append(Timetag(tt))

		// True
		case 'T':
			msg.Append(true)

		// False
		case 'F':
			msg.Append(false)
		}
	}

	return nil
}

// readBlob reads an OSC Blob from the blob byte array. Padding bytes are removed
// from the reader and not returned.
func readBlob(reader *bufio.Reader) (blob []byte, n int, err error) {
	// First, get the length
	var blobLen int
	if err = binary.Read(reader, binary.BigEndian, &blobLen); err != nil {
		return nil, 0, err
	}
	n = 4 + blobLen

	// Read the data
	blob = make([]byte, blobLen)
	if _, err = reader.Read(blob); err != nil {
		return nil, 0, err
	}

	// Remove the padding bytes
	numPadBytes := padBytesNeeded(blobLen)
	if numPadBytes > 0 {
		n += numPadBytes
		dummy := make([]byte, numPadBytes)
		if _, err = reader.Read(dummy); err != nil {
			return nil, 0, err
		}
	}

	return blob, n, nil
}

var invalidAddressRunes = []rune{'*', '?', ',', '[', ']', '{', '}', '#', ' '}

func validateAddress(addr string) error {
	for _, chr := range invalidAddressRunes {
		if strings.ContainsRune(addr, chr) {
			return ErrInvalidAddress
		}
	}
	return nil
}
