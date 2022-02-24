package web

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/common/configloader"
)

type webConfigLoader struct {
	configFilePath string
}

func newWebConfigLoader(configRootDir string) *webConfigLoader {
	return &webConfigLoader{
		configFilePath: webConfigFilePathFromRoot(configRootDir),
	}
}

func (l *webConfigLoader) ReadConfig() (*webConfig, error) {
	conf := webConfig{}.Default()
	err := configloader.ParseIntoDefault(l.configFilePath, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webConfig: %w", err)
	}

	return conf, nil
}

func (l *webConfigLoader) SaveConfigToFile(newConf *webConfig) error {
	err := configloader.SaveToFile(l.configFilePath, newConf)
	if err != nil {
		return fmt.Errorf("failed to save webConfig: %w", err)
	}
	return nil
}
