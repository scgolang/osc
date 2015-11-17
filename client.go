package osc

import "net"

// Client is an OSC client.
type Client struct {
	addr net.Addr
}

// NewClient creates a new OSC client. The Client is used to send OSC
// messages and OSC bundles over an UDP network connection. The argument ip
// specifies the IP address and port defines the target port where the messages
// and bundles will be send to.
func NewClient(addr net.Addr) (*Client, error) {
	return &Client{addr: addr}, nil
}

// Send sends an OSC Bundle or an OSC Message.
func (client *Client) Send(msg []byte) error {
	network, addr := client.addr.Network(), client.addr.String()

	udpAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	if _, err = conn.Write(msg); err != nil {
		return conn.Close()
	}
	return conn.Close()
}
