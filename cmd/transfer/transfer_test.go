package main

import (
	"bytes"
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTransferCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transfer command")
}

var _ = Describe("transfer command", func() {
	It("requires an explicit direction", func() {
		var stdout, stderr bytes.Buffer

		status := run(context.Background(), nil, func(string) string { return "" }, &stdout, &stderr)

		Expect(status).To(Equal(2))
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal("invalid transfer options; use --help\n"))
	})

	It("does not print DSNs from connection errors", func() {
		var stdout, stderr bytes.Buffer
		secret := "host=127.0.0.1 port=1 user=synthetic password=PRIVATE_TRANSFER_PASSWORD dbname=test connect_timeout=1 sslmode=disable"

		status := run(context.Background(), []string{"--direction", "verify", "--sqlite-path", "unused.db", "--timeout", "1s"}, func(string) string { return secret }, &stdout, &stderr)

		Expect(status).To(Equal(1))
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Or(Equal("database connection unavailable\n"), Equal("context deadline exceeded\n")))
		Expect(stderr.String()).NotTo(ContainSubstring("PRIVATE_TRANSFER_PASSWORD"))
	})
})
