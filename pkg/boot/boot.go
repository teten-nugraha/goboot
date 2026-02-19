package boot

import (
	"go.uber.org/fx"
)

type Options struct {
	AppName   string
	Profile   string
	ConfigDir string
}

type Module interface {
	FX() fx.Option
	Enabled(opt Options) bool
}
