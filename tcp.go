package osc

import "net"

// TCPConn is an OSC connection over TCP.
type TCPConn struct {
	net.TCPConn
	dispatcher Dispatcher
}
