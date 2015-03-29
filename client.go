package osc

import (
	"fmt"
	"net"
)

type Client struct {
	ipaddress string
	port      int
	laddr     *net.UDPAddr
}

// NewClient creates a new OSC client. The Client is used to send OSC
// messages and OSC bundles over an UDP network connection. The argument ip
// specifies the IP address and port defines the target port where the messages
// and bundles will be send to.
func NewClient(ip string, port int) (client *Client) {
	return &Client{ipaddress: ip, port: port, laddr: nil}
}

// Ip returns the IP address.
func (client *Client) Ip() string {
	return client.ipaddress
}

// SetIp sets a new IP address.
func (client *Client) SetIp(ip string) {
	client.ipaddress = ip
}

// Port returns the port.
func (client *Client) Port() int {
	return client.port
}

// SetPort sets a new port.
func (client *Client) SetPort(port int) {
	client.port = port
}

// SetLocalAddr sets the local address.
func (client *Client) SetLocalAddr(ip string, port int) error {
	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	client.laddr = laddr
	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (client *Client) Send(packet Packet) (err error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", client.ipaddress, client.port))
	conn, err := net.DialUDP("udp", client.laddr, addr)
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
