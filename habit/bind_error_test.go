package habit

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// v3 wraps every decoder failure in *fiber.BindError, whose Error() prefixes the response
// body with "bind from body: ". v2 sent the decoder's own message. These pin both branches:
// the wrapped one, and the *fiber.Error that Bind() returns for an unsupported Content-Type
// without wrapping — no probe reaches that path, since it is unreachable from the TUI.

func TestBindErrorMessage(t *testing.T) {
	t.Run("returns the wrapped message", func(t *testing.T) {
		err := &fiber.BindError{Err: errors.New("Syntax error at index 1"), Source: "body"}

		if got, want := messageFromBindError(err), "Syntax error at index 1"; got != want {
			t.Errorf("messageFromBindError() = %q, want %q", got, want)
		}
	})

	t.Run("returns the wrapped message when the field is populated", func(t *testing.T) {
		err := &fiber.BindError{Err: errors.New("mismatched type"), Source: "body", Field: "name"}

		if got, want := messageFromBindError(err), "mismatched type"; got != want {
			t.Errorf("messageFromBindError() = %q, want %q", got, want)
		}
	})

	t.Run("falls back to the error's own message when it is not a BindError", func(t *testing.T) {
		if got, want := messageFromBindError(fiber.ErrUnprocessableEntity), "Unprocessable Entity"; got != want {
			t.Errorf("messageFromBindError() = %q, want %q", got, want)
		}
	})
}
