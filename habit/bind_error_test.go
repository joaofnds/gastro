package habit

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/binder"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("messageFromBindError", func() {
	It("returns the message the decoder produced", func() {
		err := &fiber.BindError{Err: errors.New("Syntax error at index 1"), Source: "body"}

		Expect(messageFromBindError(err)).To(Equal("Syntax error at index 1"))
	})

	It("falls back to the error's own message when it is not a BindError", func() {
		Expect(messageFromBindError(fiber.ErrUnprocessableEntity)).To(Equal("Unprocessable Entity"))
	})

	When("the decoder is not configured", func() {
		It("hides the setup hint for a CBOR body", func() {
			err := &fiber.BindError{Err: binder.ErrUnimplementedCBOR, Source: "body"}

			Expect(messageFromBindError(err)).To(Equal("Unprocessable Entity"))
		})

		It("hides the setup hint for a msgpack body", func() {
			err := &fiber.BindError{Err: binder.ErrMsgPackNotConfigured, Source: "body"}

			Expect(messageFromBindError(err)).To(Equal("Unprocessable Entity"))
		})
	})
})
