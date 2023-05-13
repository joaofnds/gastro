package token

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTokenService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "token suite")
}
