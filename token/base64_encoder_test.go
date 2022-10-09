package token_test

import (
	. "astro/test/matchers"
	"astro/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("base64 adapter", Ordered, func() {
	encoder := token.NewBase64Encoder()
	msg := "foo bar baz"
	msgB64 := "Zm9vIGJhciBiYXo="

	It("encodes to base64", func() {
		out := Must2(encoder.Encode([]byte(msg)))

		Expect(out).To(Equal([]byte(msgB64)))
	})

	It("decodes base64", func() {
		out := Must2(encoder.Decode([]byte(msgB64)))

		Expect(out).To(Equal([]byte(msg)))
	})

	It("can decoded own encoded payloads", func() {
		payload := []byte("Hello, World!")
		encoded := Must2(encoder.Encode(payload))
		decoded := Must2(encoder.Decode(encoded))

		Expect(decoded).To(Equal(payload))
	})
})
