package configloader

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type IConfigLoader interface {
	ParseOverDefault() (interface{}, error) // FIXME: this is a place for generic
	SaveConfigToFile(newConf interface{}) error
}

func NewConfigLoader(configFilePath string, getNewDefault GetNewDefault) IConfigLoader {
	return &configLoader{
		configFilePath: configFilePath,
		getNewDefault:  getNewDefault,
	}
}

type GetNewDefault func() interface{}

type configLoader struct {
	configFilePath string
	getNewDefault  GetNewDefault
}

func (c *configLoader) ParseOverDefault() (interface{}, error) {
	f, err := os.Open(c.configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open config file '%s'", c.configFilePath)
	}

	loadedConfig := c.getNewDefault()
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	err = decoder.Decode(loadedConfig)
	if err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "failed to parse config file '%s'", c.configFilePath)
	}

	return loadedConfig, nil
}

func (c *configLoader) SaveConfigToFile(newConf interface{}) error {
	f, err := os.OpenFile(c.configFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create config file '%s'", c.configFilePath)
	}
	defer func() { _ = f.Close() }()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	err = encoder.Encode(newConf)
	return errors.Wrapf(err, "failed to write to config file '%s'", c.configFilePath)
}
