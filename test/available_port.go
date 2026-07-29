package test

import (
	"astro/adapters/http"
	"astro/adapters/metrics"
	"errors"
	"net"
	"strconv"

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

var NewPortMetricsConfig = fx.Decorate(func(metricsConfig metrics.Config) (metrics.Config, error) {
	address, err := availableAddress(metricsConfig.Address)
	if err != nil {
		return metricsConfig, err
	}

	metricsConfig.Address = address

	return metricsConfig, nil
})

func availableAddress(address string) (string, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	addr, err := findAvailableAddr()
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(host, strconv.Itoa(addr.Port)), nil
}

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
