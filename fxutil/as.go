package fxutil

import "go.uber.org/fx"

func As[I any](constructor any) fx.Option {
	return fx.Provide(fx.Annotate(constructor, fx.As(new(I))))
}
