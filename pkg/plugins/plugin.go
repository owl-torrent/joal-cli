package plugins

import (
	"context"
)

type IJoalPlugin interface {
	Name() string
	// decide if the plugin should be enabled (most of the time i will be based on program arguments)
	ShouldEnable() bool
	Initialize(joalRootFolder string) error
	AfterCoreLoaded(coreBridge ICoreBridge)
	Shutdown(ctx context.Context)
}
