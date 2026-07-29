package test

import (
	"net"
	"strconv"
	"testing"
)

func TestAvailableAddressKeepsTheHost(t *testing.T) {
	address, err := availableAddress("localhost:9091")
	if err != nil {
		t.Fatalf("availableAddress returned %v", err)
	}

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("splitting %q returned %v", address, err)
	}

	if host != "localhost" {
		t.Fatalf("host was %q, want localhost", host)
	}
}

func TestAvailableAddressReplacesThePort(t *testing.T) {
	address, err := availableAddress("localhost:9091")
	if err != nil {
		t.Fatalf("availableAddress returned %v", err)
	}

	_, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("splitting %q returned %v", address, err)
	}

	if port == "9091" {
		t.Fatalf("port was still the configured 9091")
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("address %q was not bindable: %v", address, err)
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("closing listener returned %v", err)
	}
}

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
