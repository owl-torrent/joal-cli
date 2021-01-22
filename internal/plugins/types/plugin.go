package types

import (
	"context"
)

type IJoalPlugin interface {
	Name() string
	// decide if the plugin should be enabled (most of the time i will be based on program arguments)
	Start() error
	// Shutdown the plugin. It should be safe to call shutdown in any case even if the plugin wasn't started
	Shutdown(ctx context.Context)
}
