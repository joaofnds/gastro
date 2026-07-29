package habit

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/binder"
)

// Bind() returns a bare *fiber.Error for an unsupported Content-Type, so the unwrap has to
// tolerate an error that is not a *fiber.BindError.
func messageFromBindError(err error) string {
	if errors.Is(err, binder.ErrUnimplementedCBOR) || errors.Is(err, binder.ErrMsgPackNotConfigured) {
		return fiber.ErrUnprocessableEntity.Message
	}

	var bindErr *fiber.BindError
	if errors.As(err, &bindErr) {
		return bindErr.Err.Error()
	}

	return err.Error()
}
