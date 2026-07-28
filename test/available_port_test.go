package test

import (
	"net"
	"strconv"
	"testing"
)

func TestFindAvailableAddr(t *testing.T) {
	addr, err := findAvailableAddr()
	if err != nil {
		t.Fatalf("findAvailableAddr returned %v", err)
	}

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(addr.Port))
	if err != nil {
		t.Fatalf("port %d was not bindable: %v", addr.Port, err)
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("closing listener returned %v", err)
	}
}
