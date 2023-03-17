package test

import (
	"github.com/magefile/mage/sh"
)

// Runs all tests
func TestAll() error {
	return sh.RunV("go", "test", "./...")
}
