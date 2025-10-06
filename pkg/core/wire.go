//go:build wireinject

package core

import "github.com/google/wire"

// InitializeApp is a placeholder for wire-based dependency injection.
func InitializeApp(opts BuildOptions) (*App, error) {
	panic(wire.Build(ProvideConfig, ProvideLogger))
}
