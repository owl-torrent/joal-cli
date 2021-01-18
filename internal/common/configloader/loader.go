package configloader

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func ParseIntoDefault(configFilePath string, defaultValue interface{}) error {
	f, err := os.Open(configFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open config file '%s'", configFilePath)
	}

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	err = decoder.Decode(defaultValue)
	if err != nil && err != io.EOF {
		return errors.Wrapf(err, "failed to parse config file '%s'", configFilePath)
	}
	return nil
}

func SaveToFile(configFilePath string, newConf interface{}) error {
	f, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create config file '%s'", configFilePath)
	}
	defer func() { _ = f.Close() }()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	err = encoder.Encode(newConf)
	return errors.Wrapf(err, "failed to write to config file '%s'", configFilePath)
}
