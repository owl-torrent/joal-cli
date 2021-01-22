package plugins

import (
	"context"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/plugins/types"
	"github.com/anthonyraymond/joal-cli/internal/plugins/web"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
	"sync"
)

type IPluginManager interface {
	BootstrapPlugins(httpClient *http.Client)
	StartPlugins()
	ShutdownPlugins(context.Context)
}

type pluginManager struct {
	pluginsRootDir string
	bridge         types.ICoreBridge
	enabledPlugins []types.IJoalPlugin
	lock           *sync.Mutex
}

func NewPluginManager(appRootDir string, coreBridge types.ICoreBridge) IPluginManager {

	pm := &pluginManager{
		pluginsRootDir: filepath.Join(appRootDir, "plugins"),
		bridge:         coreBridge,
		enabledPlugins: []types.IJoalPlugin{},
		lock:           &sync.Mutex{},
	}

	return pm
}

func (pm *pluginManager) BootstrapPlugins(httpClient *http.Client) {
	log := logs.GetLogger()

	pm.lock.Lock()
	defer pm.lock.Unlock()

	if web.ShouldEnablePlugin() {
		p, err := web.BootStrap(pm.pluginsRootDir, pm.bridge, httpClient)
		if err != nil {
			log.Warn("Web plugin has failed to bootstrap, it will stay disabled")
		} else {
			pm.enabledPlugins = append(pm.enabledPlugins, p)
			log.Debug("plugin enabled", zap.String("plugin", p.Name()))
		}
	}
}

func (pm *pluginManager) StartPlugins() {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	log := logs.GetLogger()
	for _, plugin := range pm.enabledPlugins {
		err := plugin.Start()
		if err != nil {
			log.Error("plugin has failed to start", zap.String("plugin", plugin.Name()), zap.Error(err))
		}
	}
}

func (pm *pluginManager) ShutdownPlugins(ctx context.Context) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	wg := &sync.WaitGroup{}
	for _, plugin := range pm.enabledPlugins {
		wg.Add(1)
		go func(plugin types.IJoalPlugin) {
			defer wg.Done()
			plugin.Shutdown(ctx)
		}(plugin)
	}

	wg.Wait()
}
