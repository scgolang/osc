package osc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	ulid "github.com/imdario/go-ulid"
)

func tmpListener(t *testing.T, dispatcher Dispatcher) (*UnixConn, chan error) {
	newULID := ulid.New()
	addr, err := net.ResolveUnixAddr("unixgram", filepath.Join(os.TempDir(), newULID.String()))
	if err != nil {
		t.Fatal(err)
	}
	conn, err := ListenUnixContext(context.Background(), "unixgram", addr)
	if err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		if err := conn.Serve(1, dispatcher); err != nil {
			errChan <- err
		}
		close(errChan)
	}()
	return conn, errChan
}

func TestUnixSend(t *testing.T) {
	server, errChan := tmpListener(t, Dispatcher{})
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
}
