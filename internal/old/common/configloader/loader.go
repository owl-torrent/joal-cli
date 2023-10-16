package configloader

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func ParseIntoDefault(configFilePath string, defaultValue interface{}) error {
	f, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open config file '%s': %w", configFilePath, err)
	}

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	err = decoder.Decode(defaultValue)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to parse config file '%s': %w", configFilePath, err)
	}
	return nil
}

func SaveToFile(configFilePath string, newConf interface{}) error {
	f, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config file '%s': %w", configFilePath, err)
	}
	defer func() { _ = f.Close() }()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	err = encoder.Encode(newConf)
	return fmt.Errorf("failed to write to config file '%s': %w", configFilePath, err)
}
