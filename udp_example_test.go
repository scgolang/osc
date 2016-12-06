package osc

import (
	"errors"
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
	go serverDispatch(server, errChan)

	// Setup the client.
	raddr, err := net.ResolveUDPAddr("udp", server.LocalAddr().String())
	if err != nil {
		log.Fatal(err)
	}
	client, err := DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}

	var (
		pongChan        = make(chan struct{})
		clientCloseChan = make(chan struct{})
	)
	// Clients are also servers!
	go clientDispatch(client, errChan, pongChan, clientCloseChan)

	// Send the ping message, wait for the pong, then close both connections.
	if err := client.Send(Message{Address: "/ping"}); err != nil {
		log.Fatal(err)
	}
	if err := waitPong(pongChan, errChan); err != nil {
		log.Fatal(err)
	}

	if err := client.Send(Message{Address: "/close"}); err != nil {
		log.Fatal(err)
	}
	if err := waitErr(errChan); err != nil {
		log.Fatal(err)
	}
	if err := waitClose(clientCloseChan, errChan); err != nil {
		log.Fatal(err)
	}
	// Output:
	// Server received ping.
	// Client received pong.
	// Server closing.
	// Client closing.
}

func serverDispatch(server *UDPConn, errChan chan error) {
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
}

func clientDispatch(client *UDPConn, errChan chan error, pongChan chan struct{}, closeChan chan struct{}) {
	errChan <- client.Serve(Dispatcher{
		"/pong": func(msg Message) error {
			fmt.Println("Client received pong.")
			close(pongChan)
			return nil
		},
		"/close": func(msg Message) error {
			fmt.Println("Client closing.")
			close(closeChan)
			return client.Close()
		},
	})
}

func waitPong(pongChan chan struct{}, errChan chan error) error {
	select {
	case <-time.After(2 * time.Second):
		return errors.New("timeout")
	case err := <-errChan:
		if err != nil {
			return err
		}
	case <-pongChan:
	}
	return nil
}

func waitErr(errChan chan error) error {
	select {
	case <-time.After(2 * time.Second):
		return errors.New("timeout")
	case err := <-errChan:
		return err
	}
}

func waitClose(closeChan chan struct{}, errChan chan error) error {
	select {
	case <-time.After(2 * time.Second):
		return errors.New("timeout")
	case <-closeChan:
	}
	return nil
}
