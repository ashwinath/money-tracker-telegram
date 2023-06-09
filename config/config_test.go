package config

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	p, err := os.Getwd()
	assert.Nil(t, err)

	p = path.Join(p, "../sample/sample_config.yaml")

	c, err := New(p)
	assert.Nil(t, err)

	assert.Equal(t, Config{
		APIKey:      "YourTelegramApiKey",
		AllowedUser: "yourTelegramHandle",
		Debug:       true,
		DBConfig: DBConfig{
			Host:     "localhost",
			User:     "postgres",
			Password: "password",
			DBName:   "postgres",
			Port:     5432,
		},
	}, *c)
}
