package test

import (
	"net"
	"strconv"
)

func FindPorts(start, amount int) []int {
	var found int
	ports := make([]int, amount)

	for port := start; port < start+amount && found < amount; port++ {
		if isPortAvailable(port) {
			ports[found] = port
			found++
		}
	}

	return ports[:found]
}

func isPortAvailable(port int) bool {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}

	if err := l.Close(); err != nil {
		return false
	}

	return true
}
