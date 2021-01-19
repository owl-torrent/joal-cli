package types

import (
	"context"
)

type IJoalPlugin interface {
	Name() string
	// decide if the plugin should be enabled (most of the time i will be based on program arguments)
	Start() error
	Shutdown(ctx context.Context)
}
