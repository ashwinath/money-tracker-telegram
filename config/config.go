package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey   string   `yaml:"apiKey"`
	Debug    bool     `yaml:"debug"`
	DBConfig DBConfig `yaml:"dbConfig"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbName"`
	Port     uint   `yaml:"port"`
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
