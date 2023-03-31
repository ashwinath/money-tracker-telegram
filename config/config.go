package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey string `yaml:"apiKey"`
	Debug  bool   `yaml:"debug"`
}

func New(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
