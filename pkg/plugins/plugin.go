package plugins

import (
	"context"
)

type IJoalPlugin interface {
	Name() string
	// name of the config folder in joal root
	SubFolder() string
	Enabled() bool
	Initialize(joalRootFolder string) error
	AfterCoreLoaded(coreBridge ICoreBridge)
	Shutdown(ctx context.Context)
}
