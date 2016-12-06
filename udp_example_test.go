package osc

import (
	"fmt"
	"log"
	"net"
	"time"
)

func ExampleUDPConn_pingpong() {
	errChan := make(chan error)

	// Setup the server.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	server, err := ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	// Run the server in a new goroutine.
	go func() {
		errChan <- server.Serve(Dispatcher{
			"/ping": func(msg Message) error {
				fmt.Println("Server received ping.")
				return server.SendTo(msg.Sender, Message{Address: "/pong"})
			},
			"/close": func(msg Message) error {
				if err := server.SendTo(msg.Sender, Message{Address: "/close"}); err != nil {
					_ = server.Close()
					return err
				}
				fmt.Println("Server closing.")
				return server.Close()
			},
		})
	}()

	// Setup the client.
	raddr, err := net.ResolveUDPAddr("udp", server.LocalAddr().String())
	if err != nil {
		log.Fatal(err)
	}
	client, err := DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}

	// Clients are also servers!
	var (
		pongChan        = make(chan struct{})
		clientCloseChan = make(chan struct{})
	)

	go func() {
		errChan <- client.Serve(Dispatcher{
			"/pong": func(msg Message) error {
				fmt.Println("Client received pong.")
				close(pongChan)
				return nil
			},
			"/close": func(msg Message) error {
				fmt.Println("Client closing.")
				close(clientCloseChan)
				return client.Close()
			},
		})
	}()

	// Send the ping message, wait for the pong, then close both connections.
	if err := client.Send(Message{Address: "/ping"}); err != nil {
		log.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		log.Fatal("timeout")
	case err := <-errChan:
		if err != nil {
			log.Fatal(err)
		}
	case <-pongChan:
	}
	if err := client.Send(Message{Address: "/close"}); err != nil {
		log.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		log.Fatal("timeout")
	case err := <-errChan:
		if err != nil {
			log.Fatal(err)
		}
	}
	select {
	case <-time.After(2 * time.Second):
		log.Fatal("timeout")
	case <-clientCloseChan:
	}
	// Output:
	// Server received ping.
	// Client received pong.
	// Server closing.
	// Client closing.
}
