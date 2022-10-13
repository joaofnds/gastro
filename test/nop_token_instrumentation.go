package test

import (
	"astro/token"

	"go.uber.org/fx"
)

var NopTokenInstrumentation = fx.Decorate(NewFakeTokenInstrumentation)

type nopTokenInstrumentation struct{}

func NewFakeTokenInstrumentation() token.Instrumentation {
	return &nopTokenInstrumentation{}
}
func (i nopTokenInstrumentation) TokenCreated()                  {}
func (i nopTokenInstrumentation) FailedToCreateToken(err error)  {}
func (i nopTokenInstrumentation) TokenDecrypted()                {}
func (i nopTokenInstrumentation) FailedToDecryptToken(err error) {}
