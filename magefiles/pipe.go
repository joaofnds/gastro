//go:build mage

package main

import (
	"astro/magefiles/test"

	"github.com/magefile/mage/sh"
)

func Pipe() error {
	if err := GitLeaks(); err != nil {
		return err
	}

	if err := Lint(); err != nil {
		return err
	}

	return test.TestAll()
}

func GitLeaks() error {
	return sh.RunV("gitleaks", "detect", "--no-banner")
}

func Lint() error {
	return sh.RunV("golangci-lint", "run", "./...")
}
