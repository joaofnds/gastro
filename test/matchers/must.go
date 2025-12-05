package matchers

import (
	"github.com/onsi/gomega"
)

func Must(notError error) {
	gomega.Expect(notError).To(gomega.BeNil())
}

func Must2[T any](val T, notError error) T {
	gomega.Expect(notError).To(gomega.BeNil())
	return val
}

func Must3[T1 any, T2 any](val1 T1, val2 T2, notError error) (T1, T2) {
	gomega.Expect(notError).To(gomega.BeNil())
	return val1, val2
}
