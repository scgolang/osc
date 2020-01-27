package osc

import (
	"fmt"
	"log"
	"net"
	"time"
)

func Example_customdispatcher() {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	server, err := ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	var (
		doneChan   = make(chan struct{})
		dispatcher = customDispatcher{done: doneChan}
		errChan    = make(chan error)
	)
	go func() {
		errChan <- server.Serve(1, dispatcher)
	}()

	// Send a message from the client.
	raddr, err := net.ResolveUDPAddr("udp", server.LocalAddr().String())
	if err != nil {
		log.Fatal(err)
	}
	client, err := DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Send(Message{Address: "/foo"}); err != nil {
		log.Fatal(err)
	}
	select {
	case <-doneChan:
	case err := <-errChan:
		log.Fatal(err)
	case <-time.After(5 * time.Second):
		panic("timeout waiting for custom dispatcher example to finish")
	}
	// Output:
	// method /foo received by custom dispatcher
}

type customDispatcher struct {
	done chan struct{}
}

func (d customDispatcher) Dispatch(bundle Bundle, exactMatch bool) error {
	fmt.Println("bundle received by custom dispatcher")
	close(d.done)
	return nil
}

func (d customDispatcher) Invoke(msg Message, exactMatch bool) error {
	fmt.Printf("method %s received by custom dispatcher\n", msg.Address)
	close(d.done)
	return nil
}
