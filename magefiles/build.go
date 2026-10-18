//go:build mage

package main

import "astro/magefiles/dag"

func Build() error {
	return dag.BuildProd()
}
