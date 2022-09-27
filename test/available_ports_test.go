package test_test

import (
	"astro/test"
	"reflect"
	"testing"
)

func TestFindPorts(t *testing.T) {
	want := []int{10_000, 10_001, 10_002, 10_003, 10_004}
	have := test.FindPorts(10_000, 5)

	if !reflect.DeepEqual(have, want) {
		t.Errorf("have %v. want %v", have, want)
	}
}
