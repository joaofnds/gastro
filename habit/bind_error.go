package habit

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// messageFromBindError returns the message a decoder failure had before v3 started wrapping it.
// Bind() returns a bare *fiber.Error for an unsupported Content-Type, so the unwrap has to
// tolerate an error that is not a *fiber.BindError.
func messageFromBindError(err error) string {
	var bindErr *fiber.BindError
	if errors.As(err, &bindErr) {
		return bindErr.Err.Error()
	}

	return err.Error()
}
