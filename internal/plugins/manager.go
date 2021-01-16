package plugins

/*
import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/core/config"
	"github.com/anthonyraymond/joal-cli/pkg/plugins/web"
)

var allAvailablePlugins = []IJoalPlugin{
	&web.Plugin{},
}

type IPluginManager interface {
	InitializePlugins(ICoreBridge, config.IConfigLoader)
	ShutdownPlugins(context.Context)
}

type pluginManager struct {
	enabledPlugins []IJoalPlugin
}

func newPluginManager() IPluginManager {
	pm := &pluginManager{
		enabledPlugins: []IJoalPlugin{},
	}

	for _, p := range allAvailablePlugins {
		if p.ShouldEnable() {
			pm.enabledPlugins = append(pm.enabledPlugins, p)
		}
	}
	return pm
}

func (p *pluginManager) InitializePlugins(bridge ICoreBridge, loader config.IConfigLoader) {
	for _, p := range p.enabledPlugins {
		p.I
	}
}

func (p *pluginManager) ShutdownPlugins(ctx context.Context) {
	panic("implement me")
}

*/
