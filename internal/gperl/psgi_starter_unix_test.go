//go:build unix

package gperl

import (
	"net"
	"strconv"
	"syscall"
	"testing"
)

// The sockets a supervisor hands over (SERVER_STARTER_PORT) take the place
// of the addresses to listen on; a malformed entry is an error.
func TestPSGIStarterListeners(t *testing.T) {
	held, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := held.Close(); err != nil {
			t.Error(err)
		}
	}()
	f, err := held.(*net.TCPListener).File()
	if err != nil {
		t.Fatal(err)
	}
	// The descriptor handed over becomes the server's, which closes it:
	// it is a copy no *os.File of the test owns.
	fd, err := syscall.Dup(int(f.Fd()))
	if cerr := f.Close(); err != nil || cerr != nil {
		t.Fatal(err, cerr)
	}
	c := &psgiConfig{starterPorts: held.Addr().String() + "=" + strconv.Itoa(fd)}
	c.listen.values = []string{"127.0.0.1:0"}
	lns, err := c.listeners()
	if err != nil || len(lns) != 1 || lns[0].Addr().String() != held.Addr().String() {
		t.Fatalf("SERVER_STARTER_PORT listeners = %v, %v", lns, err)
	}
	if err := lns[0].Close(); err != nil {
		t.Error(err)
	}
	for _, bad := range []string{"127.0.0.1:1", "127.0.0.1:1=fd"} {
		if _, err := (&psgiConfig{starterPorts: bad}).listeners(); err == nil {
			t.Errorf("SERVER_STARTER_PORT=%s: no error", bad)
		}
	}
}
