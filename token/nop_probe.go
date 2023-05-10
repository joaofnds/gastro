package token

import (
	"go.uber.org/fx"
)

var NopProbeProvider = fx.Decorate(func() Probe { return NopProbe{} })

type NopProbe struct{}

func (i NopProbe) TokenCreated()              {}
func (i NopProbe) FailedToCreateToken(error)  {}
func (i NopProbe) TokenDecrypted()            {}
func (i NopProbe) FailedToDecryptToken(error) {}
