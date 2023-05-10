package token

import (
	"go.uber.org/fx"
)

var NopProbeProvider = fx.Decorate(func() Probe { return NopProbe{} })

type NopProbe struct{}

func (p NopProbe) TokenCreated()              {}
func (p NopProbe) FailedToCreateToken(error)  {}
func (p NopProbe) TokenDecrypted()            {}
func (p NopProbe) FailedToDecryptToken(error) {}
