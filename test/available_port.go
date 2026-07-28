package test

import (
	"astro/adapters/http"
	"errors"
	"net"

	"go.uber.org/fx"
)

var NewPortAppConfig = fx.Decorate(func(httpConfig http.Config) (http.Config, error) {
	addr, err := findAvailableAddr()
	if err != nil {
		return httpConfig, err
	}

	httpConfig.Port = addr.Port

	return httpConfig, nil
})

func findAvailableAddr() (*net.TCPAddr, error) {
	listener, listenErr := net.Listen("tcp", ":0")
	if listenErr != nil {
		return nil, listenErr
	}

	if closeErr := listener.Close(); closeErr != nil {
		return nil, closeErr
	}

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, errors.New("listener address is not a TCP address")
	}

	return addr, nil
}
