package token_test

import (
	"astro/habit"
	"astro/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UUIDGenerator", func() {
	var generator *token.UUIDGenerator

	BeforeEach(func() { generator = token.NewUUIDGenerator() })

	Describe("NewID", func() {
		It("returns a new UUID", func() {
			id, err := generator.NewID()

			Expect(err).NotTo(HaveOccurred())
			Expect(id).To(HaveLen(36))
			Expect(habit.IsUUID(string(id))).To(BeTrue())
		})
	})
})
