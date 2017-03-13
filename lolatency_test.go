package osc_test

import (
	"net"
	"testing"

	"github.com/scgolang/osc"
)

// BenchmarkMessageSend measures the latency of sending a single OSC message on localhost.
// This clocks in at around 80000 ns avg on my MBP with dual-core i7 @ 3.1 GHz [briansorahan]
// For a sample rate of 48kHz this would come out to about 4 samples.
// Thus we should not expect that it is currently possible to achieve sample accurate synchronization with OSC over localhost.
// But 80us latency is not bad!
//
// Update 3/11/2017 [briansorahan]
// This benchmarks around 100us on my Dell Latitude E6510 with a single-core Core i7 @ 2.8GHz
func BenchmarkUDPSend(b *testing.B) {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	srv, err := osc.ListenUDP("udp", laddr)
	if err != nil {
		b.Fatal(err)
	}
	raddr, err := net.ResolveUDPAddr("udp", srv.LocalAddr().String())
	if err != nil {
		b.Fatal(err)
	}
	conn, err := osc.DialUDP("udp", nil, raddr)
	if err != nil {
		b.Fatal(err)
	}
	var (
		ch  = make(chan struct{})
		val = struct{}{}
	)
	go srv.Serve(8, osc.Dispatcher{
		"/ping": osc.Method(func(m osc.Message) error {
			ch <- val
			return nil
		}),
	})
	msg := osc.Message{Address: "/ping"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn.Send(msg)
		<-ch
	}
}

// Including a single argument does not seem to have much effect on latency [briansorahan].
func BenchmarkUDPSendOneArgument(b *testing.B) {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	srv, err := osc.ListenUDP("udp", laddr)
	if err != nil {
		b.Fatal(err)
	}
	raddr, err := net.ResolveUDPAddr("udp", srv.LocalAddr().String())
	if err != nil {
		b.Fatal(err)
	}
	conn, err := osc.DialUDP("udp", nil, raddr)
	if err != nil {
		b.Fatal(err)
	}
	var (
		ch  = make(chan struct{})
		val = struct{}{}
	)
	go srv.Serve(1, osc.Dispatcher{
		"/ping": osc.Method(func(m osc.Message) error {
			if _, err := m.Arguments[0].ReadInt32(); err != nil {
				return err
			}
			ch <- val
			return nil
		}),
	})
	msg := osc.Message{Address: "/ping", Arguments: osc.Arguments{osc.Int(0)}}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg.Arguments[0] = osc.Int(i)
		conn.Send(msg)
		<-ch
	}
}

func BenchmarkUnixSend(b *testing.B) {
	const network = "unixgram"

	laddr, err := net.ResolveUnixAddr(network, osc.TempSocket())
	if err != nil {
		b.Fatal(err)
	}
	srv, err := osc.ListenUnix("unixgram", laddr)
	if err != nil {
		b.Fatal(err)
	}
	raddr, err := net.ResolveUnixAddr(network, srv.LocalAddr().String())
	if err != nil {
		b.Fatal(err)
	}
	conn, err := osc.DialUnix(network, nil, raddr)
	if err != nil {
		b.Fatal(err)
	}
	var (
		ch  = make(chan struct{})
		val = struct{}{}
	)
	go srv.Serve(8, osc.Dispatcher{
		"/ping": osc.Method(func(m osc.Message) error {
			ch <- val
			return nil
		}),
	})
	msg := osc.Message{Address: "/ping"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn.Send(msg)
		<-ch
	}
}
