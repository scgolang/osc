package osc

import (
	"net"
)

// Client is an OSC client.
type Client struct {
	Address string
	laddr   *net.UDPAddr
}

// NewClient creates a new OSC client. The Client is used to send OSC
// messages and OSC bundles over an UDP network connection. The argument ip
// specifies the IP address and port defines the target port where the messages
// and bundles will be send to.
func NewClient(addr string) *Client {
	return &Client{Address: addr, laddr: nil}
}

// SetLocalAddr sets the local address.
func (self *Client) SetLocalAddr(addr string) error {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	self.laddr = laddr
	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (self *Client) Send(packet Packet) error {
	addr, err := net.ResolveUDPAddr("udp", self.Address)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", self.laddr, addr)
	if err != nil {
		return err
	}

	data, err := packet.ToByteArray()
	if err != nil {
		conn.Close()
		return err
	}

	_, err = conn.Write(data)
	if err != nil {
		conn.Close()
		return err
	}

	conn.Close()

	return nil
}
