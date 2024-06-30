package log

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultLevel string            `yaml:"defaultLevel"`
	Loggers      map[string]string `yaml:"loggers"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
