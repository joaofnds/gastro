package test

import (
	"astro/token"

	"go.uber.org/fx"
)

var NopTokenInstrumentation = fx.Decorate(NewFakeTokenInstrumentation)

type FakeTokenInstrumentation struct{}

func NewFakeTokenInstrumentation() token.TokenInstrumentation {
	return &FakeTokenInstrumentation{}
}
func (i FakeTokenInstrumentation) TokenCreated()                  {}
func (i FakeTokenInstrumentation) FailedToCreateToken(err error)  {}
func (i FakeTokenInstrumentation) TokenDecrypted()                {}
func (i FakeTokenInstrumentation) FailedToDecryptToken(err error) {}
