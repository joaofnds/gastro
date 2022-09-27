package test

import (
	"net"
	"strconv"
)

func FindPorts(start, amount int) []int {
	var found int
	ports := make([]int, amount)

	for port := start; port < start+amount && found < amount; port++ {
		if l, err := net.Listen("tcp", ":"+strconv.Itoa(port)); err == nil {
			l.Close()
			ports[found] = port
			found++
		}
	}

	return ports[:found]
}
