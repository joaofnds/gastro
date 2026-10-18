package http

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/gofiber/fiber/v3"
)

type server struct {
	app      *fiber.App
	address  string
	shutdown func()
	stopped  chan error
}

func (server *server) Start(ctx context.Context) error {
	network := net.ListenConfig{}
	bound, err := network.Listen(ctx, "tcp4", server.address)
	if err != nil {
		return err
	}

	ready := make(chan struct{})
	listener := &startupListener{Listener: bound, ready: sync.OnceFunc(func() { close(ready) })}

	go server.serve(listener)

	select {
	case <-ready:
		return nil
	case err := <-server.stopped:
		return errors.Join(err, listener.Close())
	case <-ctx.Done():
		closeErr := listener.Close()
		return errors.Join(ctx.Err(), closeErr, <-server.stopped)
	}
}

func (server *server) Stop(ctx context.Context) error {
	err := server.app.ShutdownWithContext(ctx)

	select {
	case serveErr := <-server.stopped:
		return errors.Join(err, serveErr)
	case <-ctx.Done():
		return errors.Join(err, ctx.Err())
	}
}

func (server *server) serve(listener net.Listener) {
	err := server.app.Listener(listener, fiber.ListenConfig{DisableStartupMessage: true})

	server.stopped <- err

	if err != nil {
		server.shutdown()
	}
}

type startupListener struct {
	net.Listener
	ready func()
}

func (listener *startupListener) Accept() (net.Conn, error) {
	listener.ready()
	return listener.Listener.Accept()
}
