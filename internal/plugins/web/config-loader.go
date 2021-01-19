package web

import (
	"github.com/anthonyraymond/joal-cli/internal/common/configloader"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "failed to parse webConfig")
	}

	return conf, nil
}

func (l *webConfigLoader) SaveConfigToFile(newConf *webConfig) error {
	err := configloader.SaveToFile(l.configFilePath, newConf)
	if err != nil {
		return errors.Wrap(err, "failed to save webConfig")
	}
	return nil
}
